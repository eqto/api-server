package api

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

var (
	normalizeRegex *regexp.Regexp
)

//Group ..
type Group struct {
	s    *Server
	name string
}

func (g *Group) Get(path string) *Route {
	return g.getRoute(MethodGet, path)
}

func (g *Group) GetAction(f func(Context) error) *Route {
	return g.action(MethodGet, f)
}

func (g *Group) GetSecureAction(f func(Context) error) *Route {
	return g.action(MethodGet, f).Secure()
}

func (g *Group) Post(path string) *Route {
	return g.getRoute(MethodPost, path)
}

func (g *Group) PostAction(f func(Context) error) *Route {
	return g.action(MethodPost, f)
}

func (g *Group) PostSecureAction(f func(Context) error) *Route {
	return g.action(MethodPost, f).Secure()
}

//AddMiddleware ..
func (g *Group) AddMiddleware(f func(Context) error) Middleware {
	m := &middlewareContainer{f: f, group: g.name}
	g.s.middlewares = append(g.s.middlewares, m)
	return m
}

func (g *Group) action(method string, f func(Context) error) *Route {
	ptr := reflect.ValueOf(f).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	name = filepath.Base(name)
	if strings.Count(name, `.`) > 1 {
		g.s.logger.E(`unsupported add inline function`, name)
		return &Route{logger: g.s.logger}
	}
	path := strings.ReplaceAll(name, `.`, `/`)
	route := g.getRoute(method, `/`+path)
	return route.AddAction(`data`, f)
}

func (g *Group) getRoute(method, path string) *Route {
	switch method {
	case MethodGet, MethodPost:
	default:
		return nil
	}
	if g.s.normalize {
		path = normalizePath(path)
	}
	route, ok := g.s.routeMap[method][path]
	if !ok {
		route = &Route{logger: g.s.logger}
		g.s.routeMap[method][path] = route
		g.s.logger.D(fmt.Sprintf(`Register route: %s %s`, method, path))
	}
	route.UseGroup(g.name)
	return route
}

func normalizePath(path string) string {
	if normalizeRegex == nil {
		normalizeRegex = regexp.MustCompile(`([A-Z]+)`)
	}
	path = normalizeRegex.ReplaceAllString(path, `_$1`)
	path = strings.ToLower(path)
	validPath := false
	if strings.HasPrefix(path, `/`) {
		validPath = true
		path = path[1:]
	}
	path = strings.TrimPrefix(path, `_`)

	if validPath {
		path = `/` + path
	}
	path = strings.ReplaceAll(path, `/_`, `/`)
	return path
}
