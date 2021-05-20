package api

import (
	"net/url"

	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

//Request ..
type Request interface {
	Header() *RequestHeader
	URL() *url.URL
	JSON() json.Object
	Body() []byte
}

type request struct {
	Request
	fastCtx *fasthttp.RequestCtx
	js      json.Object
	url     *url.URL
}

func (r *request) fastReq() *fasthttp.Request {
	return &r.fastCtx.Request
}

func (r *request) Header() *RequestHeader {
	header := &RequestHeader{httpHeader: &fasthttp.RequestHeader{}}
	r.fastReq().Header.CopyTo(header.httpHeader)
	return header
}

func (r *request) URL() *url.URL {
	if r.url == nil {
		url, e := url.Parse(string(r.fastReq().URI().FullURI()))
		if e != nil {
			return nil
		}
		r.url = url
	}
	return r.url
}

func (r *request) JSON() json.Object {
	if r.js == nil {
		body := r.Body()
		if body != nil {
			if js, e := json.Parse(body); e == nil {
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
	return r.fastReq().Body()
}

func (r *request) get(key string) interface{} {
	js := r.JSON()
	if js.Has(key) {
		return js.Get(key)
	}
	query := r.URL().Query()
	if _, ok := query[key]; ok {
		return query.Get(key)
	}
	return nil
}
