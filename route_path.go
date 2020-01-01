package api

//RoutePath ...
type RoutePath struct {
	method string
	path   string
	routes []*Route
}

//AddRoute ...
func (r *RoutePath) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
}

//Routes ...
func (r *RoutePath) Routes() []*Route {
	return r.routes
}
