package api

import (
	"net/url"

	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

type Request interface {
	Method() string
	URL() *url.URL
	JSON() json.Object
	Body() []byte
	Header() *RequestHeader
}

type request struct {
	Request
	fastCtx *fasthttp.RequestCtx
	url     *url.URL
	body    []byte
	js      json.Object
}

func (r *request) Header() *RequestHeader {
	httpHeader := &fasthttp.RequestHeader{}
	r.fastCtx.Request.Header.CopyTo(httpHeader)

	// return &RequestHeader{ fastHeader: httpHeader}
	return &RequestHeader{fastCtx: r.fastCtx}
}

func (r *request) Method() string {
	return string(r.fastCtx.Request.Header.Method())
}

func (r *request) URL() *url.URL {
	return r.url
}

func (r *request) JSON() json.Object {
	if r.js == nil {
		if r.body != nil {
			if js, e := json.Parse(r.body); e == nil {
				r.js = js
			}
		}
	}
	if r.js == nil {
		r.js = json.Object{}
	}
	return r.js.Clone()
}

func (r *request) Body() []byte {
	return r.body
}

func (r *request) get(key string) interface{} {
	js := r.JSON()
	if js.Has(key) {
		return js.Get(key)
	}
	query := r.url.Query()
	if _, ok := query[key]; ok {
		return query.Get(key)
	}
	return nil
}
