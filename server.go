package api

import (
	"fmt"
	uri "net/url"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/eqto/go-db"
	log "github.com/eqto/go-logger"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

//Server ...
type Server struct {
	routeMap map[string]map[string]*Route
	proxies  []proxy
	files    []file

	defaultContentType string
	normalize          bool

	isProduction bool

	cn                 *db.Connection
	dbConnected        bool
	routeAuthenticator []RouteAuthenticator

	middleware     []Middleware
	respMiddleware []RespMiddleware

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
	s.middleware = append(s.middleware, m)
}

//AddRouteAuthenticator ..
func (s *Server) AddRouteAuthenticator(a RouteAuthenticator) {
	s.routeAuthenticator = append(s.routeAuthenticator, a)
}

//AddResponseMiddleware ..
func (s *Server) AddResponseMiddleware(m RespMiddleware) {
	s.respMiddleware = append(s.respMiddleware, m)
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

//FileRoute serve static file. Path parameter to determine url to be processed. Dest parameter will find directory of the file reside. RedirectTo parameter to redirect non existing file, this param can be used for SPA (ex. index.html).
func (s *Server) FileRoute(path, dest, redirectTo string) error {
	f, e := newFile(path, dest, redirectTo)
	if e != nil {
		return e
	}
	s.files = append(s.files, f)
	return nil
}

//FileRouteRemove ...
func (s *Server) FileRouteRemove(path string) error {
	for i, file := range s.files {
		if file.path == path {
			s.files = append(s.files[:i], s.files[i+1:]...)
			return nil
		}
	}
	return nil
}

//SetRoute ...
func (s *Server) SetRoute(method, path string, route *Route) {
	if s.routeMap == nil {
		s.routeMap = make(map[string]map[string]*Route)
		s.routeMap[MethodGet] = make(map[string]*Route)
		s.routeMap[MethodPost] = make(map[string]*Route)
	}
	if s.normalize {
		path = s.normalizePath(path)
	}
	s.routeMap[method][path] = route
	s.debug(fmt.Sprintf(`add route %s %s`, method, path))
}

//AddFuncRoute add route with single func action. When secure is true, this route will validated using auth middlewares if any.
func (s *Server) AddFuncRoute(f func(ctx Context) (interface{}, error), secure bool) (*Route, error) {
	ptr := reflect.ValueOf(f).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	if strings.Count(name, `.`) > 1 {
		return nil, errors.New(`unsupported add inline function`)
	}
	name = name[strings.IndexRune(name, '.')+1:]
	route := NewRoute()
	if _, e := route.AddFuncAction(f, `data`); e != nil {
		return nil, e
	}
	s.SetRoute(MethodPost, `/`+name, route)
	return route, nil
}

//AddQueryRoute add route with single query action. When secure is true, this route will validated using auth middlewares if any.
func (s *Server) AddQueryRoute(path, query, params string, secure bool) (*Route, error) {
	route := NewRoute()
	if _, e := route.AddQueryAction(query, params, `data`); e != nil {
		return nil, e
	}
	s.SetRoute(MethodPost, path, route)
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

func (s *Server) execute(fastCtx *fasthttp.RequestCtx, ctx *context) (Response, error) {
	method := string(ctx.req.Header.Method())
	url := string(ctx.req.Header.RequestURI())
	s.logD(`Request url:`, url)

	u, e := uri.Parse(url)
	if e != nil {
		return nil, e
	}
	route, e := s.GetRoute(method, u.Path)
	if e == nil {
		return route.execute(s, ctx)
	}
	for _, proxy := range s.proxies {
		if proxy.match(string(url)) {
			return proxy.execute(s, ctx)
		}
	}
	for _, file := range s.files {
		if file.match(string(url)) {
			file.handler(fastCtx)
			return nil, nil
		}
	}
	return newResponseError(StatusNotFound, e)
}

//Serve ...
func (s *Server) Serve(port int) error {
	return fasthttp.ListenAndServe(fmt.Sprintf(`:%d`, port), func(fastCtx *fasthttp.RequestCtx) {
		ctx, e := newContext(s, &fastCtx.Request, &fastCtx.Response, s.cn)
		if e != nil {
			s.logW(e)
			fastCtx.WriteString(e.Error())
			return
		}
		resp, e := s.execute(fastCtx, ctx)
		if e != nil {
			s.logW(e)
		}
		if resp != nil {
			if len(s.respMiddleware) > 0 {
				if req, e := parseRequest(string(fastCtx.Method()), string(fastCtx.RequestURI()), fastCtx.Request.Header.RawHeaders(), fastCtx.Request.Body()); e == nil {
					for _, m := range s.respMiddleware {
						m(req, resp)
					}
				} else {
					s.warn(errors.Wrap(e, `unable to parse request for response middleware`))
				}
			}

			fastCtx.SetStatusCode(resp.Status())
			for key, valArr := range resp.Header() {
				if len(valArr) > 0 {
					fastCtx.Response.Header.Set(key, valArr[0])
				} else {
					fastCtx.Response.Header.Set(key, ``)
				}
			}
			hasEncoding := fastCtx.Request.Header.HasAcceptEncoding
			if hasEncoding(`br`) {
				fasthttp.WriteBrotli(fastCtx.Response.BodyWriter(), resp.Body())
				fastCtx.Response.Header.Add(`Content-Encoding`, `br`)
			} else if hasEncoding(`deflate`) {
				fasthttp.WriteDeflate(fastCtx.Response.BodyWriter(), resp.Body())
				fastCtx.Response.Header.Add(`Content-Encoding`, `deflate`)
			} else if hasEncoding(`gzip`) {
				fasthttp.WriteGzip(fastCtx.Response.BodyWriter(), resp.Body())
				fastCtx.Response.Header.Add(`Content-Encoding`, `gzip`)
			} else {
				fastCtx.SetBody(resp.Body())
			}
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
