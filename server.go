package apims

import (
	"fmt"
)

const (
	//MethodGet GET
	MethodGet = `GET`
	//MethodPost POST
	MethodPost = `POST`
)

//Server ...
type Server struct {
	//index = routeMethodGet or routeMethodPost
	routeMap map[int8]map[string]*Route
}

//NewRoute ...
func (s *Server) NewRoute(method, path string) (*Route, error) {
	idx, e := s.routeMethod(method, path)
	if e != nil {
		return nil, e
	}
	route := &Route{path: path, method: idx}
	s.init()
	s.routeMap[idx][path] = route
	return route, nil
}

//GetRoute ...
func (s *Server) GetRoute(method, path string) (*Route, error) {
	idx, e := s.routeMethod(method, path)
	if e != nil {
		return nil, e
	}
	s.init()
	if r, ok := s.routeMap[idx][path]; ok {
		return r, nil
	}
	return nil, fmt.Errorf(`route %s %s not found`, method, path)
}

//SetRoute ...
func (s *Server) SetRoute(method, path string, route *Route) error {
	idx, e := s.routeMethod(method, path)
	if e != nil {
		return e
	}
	s.init()
	s.routeMap[idx][path] = route
	return nil
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

func (s *Server) init() {
	if s.routeMap == nil {
		s.routeMap = make(map[int8]map[string]*Route)
		s.routeMap[routeMethodGet] = make(map[string]*Route)
		s.routeMap[routeMethodPost] = make(map[string]*Route)
	}
}

//NewServer ...
func NewServer() *Server {
	s := &Server{}
	s.init()
	return s
}
