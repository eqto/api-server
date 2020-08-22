package apims

import (
	"errors"
	"fmt"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
)

const (
	//MethodGet GET
	MethodGet = `GET`
	//MethodPost POST
	MethodPost = `POST`

	StatusBadRequest = 400
	StatusNotFound   = 404
	StatusOK         = 200

	StatusInternalServerError = 500
)

//Server ...
type Server struct {
	//index = routeMethodGet or routeMethodPost
	routeMap map[int8]map[string]*Route

	defaultContentType string

	cn *db.Connection
}

//OpenDatabase call SetDatabase and Connect
func (s *Server) OpenDatabase(host string, port uint16, username, password, name string) error {
	s.SetDatabase(host, port, username, password, name)
	return s.Connect()
}

//Connect ...
func (s *Server) Connect() error {
	return s.cn.Connect()
}

//SetDatabase ...
func (s *Server) SetDatabase(host string, port uint16, username, password, name string) {
	s.cn = db.NewEmptyConnection(host, port, username, password, name)
}

//AddRoute ...
func (s *Server) AddRoute(method, path string) (*Route, error) {
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

//Execute request and Response header and body
func (s *Server) Execute(method, url, contentType, body []byte) (Response, error) {
	req, e := parseRequest(url, contentType, body)
	if e != nil {
		return newResponse(StatusBadRequest), e
	}
	route, e := s.GetRoute(string(method), req.URL().Path)
	if e != nil { //route not found
		return newResponse(StatusNotFound), e

	}
	resp := newResponse(StatusOK)
	tx, e := s.cn.Begin()
	if e != nil { //db error
		return newResponse(StatusInternalServerError), e
	}
	defer tx.Commit()

	//TODO add session
	ctx := &context{tx: tx, req: req, resp: resp, vars: json.Object{}}
	for _, action := range route.action {
		result, e := action.execute(ctx)
		if e != nil {
			tx.Rollback()
			return newResponse(StatusInternalServerError), e
		}
		if prop := action.property(); prop != `` {
			ctx.put(prop, result)
		}
	}

	return resp, nil
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
