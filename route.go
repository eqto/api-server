package api

//Route ...
type Route struct {
	action []Action
	secure bool
	group  string

	logger *logger
}

//Secure ...
func (r *Route) Secure() *Route {
	r.secure = true
	return r
}

//UseGroup only use middleware that have the same name or no name
func (r *Route) UseGroup(name string) *Route {
	r.group = name
	return r
}

//AddQueryAction ...
func (r *Route) AddQueryAction(property, query, params string) *Route {
	act, e := newQueryAction(property, query, params)
	if e != nil {
		if r.logger != nil {
			r.logger.W(e)
		}
		return r
	}
	r.action = append(r.action, act)
	return r
}

//AddAction ...
func (r *Route) AddAction(property string, f func(Context) error) *Route {
	act := newFuncAction(f, property)
	r.action = append(r.action, act)
	return r
}

func (r *Route) execute(s *Server, ctx *context) error {
	// if s.routeAuthenticator != nil {
	// 	for _, m := range s.routeAuthenticator {
	// 		if r.secure {
	// 			if e := m(ctx); e != nil {
	// 				ctx.resp.setError(StatusUnauthorized, e)
	// 				return e
	// 			}
	// 		}
	// 	}
	// }

	for _, action := range r.action {
		ctx.property = action.property()
		if e := action.execute(ctx); e != nil {
			ctx.setErr(e)
			return e
		}
		if ctx.resp.stop {
			return nil
		}
	}
	return nil
}
