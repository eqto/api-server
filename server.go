package api

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

//Server ...
type Server struct {
	serv *fasthttp.Server

	routeMap map[string]map[string]*Route
	proxies  []proxy
	files    []file

	normalize bool

	cn                 *db.Connection
	dbConnected        bool
	routeAuthenticator []RouteAuthenticator
	middlewares        []*middlewareContainer
	finalHandler       []func(ctx Context)

	logger *logger

	stdGroup *Group
}

//Database ...
func (s *Server) Database() *db.Connection {
	return s.cn
}

//OpenDatabase call SetDatabase and Connect
func (s *Server) OpenDatabase(driver, host string, port int, username, password, name string) error {
	s.SetDatabase(driver, host, port, username, password, name)
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
func (s *Server) SetDatabase(driver, host string, port int, username, password, name string) {
	s.cn, _ = db.NewConnection(driver, host, port, username, password, name)
}

//AddMiddleware ..
func (s *Server) AddMiddleware(f func(Context) error) Middleware {
	return s.defGroup().AddMiddleware(f)
}

//AddFinalHandler ..
func (s *Server) AddFinalHandler(f func(Context)) {
	s.finalHandler = append(s.finalHandler, f)
}

//AddRouteAuthenticator ..
func (s *Server) AddRouteAuthenticator(a RouteAuthenticator) {
	s.routeAuthenticator = append(s.routeAuthenticator, a)
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
		path = normalizePath(path)
	}
	s.routeMap[method][path] = route
	s.logger.D(fmt.Sprintf(`add route %s %s`, method, path))
}

//Func add route with single func action. When secure is true, this route will validated using auth middlewares if any.
func (s *Server) Func(f func(Context) (interface{}, error)) (*Route, error) {
	return s.defGroup().Func(f)
}

//FuncSecure ..
func (s *Server) FuncSecure(f func(Context) (interface{}, error)) (*Route, error) {
	return s.defGroup().FuncSecure(f)
}

//Post ..
func (s *Server) Post(path string, f func(Context) (interface{}, error)) *Route {
	return s.defGroup().Post(path, f)
}

//PostSecure ..
func (s *Server) PostSecure(path string, f func(Context) (interface{}, error)) *Route {
	return s.defGroup().PostSecure(path, f)
}

//Query add route with single query action. When secure is true, this route will validated using auth middlewares if any.
func (s *Server) Query(path, query, params string) (*Route, error) {
	return s.defGroup().Query(path, query, params)
}

//QuerySecure ..
func (s *Server) QuerySecure(path, query, params string) (*Route, error) {
	return s.defGroup().QuerySecure(path, query, params)
}

//Get ..
func (s *Server) Get(path string, f func(Context) (interface{}, error)) *Route {
	return s.defGroup().Get(path, f)
}

//GetRoute ...
func (s *Server) GetRoute(method, path string) (*Route, bool) {
	method = strings.ToUpper(method)
	if r, ok := s.routeMap[method][path]; ok {
		return r, true
	}
	return nil, false
}

//NormalizeFunc if yes from this and beyond all Func added will renamed to lowercase, separated with underscore. Ex: HelloWorld registered as hello_world
func (s *Server) NormalizeFunc(n bool) {
	s.normalize = n
}

func (s *Server) execute(ctx *context) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.resp.json = json.Object{}
			if e, ok := r.(error); ok {
				ctx.resp.setError(StatusInternalServerError, e)
			} else {
				ctx.resp.setError(StatusInternalServerError, errors.New(`unknown error`))
			}
		}
	}()
	path := ctx.req.URL().Path
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

	if route, ok := s.GetRoute(ctx.req.Method(), path); ok {
		for _, m := range s.middlewares {
			if m.group == `` || m.group == route.group {
				if !m.secure || (m.secure && route.secure) {
					if e := m.f(ctx); e != nil {
						ctx.resp.json = json.Object{}
						if m.secure {
							ctx.resp.setError(StatusUnauthorized, e)
						} else {
							ctx.resp.setError(StatusInternalServerError, e)
						}
						return e
					}
				}
			}
		}

		if e := route.execute(s, ctx); e != nil {
			ctx.resp.json = json.Object{}
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
			file.handler(ctx.fastCtx)
			return nil
		}
	}
	ctx.resp.setError(StatusNotFound, errors.New(`route `+path+` not found`))
	return ctx.resp.err
}

//Serve ..
func (s *Server) Serve(port int) error {
	handler := func(fastCtx *fasthttp.RequestCtx) {
		ctx, e := newContext(fastCtx)
		if e != nil {
			s.logger.W(e)
			fastCtx.WriteString(e.Error())
			return
		}
		ctx.logger = s.logger
		ctx.cn = s.cn
		s.execute(ctx)

		for _, h := range s.finalHandler {
			h(ctx)
		}
		if ctx.resp.json != nil {
			ctx.resp.fastResp().Header.Set(`Content-type`, `application/json`)
			ctx.resp.json.Put(`status`, ctx.resp.getStatus())
			ctx.resp.json.Put(`message`, ctx.resp.getMessage())

			fastCtx.Write(ctx.resp.json.ToBytes())
		}
	}
	if s.serv != nil {
		if e := s.Shutdown(); e != nil {
			return e
		}
	}
	s.serv = &fasthttp.Server{Handler: fasthttp.CompressHandlerBrotliLevel(
		fasthttp.TimeoutHandler(handler, 60*time.Second, `Timeout`),
		fasthttp.CompressBrotliDefaultCompression,
		fasthttp.CompressDefaultCompression,
	),
		DisableHeaderNamesNormalizing: true,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  10 * time.Second,
		MaxKeepaliveDuration:          10 * time.Second,
	}
	return s.serv.ListenAndServe(fmt.Sprintf(`:%d`, port))
}

//Shutdown ..
func (s *Server) Shutdown() error {
	if s.serv != nil {
		s.serv.DisableKeepalive = true
		if e := s.serv.Shutdown(); e != nil {
			return e
		}
	}
	s.serv = nil
	return nil
}

//SetLogger ...
func (s *Server) SetLogger(debug func(...interface{}), info func(...interface{}), warn func(...interface{}), err func(...interface{})) {
	s.logger = &logger{D: debug, I: info, W: warn, E: err}
}

//Group ..
func (s *Server) Group(name string) *Group {
	return &Group{s: s, name: name}
}

func (s *Server) defGroup() *Group {
	if s.stdGroup == nil {
		s.stdGroup = s.Group(``)
	}
	return s.stdGroup
}

//New ...
func New() *Server {
	s := &Server{
		routeMap: make(map[string]map[string]*Route),
	}
	s.SetLogger(log.Println, log.Println, log.Println, log.Println)

	s.routeMap[MethodGet] = make(map[string]*Route)
	s.routeMap[MethodPost] = make(map[string]*Route)
	return s
}
