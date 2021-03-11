package api

import (
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

//Group ..
type Group struct {
	s    *Server
	name string
}

func (g *Group) newRoute() *Route {
	return new(Route).UseGroup(g.name)
}

//Post ..
func (g *Group) Post(path string, f func(Context) (interface{}, error)) *Route {
	route := g.newRoute()
	route.AddFuncAction(f, `data`)
	g.s.SetRoute(MethodPost, path, route)
	return route
}

//PostSecure ..
func (g *Group) PostSecure(path string, f func(Context) (interface{}, error)) *Route {
	return g.Post(path, f).Secure()
}

//Query add route with single query action. When secure is true, this route will validated using auth middlewares if any.
func (g *Group) Query(path, query, params string) (*Route, error) {
	route := g.newRoute()
	if _, e := route.AddQueryAction(query, params, `data`); e != nil {
		return nil, e
	}
	g.s.SetRoute(MethodPost, path, route)
	return route, nil
}

//QuerySecure ..
func (g *Group) QuerySecure(path, query, params string) (*Route, error) {
	route, e := g.Query(path, query, params)
	if e != nil {
		return nil, e
	}
	return route.Secure(), nil
}

//Func add route with single func action. When secure is true, this route will validated using auth middlewares if any.
func (g *Group) Func(f func(Context) (interface{}, error)) (*Route, error) {
	ptr := reflect.ValueOf(f).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	name = filepath.Base(name)
	if strings.Count(name, `.`) > 1 {
		g.s.debug(`unsupported add inline function`, name)
		return nil, errors.New(`unsupported add inline function`)
	}
	name = strings.ReplaceAll(name, `.`, `/`)
	route := g.newRoute()
	route.AddFuncAction(f, `data`)
	g.s.SetRoute(MethodPost, `/`+name, route)
	return route, nil
}

//FuncSecure ..
func (g *Group) FuncSecure(f func(Context) (interface{}, error)) (*Route, error) {
	route, e := g.s.Func(f)
	if e != nil {
		return nil, e
	}
	return route.Secure(), nil
}

//AddMiddleware ..
func (g *Group) AddMiddleware(f func(Context) error) Middleware {
	m := &middlewareContainer{f: f, group: g.name}
	g.s.middlewares = append(g.s.middlewares, m)
	return m
}
