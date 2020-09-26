package api

import (
	"strings"
)

const (
	routeMethodGet  int8 = 1
	routeMethodPost int8 = 2
)

//Route ...
type Route struct {
	path   string
	method string
	action []Action
	secure bool
}

//SetSecure ...
func (r *Route) SetSecure(secure bool) {
	r.secure = secure
}

//AddQueryAction ...
func (r *Route) AddQueryAction(query, params, property string) (Action, error) {
	act, e := newQueryAction(query, property, params)
	if e != nil {
		return nil, e
	}
	r.action = append(r.action, act)
	return act, nil
}

//AddFuncAction ...
func (r *Route) AddFuncAction(f func(ctx Context) (interface{}, error), property string) (Action, error) {
	act, e := newFuncAction(f, property)
	if e != nil {
		return nil, e
	}
	r.action = append(r.action, act)
	return act, nil
}

//NewRoute create route
func NewRoute(method, path string) *Route {
	method = strings.ToUpper(method)
	return &Route{path: path, method: method, secure: true}
}

//NewFuncRoute create POST route with single func action
func NewFuncRoute(path string, f func(ctx Context) (interface{}, error)) *Route {
	route := NewRoute(MethodPost, path)
	route.AddFuncAction(f, `data`)
	return route
}
