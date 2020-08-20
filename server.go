package apims

import (
	"fmt"
)

const (
	MethodGet  = `GET`
	MethodPost = `POST`
)

//Server ...
type Server struct {
	routeMap map[string]*Route
}

//NewRoute ...
func (s *Server) NewRoute(method, path string) (*Route, error) {
	route := &Route{path: path}
	switch method {
	case MethodGet:
		route.method = routeMethodGet
	case MethodPost:
		route.method = routeMethodPost
	default:
		return nil, fmt.Errorf(`unrecognized method %s, choose between apims.MethodGet or apims.MethodPost`, method)
	}
	s.routeMap[fmt.Sprintf(`%s-%s`, method, path)] = route
	return route, nil
}

//NewServer ...
func NewServer() *Server {
	return &Server{routeMap: make(map[string]*Route)}
}
