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
	app    *App // <-- change from router *Router
}

type RouteFunc func(app *App)

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
		app:    a,
	}
}
func (a *App) GET(path string, handlers ...HandlerFunc) {
	a.Router.GET(path, handlers...)
}

func (a *App) POST(path string, handlers ...HandlerFunc) {
	a.Router.POST(path, handlers...)
}

func (a *App) PUT(path string, handlers ...HandlerFunc) {
	a.Router.PUT(path, handlers...)
}

func (a *App) DELETE(path string, handlers ...HandlerFunc) {
	a.Router.DELETE(path, handlers...)
}
func (g *Group) GET(path string, handlers ...HandlerFunc) {
	full := g.Prefix + path
	g.app.GET(full, handlers...)
}
func (g *Group) POST(path string, handlers ...HandlerFunc) {
	full := g.Prefix + path
	g.app.POST(full, handlers...)
}

func (g *Group) PUT(path string, handlers ...HandlerFunc) {
	full := g.Prefix + path
	g.app.PUT(full, handlers...)
}
func (g *Group) DELETE(path string, handlers ...HandlerFunc) {
	full := g.Prefix + path
	g.app.DELETE(full, handlers...)
}
func (a *App) Use(middleware ...HandlerFunc) {
	a.Router.Use(middleware...)
}
func (g *Group) Use(middleware ...HandlerFunc) {
	g.app.Router.Use(middleware...)
}
func (a *App) UseRoutes(routes ...RouteFunc) {
	for _, route := range routes {
		route(a)
	}
}
func (g *Group) UseRoutes(routes ...RouteFunc) {
	for _, route := range routes {
		route(g.app)
	}
}
