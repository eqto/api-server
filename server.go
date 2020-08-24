package apims

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
)

const (
	//MethodGet GET
	MethodGet = `GET`
	//MethodPost POST
	MethodPost = `POST`

	StatusBadRequest   = 400
	StatusUnauthorized = 401
	StatusNotFound     = 404
	StatusOK           = 200

	StatusInternalServerError = 500
)

//Server ...
type Server struct {
	//index = routeMethodGet or routeMethodPost
	routeMap map[int8]map[string]*Route

	defaultContentType string

	isProduction bool

	cn          *db.Connection
	dbConnected bool

	middleware []middleware

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
	if e := s.cn.Connect(); e != nil {
		return e
	}
	s.dbConnected = true
	return nil
}

//SetDatabase ...
func (s *Server) SetDatabase(host string, port uint16, username, password, name string) {
	s.cn = db.NewEmptyConnection(host, port, username, password, name)
}

//AddMiddleware ..
func (s *Server) AddMiddleware(m Middleware) {
	s.middleware = append(s.middleware, middleware{f: m, isAuth: false})
}

//AddAuthMiddleware ..
func (s *Server) AddAuthMiddleware(m Middleware) {
	s.middleware = append(s.middleware, middleware{f: m, isAuth: true})
}

//AddPostRoute ...
func (s *Server) AddPostRoute(path string) (*Route, error) {
	return s.AddRoute(MethodPost, path)
}

//AddSecureFunc add secure route with single func action, secure means this route will validated using auth middlewares if any.
func (s *Server) AddSecureFunc(f ActionFunc) (*Route, error) {
	ptr := reflect.ValueOf(f).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	if strings.Count(name, `.`) > 1 {
		return nil, errors.New(`unsupported add inline function`)
	}
	name = name[strings.IndexRune(name, '.')+1:]
	r, e := s.AddPostRoute(`/` + name)
	if e != nil {
		return nil, e
	}
	_, e = r.AddFuncAction(f, `data`)
	if e != nil {
		return r, e
	}
	return r, nil
}

//AddFunc add insecure route with single func action, insecure means this route will not validated by auth middlewares.
func (s *Server) AddFunc(f ActionFunc) (*Route, error) {
	r, e := s.AddFunc(f)
	if r != nil {
		r.secure = false
	}
	return r, e
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
	route := &Route{path: path, method: idx, secure: true}
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
	if s.cn == nil {
		return s.newErrorResponse(StatusInternalServerError, errors.New(`no database connection, call SetDatabase or OpenDatabase first`))
	}
	if !s.dbConnected {
		return s.newErrorResponse(StatusInternalServerError, errors.New(`database not opened, call Connect() to open database`))
	}
	req, e := parseRequest(method, url, header, body)
	if e != nil {
		return s.newErrorResponse(StatusBadRequest, e)
	}

	route, e := s.GetRoute(string(method), req.URL().Path)
	if e != nil { //route not found
		return s.newErrorResponse(StatusNotFound, e)
	}

	reqCtx := newRequestCtx(req)

	if s.middleware != nil {
		for _, m := range s.middleware {
			if m.isAuth {
				if route.secure {
					if e := m.f(reqCtx); e != nil {
						return s.newErrorResponse(StatusUnauthorized, e)
					}
				}
			} else {
				if e := m.f(reqCtx); e != nil {
					return s.newErrorResponse(StatusInternalServerError, e)
				}
			}
		}
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
