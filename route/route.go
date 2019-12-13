package route

import "net/http"

const (
	MethodGet  = http.MethodGet
	MethodPost = http.MethodPost
)

//Route ...
type Route struct {
	path   string
	method string
}

//New ...
func New(path, method string) *Route {
	return &Route{
		path:   path,
		method: method,
	}
}
