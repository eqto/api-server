package apims

import (
	"errors"
	"fmt"
	"log"

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

	isProduction bool

	cn *db.Connection

	logD func(v ...interface{})
	logW func(v ...interface{})
	logE func(v ...interface{})
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
	s.debug(fmt.Sprintf(`add route %s %s`, method, path))
	return route, nil
}

//AddDataRoute add query with action result to property named "data"
func (s *Server) AddDataRoute(method, path, query, params string) (*Route, error) {
	r, e := s.AddRoute(method, path)
	if e != nil {
		return nil, e
	}
	_, e = r.AddQueryAction(query, params, `data`)
	if e != nil {
		return r, e
	}
	return r, nil
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
func (s *Server) Execute(method, url, header, body []byte) (Response, error) {
	req, e := parseRequest(method, url, header, body)
	if e != nil {
		return s.newErrorResponse(StatusBadRequest, e)
	}
	route, e := s.GetRoute(string(method), req.URL().Path)
	if e != nil { //route not found
		return s.newErrorResponse(StatusNotFound, e)

	}
	resp := s.newResponse(StatusOK)
	tx, e := s.cn.Begin()
	if e != nil { //db error
		return s.newErrorResponse(StatusInternalServerError, e)
	}
	defer tx.Commit()

	//TODO add session
	ctx := &context{tx: tx, req: req, resp: resp, vars: json.Object{}}
	for _, action := range route.action {
		result, e := action.execute(ctx)
		if e != nil {
			tx.Rollback()
			return s.newErrorResponse(StatusInternalServerError, e)
		}
		if prop := action.property(); prop != `` {
			ctx.put(prop, result)
		}
	}

	return resp, nil
}

//SetProduction ...
func (s *Server) SetProduction() {
	s.isProduction = true
}

//SetDebug ...
func (s *Server) SetDebug() {
	s.isProduction = false
}

//SetLogger ...
func (s *Server) SetLogger(debug func(v ...interface{}), warn func(v ...interface{}), err func(v ...interface{})) {
	s.logD = debug
	s.logW = warn
	s.logE = err
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

func (s *Server) newErrorResponse(status uint16, err error) (*response, error) {
	resp := s.newResponse(status)
	resp.setError(err)
	return resp, err
}

func (s *Server) newResponse(status uint16) *response {
	return &response{server: s, status: status, header: Header{}, Object: json.Object{}}
}

func (s *Server) debug(v ...interface{}) {
	if !s.isProduction {
		s.logD(v...)
	}
}
func (s *Server) warn(v ...interface{}) {
	s.logW(v...)
}
func (s *Server) error(v ...interface{}) {
	s.logW(v...)
}

//NewServer ...
func NewServer() *Server {
	s := &Server{
		routeMap: make(map[int8]map[string]*Route),
		logD:     log.Println,
		logW:     log.Println,
		logE:     log.Println,
	}
	s.routeMap[routeMethodGet] = make(map[string]*Route)
	s.routeMap[routeMethodPost] = make(map[string]*Route)
	return s
}
