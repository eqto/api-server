package api

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

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
	middleware         []Middleware
	finalHandler       []func(ctx Context)

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

//AddFinalHandler ..
func (s *Server) AddFinalHandler(f func(ctx Context)) {
	s.finalHandler = append(s.finalHandler, f)
}

//AddRouteAuthenticator ..
func (s *Server) AddRouteAuthenticator(a RouteAuthenticator) {
	s.routeAuthenticator = append(s.routeAuthenticator, a)
}

//Proxy ...
func (s *Server) Proxy(path, dest string) error {
	if strings.HasPrefix(dest, `http://`) {
		dest = dest[7:]
	}
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
	name = filepath.Base(name)
	if strings.Count(name, `.`) > 1 {
		s.debug(`unsupported add inline function`, name)
		return nil, errors.New(`unsupported add inline function`)
	}
	name = strings.ReplaceAll(name, `.`, `/`)
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

func (s *Server) execute(fastCtx *fasthttp.RequestCtx, ctx *context) error {
	path := ctx.req.url.Path
	s.logD(`Request path:`, path)
	if e := ctx.begin(); e != nil {
		ctx.resp.setError(StatusInternalServerError, e)
		return e
	}
	defer func() {
		if ctx.resp.err != nil {
			ctx.rollback()
		} else {
			ctx.commit()
		}
	}()

	for _, m := range s.middleware {
		if e := m(ctx); e != nil {
			ctx.resp.setError(StatusInternalServerError, e)
			return e
		}
	}

	if route, e := s.GetRoute(ctx.req.Method(), path); e == nil {
		if e := route.execute(s, ctx); e != nil {
			ctx.resp.setError(StatusInternalServerError, e)
			return e
		}
		return nil
	}
	for _, proxy := range s.proxies {
		if proxy.match(string(path)) {
			if e := proxy.execute(s, ctx); e != nil {
				ctx.resp.setError(StatusInternalServerError, e)
				return e
			}
			return nil
		}
	}
	for _, file := range s.files {
		if file.match(string(path)) {
			file.handler(fastCtx)
			return nil
		}
	}
	ctx.resp.setError(StatusNotFound, errors.New(`route `+path+` not found`))
	return ctx.resp.err
}

//Serve ..
func (s *Server) Serve(port int) error {
	handler := func(fastCtx *fasthttp.RequestCtx) {
		ctx, e := newContext(s, &fastCtx.Request, &fastCtx.Response, s.cn)
		if e != nil {
			s.logW(e)
			fastCtx.WriteString(e.Error())
			return
		}
		if e := s.execute(fastCtx, ctx); e != nil {
			s.logW(e)
		}

		for _, h := range s.finalHandler {
			h(ctx)
		}
		if ctx.resp.json != nil {
			ctx.resp.httpResp.Header.Set(`Content-type`, `application/json`)
			ctx.resp.json.Put(`status`, ctx.resp.httpResp.StatusCode())
			msg := `success`
			if ctx.resp.err != nil {
				msg = ctx.resp.err.Error()
			}
			ctx.resp.json.Put(`message`, msg)

			fastCtx.Write(ctx.resp.json.ToBytes())
		}
	}
	handler = fasthttp.CompressHandlerBrotliLevel(
		fasthttp.TimeoutHandler(handler, 60*time.Second, `Timeout`),
		fasthttp.CompressBrotliDefaultCompression,
		fasthttp.CompressDefaultCompression,
	)

	return fasthttp.ListenAndServe(fmt.Sprintf(`:%d`, port), handler)
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
	path = strings.ReplaceAll(path, `/_`, `/`)
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
