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
func (r *Route) AddFuncAction(f func(ctx Context) (interface{}, error), property string) (Action, error) {
	act, e := newFuncAction(f, property)
	if e != nil {
		return nil, e
	}
	r.action = append(r.action, act)
	return act, nil
}

func (r *Route) execute(s *Server, ctx *context) error {
	if e := ctx.begin(); e != nil {
		ctx.resp.setError(StatusInternalServerError, e)
		return e
	}
	defer func() {
		if ctx.resp.err != nil {
			ctx.rollback()
		} else {
			ctx.commit()
		}
	}()

	if s.routeAuthenticator != nil {
		for _, m := range s.routeAuthenticator {
			if r.secure {
				if e := m(ctx); e != nil {
					ctx.resp.setError(StatusUnauthorized, e)
					return e
				}
			}
		}
	}

	for _, action := range r.action {
		if result, e := action.execute(ctx); e == nil {
			if prop := action.property(); prop != `` {
				ctx.put(prop, result)
			}
		} else {
			ctx.resp.setError(StatusInternalServerError, e)
		}
	}
	return nil
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
