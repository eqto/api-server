package apims

const (
	routeMethodGet  int8 = 1
	routeMethodPost int8 = 2
)

//Route ...
type Route struct {
	path   string
	method int8
	action []*Action
}

//AddQueryList ...
func (r *Route) AddQueryList(query, params, property string) *Action {
	act := &Action{command: query, params: params, property: property, typ: actionTypeList}
	r.addAction(act)
	return act
}

//AddQueryGet ...
func (r *Route) AddQueryGet(query, params, property string) *Action {
	act := &Action{command: query, params: params, property: property, typ: actionTypeGet}
	r.addAction(act)
	return act
}

func (r *Route) addAction(act *Action) {
	r.action = append(r.action, act)
}
