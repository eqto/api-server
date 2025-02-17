package api

import (
	"errors"
	"mime/multipart"
	"net/url"

	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

type Request struct {
	fastCtx *fasthttp.RequestCtx
	js      json.Object
	url     *url.URL
}

func (r *Request) Method() string {
	return string(r.fastCtx.Method())
}

func (r *Request) QueryParam(name string) string {
	return r.url.Query().Get(name)
}

func (r *Request) File(name string) (*multipart.FileHeader, error) {
	return r.fastCtx.FormFile(name)
}

func (r *Request) Form() (*multipart.Form, error) {
	return r.fastCtx.MultipartForm()
}

func (r *Request) Header() *RequestHeader {
	return &RequestHeader{&r.fastCtx.Request.Header}
}

func (r *Request) URL() *url.URL {
	if r.url == nil {
		u, e := url.Parse(string(r.fastCtx.URI().FullURI()))
		if e != nil {
			return &url.URL{}
		}
		r.url = u
	}
	return r.url
}

func (r *Request) JSON() json.Object {
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

func (r *Request) ValidJSON(names ...string) (json.Object, error) {
	js := r.JSON()
	for _, name := range names {
		if !js.Has(name) {
			return nil, errors.New(`required parameter:` + name)
		}
	}
	return js, nil
}

func (r *Request) Body() []byte {
	return r.fastCtx.Request.Body()
}

func (r *Request) get(key string) interface{} {
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
