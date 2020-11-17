package api

import (
	"github.com/eqto/go-json"
)

const (
	routeMethodGet  int8 = 1
	routeMethodPost int8 = 2
)

//Route ...
type Route struct {
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

func (r *Route) execute(s *Server, reqCtx *requestCtx) (Response, error) {
	if s.authMiddleware != nil {
		for _, m := range s.authMiddleware {
			if e := reqCtx.begin(); e != nil {
				return newResponseError(StatusInternalServerError, e)
			}
			defer reqCtx.rollback()
			if r.secure {
				if e := m.f(reqCtx); e != nil {
					reqCtx.rollback()
					return newResponseError(StatusUnauthorized, e)
				}
			}
			reqCtx.commit()
		}
	}
	if s.middleware != nil {
		for _, m := range s.middleware {
			if e := reqCtx.begin(); e != nil {
				return newResponseError(StatusInternalServerError, e)
			}
			defer reqCtx.rollback()
			if e := m.f(reqCtx); e != nil {
				reqCtx.rollback()
				return newResponseError(StatusInternalServerError, e)
			}
			reqCtx.commit()
		}
	}

	resp := newResponse(StatusOK)

	if e := reqCtx.begin(); e != nil {
		return newResponseError(StatusInternalServerError, e)
	}
	defer reqCtx.commit()

	//TODO add session
	ctx := &context{server: s, tx: reqCtx.tx, req: reqCtx.req, resp: resp, vars: json.Object{}, sess: reqCtx.sess}

	for _, action := range r.action {
		result, e := action.execute(ctx)
		if result != nil {
			if resp, ok := result.(Response); ok {
				if resp.Status() != StatusOK && e == nil {
					url := ctx.req.url.RawPath
					if ctx.req.url.RawQuery != `` {
						url += `?` + ctx.req.url.RawQuery
					}
					e = resp.Error()
				}
			}
		}
		if e != nil {
			reqCtx.rollback()
			if result != nil {
				if resp, ok := result.(Response); ok {
					return resp, e
				}
			}
			return newResponseError(StatusInternalServerError, e)
		}
		if prop := action.property(); prop != `` {
			ctx.put(prop, result)
		}
	}

	return resp, nil
}

//NewRoute create route
func NewRoute() *Route {
	return &Route{secure: true}
}

//NewFuncRoute create POST route with single func action
func NewFuncRoute(f func(ctx Context) (interface{}, error)) *Route {
	route := NewRoute()
	route.AddFuncAction(f, `data`)
	return route
}
