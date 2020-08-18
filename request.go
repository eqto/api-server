package api

import (
	"io/ioutil"
	"net/http"

	"gitlab.com/tuxer/go-db"

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
	tx      *db.Tx

	server *Server
}

//MustString ...
func (r Request) MustString(key string) string {
	val := r.GetStringNull(key)
	if val == nil {
		panic(`required parameter: ` + key)
	}
	return *val
}

//MustInt ...
func (r Request) MustInt(key string) int {
	val := r.GetIntNull(key)
	if val == nil {
		panic(`required parameter: ` + key)
	}
	return *val
}

//Method ...
func (r Request) Method() string {
	return r.httpReq.Method
}

//Path ...
func (r Request) Path() string {
	return r.httpReq.URL.Path
}

func parseRequest(s *Server, r *http.Request, tx *db.Tx) *Request {
	req := &Request{server: s, httpReq: r, tx: tx}

	js := make(json.Object)
	for key := range r.URL.Query() {
		js.Put(key, r.URL.Query().Get(key))
	}
	req.Object = js

	if r.Method == http.MethodPost {
		if body, e := ioutil.ReadAll(r.Body); e == nil {
			js = json.Parse(body)
			for key, val := range js {
				req.Object.Put(key, val)
			}
		}
	}

	return req
}
