package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/tuxer/go-json"
)

const (
	AuthNone = ``
	AuthJWT  = `jwt`
)

//Request ...
type Request struct {
	json.Object

	httpReq *http.Request

	server *Server
}

func (r Request) Path() string {
	return r.httpReq.URL.Path
}

//Authenticate ...
func (r Request) Authenticate() error {
	var e error
	if r.server.authType == `jwt` {
		e = jwtAuthorize(r.server, r.httpReq)
	}
	if e != nil {
		panic(fmt.Sprintf(`authentication for %s failed`, r.Path()))
		return e
	}
	return nil
}

//MustString ...
func (r Request) MustString(key string) string {
	val := r.GetStringNull(key)
	if val == nil {
		panic(`required parameter: ` + key)
	}
	return *val
}

func parseRequest(s *Server, r *http.Request) *Request {
	req := &Request{server: s, httpReq: r}
	switch r.Method {
	case http.MethodPost:
		if body, e := ioutil.ReadAll(r.Body); e == nil {
			req.Object = json.Parse(body)
		}
	case http.MethodGet:
		js := make(json.Object)
		for key := range r.URL.Query() {
			js.Put(key, r.URL.Query().Get(key))
		}
		req.Object = js
	}
	return req
}
