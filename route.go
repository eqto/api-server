package api

// Route ...
type Route struct {
	action []Action
	secure bool
	group  string

	isWs   bool
	logger *logger
}

// Secure ...
func (r *Route) Secure() *Route {
	r.secure = true
	return r
}

// UseGroup only use middleware that have the same name or no name
func (r *Route) UseGroup(name string) *Route {
	r.group = name
	return r
}

// AddQueryAction ...
func (r *Route) AddQueryAction(query, params string) Action {
	act, e := newQueryAction(query, params)
	if e != nil {
		if r.logger != nil {
			r.logger.E(e)
		}
		return act
	}
	act.AssignTo(`data`)
	r.action = append(r.action, act)
	return act
}

// AddAction ...
func (r *Route) AddAction(f func(Context) error) Action {
	act := newFuncAction(f).AssignTo(`data`)
	r.action = append(r.action, act)
	return act
}

func (r *Route) execute(s *Server, ctx *context) error {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				ctx.StatusInternalServerError(r.Error())
			case string:
				ctx.StatusInternalServerError(r)
			}
		}
	}()
	if r.isWs {
		s.wsServ.Upgrade(ctx.fastCtx)
	} else {
		for _, action := range r.action {
			ctx.property = action.property()
			if e := action.execute(ctx); e != nil {
				return e
			}
			if ctx.resp.stop {
				return nil
			}
		}
	}
	return nil
}
