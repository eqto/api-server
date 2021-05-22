package api

import (
	"net/url"

	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

//Request ..
type Request interface {
	Header() *RequestHeader
	Method() string
	URL() *url.URL
	JSON() json.Object
	Body() []byte
}

type request struct {
	Request
	httpReq fasthttp.Request
	js      json.Object
	url     *url.URL
}

func (r *request) Method() string {
	return string(r.httpReq.Header.Method())
}

func (r *request) Header() *RequestHeader {
	return &RequestHeader{&r.httpReq.Header}
}

func (r *request) URL() *url.URL {
	if r.url == nil {
		u, e := url.Parse(string(r.httpReq.URI().FullURI()))
		if e != nil {
			return &url.URL{}
		}
		r.url = u
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
	return r.httpReq.Body()
}

func (r *request) get(key string) interface{} {
	js := r.JSON()
	if js.Has(key) {
		return js.Get(key)
	}
	u := r.URL()
	query := u.Query()
	if _, ok := query[key]; ok {
		return query.Get(key)
	}
	return nil
}
