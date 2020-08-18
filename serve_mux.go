package api

import (
	"net/http"
	"strings"

	"gitlab.com/tuxer/go-db"

	"github.com/pkg/errors"

	"gitlab.com/tuxer/go-json"
)

//Middleware ...
type Middleware func(ctx Context)

//ServeMux ...
type ServeMux struct {
	http.Handler
	server *Server

	middlewares []Middleware

	module    map[string]Module
	stdModule Module
}

//AddMiddleware ...
func (s *ServeMux) AddMiddleware(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

func (s *ServeMux) getRoutePath(module, method, path string) *RoutePath {
	if module == `` {
		if s.stdModule == nil {
			s.stdModule = make(Module)
		}
		return s.stdModule.Get(method, path)
	}
	if s.module == nil {
		s.module = make(map[string]Module)
	}
	if s.module[module] == nil {
		s.module[module] = make(Module)
	}
	if m, ok := s.module[module]; ok {
		return m.Get(method, path)
	}
	return nil
}

func (s *ServeMux) setRoutePath(module, method, path string, routePath *RoutePath) {
	if module == `` {
		if s.stdModule == nil {
			s.stdModule = make(Module)
		}
		s.stdModule.Set(method, path, routePath)
	} else {
		if s.module == nil {
			s.module = make(map[string]Module)
		}
		if s.module[module] == nil {
			s.module[module] = make(Module)
		}
		s.module[module].Set(method, path, routePath)
	}
}

func (s *ServeMux) cn() *db.Connection {
	return s.server.cn
}

func (s *ServeMux) logger() Logger {
	return s.server.logger
}

func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &Response{Object: json.Object{`status`: 0, `message`: `success`}}
	var route *Route

	defer func() {
		if r := recover(); r != nil {
			resp.Put(`status`, 99).Put(`message`, `Error`)
			switch r := r.(type) {
			case error:
				resp.Put(`message`, r.Error())
				if e := errors.Cause(r); e != nil && e != r {
					resp.Put(`message`, r.Error()[:len(r.Error())-len(e.Error())-2])
					s.logger().W(resp.GetString(`message`))
					s.logger().E(e)
				}
			case string:
				resp.Put(`message`, r)
				s.logger().E(r)
			}
		}
		w.Header().Set(`Content-Type`, `application/json`)
		if resp.header != nil {
			for key := range resp.header {
				w.Header().Set(key, resp.header.GetString(key))
			}
		}
		if len(resp.cookies) > 0 {
			for _, cookie := range resp.cookies {
				http.SetCookie(w, &cookie)
			}
		}
		w.Write(resp.ToBytes())

	}()

	tx, e := s.cn().Begin()
	if e != nil {
		panic(wrapError(e, ErrDatabaseConnection))
	}
	defer tx.MustRecover()
	req := parseRequest(s.server, r, tx)

	ctx := Context{
		s: s.server, req: req, resp: resp, tx: tx,
	}

	if s.middlewares != nil {
		for _, middleware := range s.middlewares {
			middleware(ctx)
		}
	}

	if strings.HasPrefix(req.Path(), `/`) {
		routePath := s.getRoutePath(req.GetString(`module`), req.Method(), req.Path())
		if routePath == nil {
			panic(ErrResourceNotFound)
		}

		for _, route = range routePath.Routes() {
			if route.authType != `` {
				if a := s.server.authManager.Get(route.authType); a != nil {
					if e := a.Authenticate(tx, r); e != nil {
						panic(wrapError(e, ErrAuthentication))
					}
				}
			}

			if route.routeFunc != nil {
				if output, e := route.routeFunc(ctx); e == nil {
					ctx.put(route.output, output)
				} else {
					panic(wrapError(e, ErrRouting))
				}
			} else {
				if output, e := route.process(ctx); e == nil {
					ctx.put(route.output, output)
				} else {
					panic(wrapError(e, ErrRouting))
				}
			}
		}
	} else {

	}
}

func newServeMux(s *Server) *ServeMux {
	return &ServeMux{server: s}
}
