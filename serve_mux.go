package api

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/tuxer/go-json"
)

//ServeMux ...
type ServeMux struct {
	http.Handler
	s *Server
}

//Serve ...
func (m *ServeMux) Serve(r *http.Request) (resp *Response, err error) {
	resp = &Response{Object: json.Object{`status`: 0, `message`: `success`}}
	var route *Route

	defer func() {
		if r := recover(); r != nil {
			resp.Put(`status`, 99).Put(`message`, `Error`)
			switch r := r.(type) {
			case error:
				resp.Put(`message`, r.Error())
				if e := errors.Cause(r); e != nil && e != r {
					resp.Put(`message`, r.Error()[:len(r.Error())-len(e.Error())-2])
					m.s.logger.W(resp.GetString(`message`))
					m.s.logger.E(e)
				}
			case string:
				resp.Put(`message`, r)
				err = errors.New(r)
			}
		}
	}()

	path := r.URL.Path
	if strings.HasPrefix(path, `/`) {
		routePath, ok := m.s.routePathMap[r.Method+` `+r.URL.Path]
		if !ok {
			panic(ErrResourceNotFound)
		}

		tx, e := m.s.cn.Begin()
		if e != nil {
			panic(wrapError(e, ErrDatabaseConnection))
		}
		defer tx.MustRecover()

		req := parseRequest(m.s, r, tx)

		ctx := &Context{
			s: m.s, req: req, resp: resp, tx: tx,
		}

		for _, route = range routePath.Routes() {
			if route.authType != `` {
				if a := m.s.authManager.Get(route.authType); a != nil {
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
	return
}

func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, e := m.Serve(r)
	if e != nil {
		m.s.logger.E(e)
	}
	w.Header().Set(`Content-Type`, `application/json`)
	w.Write(resp.ToBytes())
}

func newServeMux(s *Server) *ServeMux {
	return &ServeMux{s: s}
}
