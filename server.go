package api

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/eqto/dbm"
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

// Server ...
type Server struct {
	serv *fasthttp.Server

	routeMap map[string]map[string]*Route
	proxies  []proxy
	files    []file

	normalize bool

	cn          *dbm.Connection
	dbConnected bool
	middlewares []*middlewareContainer
	render      Render

	logger *logger

	stdGroup       *Group
	maxRequestSize int
}

// Database ...
func (s *Server) Database() *dbm.Connection {
	return s.cn
}

func (s *Server) SetDatabase(cn *dbm.Connection) {
	s.cn = cn
}

func (s *Server) MaxRequestSize(size int) {
	s.maxRequestSize = size
}

// OpenDatabase ..
func (s *Server) OpenDatabase(driver, host string, port int, username, password, name string) error {
	cn, e := dbm.Connect(driver, host, port,
		username, password, name)
	if e != nil {
		return e
	}
	s.cn = cn
	s.dbConnected = true
	return nil
}

// Connect ...
func (s *Server) Connect() error {
	if e := s.cn.Connect(); e != nil {
		return e
	}
	s.dbConnected = true
	return nil
}

// AddMiddleware ..
func (s *Server) AddMiddleware(f func(Context) error) Middleware {
	return s.defGroup().AddMiddleware(f)
}

func (s *Server) SetRender(r Render) {
	s.render = r
}

// Proxy ...
func (s *Server) Proxy(path, dest string) error {
	p, e := newProxy(path, dest)
	if e != nil {
		return e
	}
	s.proxies = append(s.proxies, p)
	return nil
}

// FileRoute serve static file. Path parameter to determine url to be processed. Dest parameter will find directory of the file reside. RedirectTo parameter to redirect non existing file, this param can be used for SPA (ex. index.html).
func (s *Server) FileRoute(path, dest, redirectTo string) error {
	f, e := newFile(path, dest, redirectTo)
	if e != nil {
		return e
	}
	s.files = append(s.files, f)
	return nil
}

// FileRouteRemove ...
func (s *Server) FileRouteRemove(path string) error {
	for i, file := range s.files {
		if file.path == path {
			s.files = append(s.files[:i], s.files[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *Server) SetPrefixPath(path string) {
	if path == `` {
		return
	}
	if path[0] != '/' {
		path = `/` + path
	}
	s.defGroup().prefixPath = path
}

func (s *Server) Post(path string) *Route {
	return s.defGroup().Post(path)
}

func (s *Server) PostAction(f func(Context) error) *Route {
	return s.defGroup().PostAction(f)
}

func (s *Server) PostSecureAction(f func(Context) error) *Route {
	return s.defGroup().PostSecureAction(f)
}

func (s *Server) Get(path string) *Route {
	return s.defGroup().Get(path)
}

func (s *Server) GetAction(f func(Context) error) *Route {
	return s.defGroup().GetAction(f)
}

// NormalizeFunc this func need to called before adding any routes. parameter n=true for renaming all route paths to lowercase, separated with underscore. Ex: /HelloWorld registered as /hello_world
func (s *Server) NormalizeFunc(n bool) {
	s.normalize = n
}

func (s *Server) executeRoutes(ctx *context, path string) bool {
	currPath := path
	var route *Route
	for {
		r, ok := s.routeMap[ctx.Method()][currPath]
		if !ok {
			split := strings.Split(currPath, `/`)
			if len(split) > 1 {
				currPath = strings.Join(split[0:len(split)-1], `/`)
			} else {
				return false
			}
		} else {
			route = r
			break
		}
	}
	if route == nil {
		return false
	}
	for _, m := range s.middlewares {
		if m.group == `` || m.group == route.group {
			if !m.secure || (m.secure && route.secure) {
				if e := m.f(ctx); e != nil {
					ctx.setErr(e)
					if m.secure {
						ctx.StatusUnauthorized(`Authorization error: ` + e.Error())
					} else {
						s.logger.W(e)
						ctx.StatusInternalServerError(`Internal server error`)
					}
					return true
				}
			}
		}
	}
	httpResp := ctx.resp.httpResp

	e := route.execute(s, ctx)

	if e != nil {
		ctx.setErr(e)
	} else if !httpResp.IsBodyStream() && ctx.resp.data == nil && len(httpResp.Body()) == 0 {
		ctx.resp.data = json.Object{}
	}
	ctx.closeTx()
	return true
}

func (s *Server) executeProxies(ctx *context, fastCtx *fasthttp.RequestCtx, path string) bool {
	for _, proxy := range s.proxies {
		if proxy.match(string(path)) {
			if e := proxy.execute(s, fastCtx); e != nil {
				ctx.setErr(e)
				ctx.StatusInternalServerError(`Unable to execute proxy`)
			}
			return true
		}
	}
	return false
}

func (s *Server) executeFiles(fastCtx *fasthttp.RequestCtx, path string) bool {
	for _, file := range s.files {
		if file.match(string(path)) {
			file.handler(fastCtx)
			return true
		}
	}
	return false
}

func (s *Server) serve(ln net.Listener) error {
	handler := func(fastCtx *fasthttp.RequestCtx) {
		ctx, e := newContext(s, fastCtx)
		if e != nil {
			s.logger.W(e)
			fastCtx.WriteString(e.Error())
			return
		}

		path := ctx.URL().Path

		ctx.resp.SetContentType(`application/json`)

		if ok := s.executeRoutes(ctx, path); !ok {
			if ok := s.executeProxies(ctx, fastCtx, path); !ok {
				if ok := s.executeFiles(fastCtx, path); !ok {
					ctx.setErr(errors.New(`route ` + path + ` not found`))
					ctx.StatusServiceUnavailable(`route ` + path + ` not found`)
				}
			}
		}
		renderOk := false
		if s.render != nil {
			renderOk = s.render(ctx)
		}
		if !renderOk {
			render(ctx)
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

		CloseOnShutdown:               true,
		NoDefaultServerHeader:         true,
		NoDefaultContentType:          true,
		NoDefaultDate:                 true,
		DisableHeaderNamesNormalizing: true,
	}
	if s.maxRequestSize > 0 {
		s.serv.MaxRequestBodySize = s.maxRequestSize
	}
	return s.serv.Serve(ln)
}

func (s *Server) ServeUnix(filename string) error {
	ln, e := net.Listen(`unix`, filename)
	if e != nil {
		return e
	}
	defer ln.Close()

	return s.serve(ln)
}

// Serve ..
func (s *Server) Serve(port int) error {
	ln, e := net.Listen(`tcp4`, fmt.Sprintf(`:%d`, port))
	if e != nil {
		return e
	}
	defer ln.Close()

	return s.serve(ln)
}

// Shutdown ..
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

// SetLogger ...
func (s *Server) SetLogger(debug func(...interface{}), info func(...interface{}), warn func(...interface{}), err func(...interface{})) {
	s.logger = &logger{D: debug, I: info, W: warn, E: err}
}

// Group ..
func (s *Server) Group(name string) *Group {
	return &Group{s: s, name: name}
}

func (s *Server) defGroup() *Group {
	if s.stdGroup == nil {
		s.stdGroup = s.Group(``)
	}
	return s.stdGroup
}

// New ...
func New() *Server {
	s := &Server{
		routeMap: make(map[string]map[string]*Route),
	}
	s.SetLogger(log.Println, log.Println, log.Println, log.Println)

	s.routeMap[MethodGet] = make(map[string]*Route)
	s.routeMap[MethodPost] = make(map[string]*Route)
	return s
}
