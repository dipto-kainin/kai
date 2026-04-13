package example

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dipto-kainin/kai"
)

type showcaseFile struct {
	OriginalName string `json:"original_name"`
	Size         int    `json:"size"`
	MIME         string `json:"mime"`
	DownloadURL  string `json:"download_url"`
	StoredPath   string `json:"-"`
}

type showcasePost struct {
	ID         int           `json:"id"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Published  bool          `json:"published"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Attachment *showcaseFile `json:"attachment,omitempty"`
}

type showcaseStore struct {
	mu     sync.RWMutex
	nextID int
	posts  map[int]showcasePost
}

var crudShowcaseStore = newShowcaseStore()

func newShowcaseStore() *showcaseStore {
	now := time.Now().UTC()
	return &showcaseStore{
		nextID: 3,
		posts: map[int]showcasePost{
			1: {
				ID:        1,
				Title:     "Ship the first endpoint",
				Content:   "This seeded record helps you try GET and PUT immediately.",
				Published: true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			2: {
				ID:        2,
				Title:     "Draft the file upload flow",
				Content:   "Use POST /api/posts/2/file with multipart form data.",
				Published: false,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

func CRUD_SHOWCASE_ROUTES(app *kai.App) {
	api := app.Group("/api")

	api.GET("/posts", listShowcasePosts())
	api.GET("/posts/:id", getShowcasePost())
	api.POST("/posts", createShowcasePost())
	api.PUT("/posts/:id", updateShowcasePost())
	api.DELETE("/posts/:id", deleteShowcasePost())

	api.POST("/posts/:id/file", uploadShowcaseFile())
	api.GET("/posts/:id/file", downloadShowcaseFile())
	api.DELETE("/posts/:id/file", deleteShowcaseFile())
}

func listShowcasePosts() kai.HandlerFunc {
	return func(c *kai.Context) {
		limit := 50
		if rawLimit := c.Query("limit"); rawLimit != "" {
			parsed, err := strconv.Atoi(rawLimit)
			if err != nil || parsed <= 0 {
				c.JSON(http.StatusBadRequest, map[string]any{
					"error": "limit must be a positive integer",
				})
				return
			}
			limit = parsed
		}

		var publishedFilter *bool
		if rawPublished := c.Query("published"); rawPublished != "" {
			parsed, err := strconv.ParseBool(rawPublished)
			if err != nil {
				c.JSON(http.StatusBadRequest, map[string]any{
					"error": "published must be true or false",
				})
				return
			}
			publishedFilter = &parsed
		}

		posts := crudShowcaseStore.list(limit, publishedFilter)
		c.JSON(http.StatusOK, map[string]any{
			"items": posts,
			"count": len(posts),
			"query": map[string]any{
				"limit":     limit,
				"published": publishedFilter,
			},
		})
	}
}

func getShowcasePost() kai.HandlerFunc {
	return func(c *kai.Context) {
		id, ok := parsePostID(c)
		if !ok {
			return
		}

		post, found := crudShowcaseStore.get(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "post not found",
			})
			return
		}

		c.JSON(http.StatusOK, post)
	}
}

func createShowcasePost() kai.HandlerFunc {
	return func(c *kai.Context) {
		payload, ok := parsePostPayload(c)
		if !ok {
			return
		}

		post := crudShowcaseStore.create(payload)
		c.JSON(http.StatusCreated, map[string]any{
			"message": "post created",
			"item":    post,
		})
	}
}

func updateShowcasePost() kai.HandlerFunc {
	return func(c *kai.Context) {
		id, ok := parsePostID(c)
		if !ok {
			return
		}

		payload, ok := parsePostPayload(c)
		if !ok {
			return
		}

		post, found := crudShowcaseStore.update(id, payload)
		if !found {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "post not found",
			})
			return
		}

		c.JSON(http.StatusOK, map[string]any{
			"message": "post updated",
			"item":    post,
		})
	}
}

func deleteShowcasePost() kai.HandlerFunc {
	return func(c *kai.Context) {
		id, ok := parsePostID(c)
		if !ok {
			return
		}

		deleted, found := crudShowcaseStore.delete(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "post not found",
			})
			return
		}

		if deleted.Attachment != nil && deleted.Attachment.StoredPath != "" {
			_ = os.Remove(deleted.Attachment.StoredPath)
		}

		c.JSON(http.StatusOK, map[string]any{
			"message": "post deleted",
			"item":    deleted,
		})
	}
}

func uploadShowcaseFile() kai.HandlerFunc {
	return func(c *kai.Context) {
		id, ok := parsePostID(c)
		if !ok {
			return
		}

		post, found := crudShowcaseStore.get(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "post not found",
			})
			return
		}

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{
				"error": "multipart field 'file' is required",
			})
			return
		}
		_ = file.Close()

		fileBytes, err := c.GetFileBytes("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
			return
		}

		safeName := sanitizeFilename(header)
		dest := filepath.Join(os.TempDir(), "kai-showcase-uploads", fmt.Sprintf("post-%d-%s", id, safeName))
		if err := c.SaveToDest(dest, "file"); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "failed to save file",
			})
			return
		}

		if post.Attachment != nil && post.Attachment.StoredPath != "" && post.Attachment.StoredPath != dest {
			_ = os.Remove(post.Attachment.StoredPath)
		}

		updated := crudShowcaseStore.setAttachment(id, &showcaseFile{
			OriginalName: header.Filename,
			Size:         len(fileBytes),
			MIME:         header.Header.Get("Content-Type"),
			DownloadURL:  fmt.Sprintf("/api/posts/%d/file", id),
			StoredPath:   dest,
		})

		c.JSON(http.StatusOK, map[string]any{
			"message": "file uploaded",
			"item":    updated,
		})
	}
}

func downloadShowcaseFile() kai.HandlerFunc {
	return func(c *kai.Context) {
		id, ok := parsePostID(c)
		if !ok {
			return
		}

		post, found := crudShowcaseStore.get(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "post not found",
			})
			return
		}

		if post.Attachment == nil || post.Attachment.StoredPath == "" {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "file not found for this post",
			})
			return
		}

		c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%q", post.Attachment.OriginalName))
		c.ServeFile(post.Attachment.StoredPath)
	}
}

func deleteShowcaseFile() kai.HandlerFunc {
	return func(c *kai.Context) {
		id, ok := parsePostID(c)
		if !ok {
			return
		}

		post, found := crudShowcaseStore.get(id)
		if !found {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "post not found",
			})
			return
		}

		if post.Attachment == nil || post.Attachment.StoredPath == "" {
			c.JSON(http.StatusNotFound, map[string]any{
				"error": "file not found for this post",
			})
			return
		}

		if err := os.Remove(post.Attachment.StoredPath); err != nil && !os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "failed to delete file",
			})
			return
		}

		updated, _ := crudShowcaseStore.clearAttachment(id)
		c.JSON(http.StatusOK, map[string]any{
			"message": "file deleted",
			"item":    updated,
		})
	}
}

func parsePostID(c *kai.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, map[string]any{
			"error": "id must be a positive integer",
		})
		return 0, false
	}
	return id, true
}

func parsePostPayload(c *kai.Context) (showcasePost, bool) {
	body, err := c.GetJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"error": "request body must be valid JSON",
		})
		return showcasePost{}, false
	}

	title, ok := body["title"].(string)
	if !ok || strings.TrimSpace(title) == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"error": "title is required and must be a non-empty string",
		})
		return showcasePost{}, false
	}

	content, _ := body["content"].(string)
	published, _ := body["published"].(bool)

	return showcasePost{
		Title:     strings.TrimSpace(title),
		Content:   strings.TrimSpace(content),
		Published: published,
	}, true
}

func sanitizeFilename(header *multipart.FileHeader) string {
	name := filepath.Base(header.Filename)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, name)
	if name == "" {
		return "upload.bin"
	}
	return name
}

func (s *showcaseStore) list(limit int, published *bool) []showcasePost {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]showcasePost, 0, len(s.posts))
	for id := 1; id < s.nextID; id++ {
		post, ok := s.posts[id]
		if !ok {
			continue
		}
		if published != nil && post.Published != *published {
			continue
		}
		posts = append(posts, clonePost(post))
		if len(posts) == limit {
			break
		}
	}

	return posts
}

func (s *showcaseStore) get(id int) (showcasePost, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, ok := s.posts[id]
	if !ok {
		return showcasePost{}, false
	}
	return clonePost(post), true
}

func (s *showcaseStore) create(payload showcasePost) showcasePost {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	post := showcasePost{
		ID:        s.nextID,
		Title:     payload.Title,
		Content:   payload.Content,
		Published: payload.Published,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.posts[post.ID] = post
	s.nextID++

	return clonePost(post)
}

func (s *showcaseStore) update(id int, payload showcasePost) (showcasePost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.posts[id]
	if !ok {
		return showcasePost{}, false
	}

	post.Title = payload.Title
	post.Content = payload.Content
	post.Published = payload.Published
	post.UpdatedAt = time.Now().UTC()
	s.posts[id] = post

	return clonePost(post), true
}

func (s *showcaseStore) delete(id int) (showcasePost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.posts[id]
	if !ok {
		return showcasePost{}, false
	}

	delete(s.posts, id)
	return clonePost(post), true
}

func (s *showcaseStore) setAttachment(id int, attachment *showcaseFile) showcasePost {
	s.mu.Lock()
	defer s.mu.Unlock()

	post := s.posts[id]
	post.Attachment = attachment
	post.UpdatedAt = time.Now().UTC()
	s.posts[id] = post

	return clonePost(post)
}

func (s *showcaseStore) clearAttachment(id int) (showcasePost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.posts[id]
	if !ok {
		return showcasePost{}, false
	}

	post.Attachment = nil
	post.UpdatedAt = time.Now().UTC()
	s.posts[id] = post

	return clonePost(post), true
}

func clonePost(post showcasePost) showcasePost {
	if post.Attachment == nil {
		return post
	}

	attachment := *post.Attachment
	post.Attachment = &attachment
	return post
}
