package kai

import (
	"fmt"
	"net/http"
	"strconv"
)

type App struct {
	Router *Router
}

type Group struct {
	Prefix string
	router *Router
}

func NewApp() *App {
	return &App{
		Router: NewRouter(),
	}
}

func (a *App) Play(port int, message ...string) error {
	if len(message) > 0 && message[0] != "" {
		fmt.Println(message[0])
	} else {
		fmt.Println("Server is running on port", port)
	}
	return http.ListenAndServe(":"+strconv.Itoa(port), a.Router)
}

func (a *App) Group(prefix string) *Group {
	return &Group{
		Prefix: prefix,
		router: a.Router,
	}
}

func (a *App) GET(path string, handlers  ...HandlerFunc) {
	a.Router.GET(path, handlers...)
}

func (g *Group) GET(path string, handlers ...HandlerFunc) {
	fullPath := g.Prefix + path
	g.router.GET(fullPath, handlers...)
}

func (a *App) POST(path string, handlers  ...HandlerFunc) {
	a.Router.POST(path, handlers...)
}
func (g *Group) POST(path string, handlers ...HandlerFunc) {
	fullPath := g.Prefix + path
	g.router.POST(fullPath, handlers...)
}
func (a *App) PUT(path string, handlers  ...HandlerFunc) {
	a.Router.PUT(path, handlers...)
}
func (g *Group) PUT(path string, handlers ...HandlerFunc) {
	fullPath := g.Prefix + path
	g.router.PUT(fullPath, handlers...)
}

func (a *App) DELETE(path string, handlers  ...HandlerFunc) {
	a.Router.DELETE(path, handlers...)
}
func (g *Group) DELETE(path string, handlers ...HandlerFunc) {
	fullPath := g.Prefix + path
	g.router.DELETE(fullPath, handlers...)
}

func (a *App) Use(middleware ...HandlerFunc) {
	a.Router.Use(middleware...)
}
func (g *Group) Use(middleware ...HandlerFunc) {
	g.router.Use(middleware...)
}