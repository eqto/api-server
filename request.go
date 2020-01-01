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

func parseRequest(s *Server, r *http.Request, tx *db.Tx) *Request {
	req := &Request{server: s, httpReq: r, tx: tx}
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
