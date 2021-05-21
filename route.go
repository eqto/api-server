package api

import (
	"github.com/eqto/api-server/status"
	"github.com/valyala/fasthttp"
)

//Route ...
type Route struct {
	action []Action
	secure bool
	group  string
}

//Secure ...
func (r *Route) Secure() *Route {
	r.secure = true
	return r
}

//UseGroup only use middleware that have the same name or no name
func (r *Route) UseGroup(name string) *Route {
	r.group = name
	return r
}

//AddQueryAction ...
func (r *Route) AddQueryAction(query, params, property string) (Action, error) {
	act, e := newQueryAction(query, property, params)
	if e != nil {
		return nil, e
	}
	r.action = append(r.action, act)
	return act, nil
}

//AddFuncAction ...
func (r *Route) AddFuncAction(f func(Context) (interface{}, error), property string) Action {
	act := newFuncAction(f, property)
	r.action = append(r.action, act)
	return act
}

func (r *Route) execute(s *Server, ctx *context) error {
	if s.routeAuthenticator != nil {
		for _, m := range s.routeAuthenticator {
			if r.secure {
				if e := m(ctx); e != nil {
					ctx.resp.setError(StatusUnauthorized, e)
					return e
				}
			}
		}
	}

	for _, action := range r.action {
		if result, e := action.execute(ctx); e == nil {
			if redirect, ok := result.(status.Redirect); ok {
				// regex := regexp.MustCompile(`http(s|):|//`)
				// if url := string(redirect); regex.MatchString(url) { //absolute url
				// 	ctx.fastCtx.Redirect(url, fasthttp.StatusFound)
				// } else {

				// }
				url := string(redirect)
				ctx.fastCtx.Redirect(url, fasthttp.StatusFound)
				return nil
			} else if data, ok := result.(Data); ok {
				if data.Status > 0 {
					ctx.resp.fastResp().SetStatusCode(data.Status)
				}
				ctx.resp.fastResp().Header.Set(`Content-type`, data.ContentType)
				ctx.resp.SetBody(data.Body)
			} else {
				if prop := action.property(); prop != `` {
					ctx.put(prop, result)
				}
			}
		} else {
			ctx.resp.setError(StatusInternalServerError, e)
			return e
		}
	}
	return nil
}
