package apims

import (
	"errors"
	"fmt"

	"github.com/eqto/go-db"
)

const (
	//MethodGet GET
	MethodGet = `GET`
	//MethodPost POST
	MethodPost = `POST`

	StatusBadRequest = 400
	StatusNotFound   = 404
	StatusOK         = 200
)

//Server ...
type Server struct {
	//index = routeMethodGet or routeMethodPost
	routeMap map[int8]map[string]*Route

	defaultContentType string

	db struct {
		host     string
		port     uint16
		username string
		password string
		name     string
	}
	cn *db.Connection
}

//OpenDatabase call SetDatabase and Connect
func (s *Server) OpenDatabase(host string, port uint16, username, password, name string) error {
	s.SetDatabase(host, port, username, password, name)
	return s.Connect()
}

//Connect ...
func (s *Server) Connect() error {
	cn, e := db.NewConnection(s.db.host, s.db.port, s.db.username, s.db.password, s.db.name)
	if e != nil {
		return e
	}
	if e := cn.Ping(); e != nil {
		return e
	}
	s.cn = cn
	return nil
}

//SetDatabase ...
func (s *Server) SetDatabase(host string, port uint16, username, password, name string) {
	s.db.host = host
	s.db.port = uint16(port)
	s.db.username = username
	s.db.password = password
	s.db.name = name
}

//NewRoute ...
func (s *Server) NewRoute(method, path string) (*Route, error) {
	if s.routeMap == nil {
		return nil, errors.New(`unable to create route, please use NewServer() to create new server`)
	}
	idx, e := s.routeMethod(method, path)
	if e != nil {
		return nil, e
	}
	route := &Route{path: path, method: idx}
	s.routeMap[idx][path] = route
	return route, nil
}

//GetRoute ...
func (s *Server) GetRoute(method, path string) (*Route, error) {
	idx, e := s.routeMethod(method, path)
	if e != nil {
		return nil, e
	}
	if r, ok := s.routeMap[idx][path]; ok {
		return r, nil
	}
	return nil, fmt.Errorf(`route %s %s not found`, method, path)
}

//SetRoute ...
func (s *Server) SetRoute(method, path string, route *Route) error {
	if s.routeMap == nil {
		return errors.New(`unable to set route, please use NewServer() to create new server`)
	}
	idx, e := s.routeMethod(method, path)
	if e != nil {
		return e
	}
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

//NewServer ...
func NewServer() *Server {
	s := &Server{routeMap: make(map[int8]map[string]*Route)}
	s.routeMap[routeMethodGet] = make(map[string]*Route)
	s.routeMap[routeMethodPost] = make(map[string]*Route)
	return s
}
