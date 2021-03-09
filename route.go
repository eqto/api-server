package api

const (
	routeMethodGet  int8 = 1
	routeMethodPost int8 = 2
)

//Route ...
type Route struct {
	action         []Action
	secure         bool
	middlewareName string
}

//Secure ...
func (r *Route) Secure() *Route {
	r.secure = true
	return r
}

//UseGroup only use middleware that have the same name or no name
func (r *Route) UseGroup(name string) *Route {
	r.middlewareName = name
	return r
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
func (r *Route) AddFuncAction(f func(Context) (interface{}, error), property string) Action {
	act := newFuncAction(f, property)
	r.action = append(r.action, act)
	return act
}

func (r *Route) execute(s *Server, ctx *context) error {
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
			return e
		}
	}
	return nil
}
