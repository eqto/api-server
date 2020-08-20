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
	m, e := s.routeMethod(method, path)
	if e != nil {
		return nil, e
	}
	route := &Route{path: path, method: m}
	s.routeMap[fmt.Sprintf(`%s-%s`, method, path)] = route
	return route, nil
}

func (s *Server) routeMethod(method, path string) (int8, error) {
	switch method {
	case MethodGet:
		return routeMethodGet, nil
	case MethodPost:
		return routeMethodPost, nil
	default:
		return 0, fmt.Errorf(`unrecognized method %s, choose between apims.MethodGet or apims.MethodPost`, method)
	}
}

//GetRoute ...
func (s *Server) GetRoute(method, path string) (*Route, error) {
	_, e := s.routeMethod(method, path)
	if e != nil {
		return nil, e
	}
	if r, ok := s.routeMap[fmt.Sprintf(`%s-%s`, method, path)]; ok {
		return r, nil
	}
	return nil, fmt.Errorf(`route %s %s not found`, method, path)
}

//NewServer ...
func NewServer() *Server {
	return &Server{routeMap: make(map[string]*Route)}
}
