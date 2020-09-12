package api

const (
	routeMethodGet  int8 = 1
	routeMethodPost int8 = 2
)

//Route ...
type Route struct {
	path   string
	method int8
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
func (r *Route) AddFuncAction(f ActionFunc, property string) (Action, error) {
	act, e := newFuncAction(f, property)
	if e != nil {
		return nil, e
	}
	r.action = append(r.action, act)
	return act, nil
}
