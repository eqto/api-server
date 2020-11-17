package api

import (
	"net/url"

	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

//Request ..
type Request interface {
	Method() string
	URL() *url.URL
}

type request struct {
	Request
	httpReq *fasthttp.Request
	url     *url.URL
	json    json.Object
}

func (r request) Method() string {
	return string(r.httpReq.Header.Method())
}

func (r request) get(key string) interface{} {
	if r.json.Has(key) {
		return r.json.Get(key)
	}
	return r.url.Query().Get(key)
}
