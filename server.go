package api

import (
	"errors"
	"fmt"
	uri "net/url"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/eqto/go-db"
	log "github.com/eqto/go-logger"
	"github.com/valyala/fasthttp"
)

//Server ...
type Server struct {
	routeMap map[string]map[string]*Route
	proxies  []proxy

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

//Proxy ...
func (s *Server) Proxy(path, dest string) error {
	p, e := newProxy(path, dest)
	if e != nil {
		return e
	}
	s.proxies = append(s.proxies, p)
	return nil
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
func (s *Server) AddFuncRoute(f func(ctx Context) (interface{}, error), secure bool) (*Route, error) {
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
//!deprecated
func (s *Server) Execute(method, url string, header, body []byte) (Response, error) {
	req, e := parseRequest(method, url, header, body)
	if e != nil {
		return newResponseError(StatusBadRequest, e)
	}

	route, e := s.GetRoute(string(method), req.URL().Path)
	if e == nil {
		sess := &session{}
		reqCtx := newRequestCtx(s.cn, req, sess)
		return route.execute(s, reqCtx)
	}
	for _, proxy := range s.proxies {
		s.logD(string(url))
		// s.logD(proxy)
		if proxy.match(string(url)) {
			// return proxy.execute(method, url, header, body)
		}
	}
	//route not found
	return newResponseError(StatusNotFound, e)
}

func (s *Server) execute(ctx *fasthttp.RequestCtx) (Response, error) {
	method := string(ctx.Method())
	url := string(ctx.RequestURI())
	s.logD(`Request url:`, url)

	u, e := uri.Parse(url)
	if e != nil {
		return nil, e
	}
	route, e := s.GetRoute(method, u.Path)
	if e == nil {
		header := ctx.Request.Header.RawHeaders()
		body := ctx.Request.Body()
		sess := &session{}
		req, e := parseRequest(method, url, header, body)
		if e != nil {
			return newResponseError(StatusBadRequest, e)
		}
		reqCtx := newRequestCtx(s.cn, req, sess)
		return route.execute(s, reqCtx)
	}
	for _, proxy := range s.proxies {
		if proxy.match(string(url)) {
			return proxy.execute(s, ctx)
		}
	}
	return newResponseError(StatusNotFound, e)
}

//Serve ...
func (s *Server) Serve(port int) error {
	return fasthttp.ListenAndServe(fmt.Sprintf(`:%d`, port), func(ctx *fasthttp.RequestCtx) {
		resp, e := s.execute(ctx)
		if e != nil {
			s.logW(e)
		}
		if resp == nil {
			resp, _ = newResponseError(StatusInternalServerError, nil)
		}
		if resp != nil {
			ctx.SetStatusCode(resp.Status())
			for key, valArr := range resp.Header() {
				if len(valArr) > 0 {
					ctx.Response.Header.Set(key, valArr[0])
				} else {
					ctx.Response.Header.Set(key, ``)
				}
			}
			ctx.SetBody(resp.Body())
		}
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
