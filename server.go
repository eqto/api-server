package api

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

const (
	//MethodGet GET
	MethodGet string = `GET`
	//MethodPost POST
	MethodPost string = `POST`

	//StatusBadRequest ...
	StatusBadRequest = 400
	//StatusUnauthorized ...
	StatusUnauthorized = 401
	//StatusNotFound ...
	StatusNotFound = 404
	//StatusOK ...
	StatusOK = 200

	//StatusInternalServerError ...
	StatusInternalServerError = 500
)

//Server ...
type Server struct {
	routeMap map[string]map[string]*Route

	defaultContentType string
	normalize          bool

	isProduction bool

	cn          *db.Connection
	dbConnected bool

	middleware []middleware

	logD func(v ...interface{})
	logW func(v ...interface{})
	logE func(v ...interface{})
}

//Database ...
func (s *Server) Database() *db.Connection {
	return s.cn
}

//OpenDatabase call SetDatabase and Connect
func (s *Server) OpenDatabase(host string, port int, username, password, name string) error {
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
func (s *Server) SetDatabase(host string, port int, username, password, name string) {
	s.cn, _ = db.NewEmptyConnection(host, port, username, password, name)
}

//AddMiddleware ..
func (s *Server) AddMiddleware(m Middleware) {
	s.middleware = append(s.middleware, middleware{f: m, isAuth: false})
}

//AddAuthMiddleware ..
func (s *Server) AddAuthMiddleware(m Middleware) {
	s.middleware = append(s.middleware, middleware{f: m, isAuth: true})
}

//SetRoute ...
func (s *Server) SetRoute(route *Route) {
	if s.routeMap == nil {
		s.routeMap = make(map[string]map[string]*Route)
		s.routeMap[MethodGet] = make(map[string]*Route)
		s.routeMap[MethodPost] = make(map[string]*Route)
	}
	path := route.path
	if s.normalize {
		path = s.normalizePath(path)
		route.path = path
	}
	s.routeMap[route.method][path] = route
	s.debug(fmt.Sprintf(`add route %s %s`, route.method, route.path))
}

//AddFuncRoute add route with single func action. When secure is true, this route will validated using auth middlewares if any.
func (s *Server) AddFuncRoute(f ActionFunc, secure bool) (*Route, error) {
	ptr := reflect.ValueOf(f).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	if strings.Count(name, `.`) > 1 {
		return nil, errors.New(`unsupported add inline function`)
	}
	name = name[strings.IndexRune(name, '.')+1:]
	route := NewRoute(MethodPost, `/`+name)
	if _, e := route.AddFuncAction(f, `data`); e != nil {
		return nil, e
	}
	s.SetRoute(route)
	return route, nil
}

//AddQueryRoute add route with single query action. When secure is true, this route will validated using auth middlewares if any.
func (s *Server) AddQueryRoute(path, query, params string, secure bool) (*Route, error) {
	route := NewRoute(MethodPost, path)
	if _, e := route.AddQueryAction(query, params, `data`); e != nil {
		return nil, e
	}
	s.SetRoute(route)
	return route, nil
}

//GetRoute ...
func (s *Server) GetRoute(method, path string) (*Route, error) {
	method = strings.ToUpper(method)
	if r, ok := s.routeMap[method][path]; ok {
		return r, nil
	}
	return nil, fmt.Errorf(`route %s %s not found`, method, path)
}

//NormalizeFunc if yes from this and beyond all Func added will renamed to lowercase, separated with underscore. Ex: HelloWorld registered as hello_world
func (s *Server) NormalizeFunc(n bool) {
	s.normalize = n
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

	sess := &session{}

	reqCtx := newRequestCtx(s.cn, req, sess)

	if s.middleware != nil {
		for _, m := range s.middleware {
			if e := reqCtx.begin(); e != nil {
				return s.newErrorResponse(StatusInternalServerError, e)
			}
			defer reqCtx.rollback()
			if m.isAuth {
				if route.secure {
					if e := m.f(reqCtx); e != nil {
						reqCtx.rollback()
						return s.newErrorResponse(StatusUnauthorized, e)
					}
				}
			} else {
				if e := m.f(reqCtx); e != nil {
					reqCtx.rollback()
					return s.newErrorResponse(StatusInternalServerError, e)
				}
			}
			reqCtx.commit()
		}
	}

	resp := s.newResponse(StatusOK)

	if e := reqCtx.begin(); e != nil {
		return s.newErrorResponse(StatusInternalServerError, e)
	}
	defer reqCtx.commit()

	//TODO add session
	ctx := &context{tx: reqCtx.tx, req: req, resp: resp, vars: json.Object{}, sess: sess}

	for _, action := range route.action {
		result, e := action.execute(ctx)

		if e != nil {
			reqCtx.rollback()
			return s.newErrorResponse(StatusInternalServerError, e)
		}
		if prop := action.property(); prop != `` {
			ctx.put(prop, result)
		}
	}

	return resp, nil
}

//Serve ...
func (s *Server) Serve(port int) error {
	return fasthttp.ListenAndServe(fmt.Sprintf(`:%d`, port), func(ctx *fasthttp.RequestCtx) {
		resp, e := s.Execute(ctx.Method(), ctx.RequestURI(), ctx.Request.Header.RawHeaders(), ctx.Request.Body())
		if e != nil {
			s.logW(e)
		}
		ctx.SetStatusCode(resp.Status())
		for key, valArr := range resp.Header() {
			if len(valArr) > 0 {
				ctx.Response.Header.Set(key, valArr[0])
			} else {
				ctx.Response.Header.Set(key, ``)
			}
		}
		ctx.SetBody(resp.Body())
	})
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
		return 0, fmt.Errorf(`unrecognized method %s, choose between api.MethodGet or api.MethodPost`, method)
	}
}

func (s *Server) normalizePath(path string) string {
	regex := regexp.MustCompile(`([A-Z]+)`)
	path = regex.ReplaceAllString(path, `_$1`)
	path = strings.ToLower(path)
	validPath := false
	if strings.HasPrefix(path, `/`) {
		validPath = true
		path = path[1:]
	}
	if strings.HasPrefix(path, `_`) {
		path = path[1:]
	}
	if validPath {
		path = `/` + path
	}
	return path
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

//New ...
func New() *Server {
	s := &Server{
		routeMap: make(map[string]map[string]*Route),
		logD:     log.Println,
		logW:     log.Println,
		logE:     log.Println,
	}
	s.routeMap[MethodGet] = make(map[string]*Route)
	s.routeMap[MethodPost] = make(map[string]*Route)
	return s
}
