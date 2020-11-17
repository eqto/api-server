package api

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
func (r *Route) AddFuncAction(f func(ctx Ctx) (interface{}, error), property string) (Action, error) {
	act, e := newFuncAction(f, property)
	if e != nil {
		return nil, e
	}
	r.action = append(r.action, act)
	return act, nil
}

func (r *Route) execute(s *Server, ctx *ctx) (Response, error) {
	if e := ctx.begin(); e != nil {
		return newResponseError(StatusInternalServerError, e)
	}
	defer ctx.commit()

	if s.routeAuthenticator != nil {
		for _, m := range s.routeAuthenticator {
			if r.secure {
				if e := m(ctx); e != nil {
					ctx.rollback()
					return newResponseError(StatusUnauthorized, e)
				}
			}
		}
	}
	resp := newResponse(StatusOK)

	for _, action := range r.action {
		result, e := action.execute(ctx)
		if result != nil {
			if resp, ok := result.(Response); ok {
				if resp.Status() != StatusOK && e == nil {
					e = resp.Error()
				}
			}
		}
		if e != nil {
			ctx.rollback()
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
func NewFuncRoute(f func(ctx Ctx) (interface{}, error)) *Route {
	route := NewRoute()
	route.AddFuncAction(f, `data`)
	return route
}
