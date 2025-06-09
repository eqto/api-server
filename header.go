package api

import (
	"github.com/valyala/fasthttp"
)

type RequestHeader struct {
	fastCtx *fasthttp.RequestCtx
}

func (r *RequestHeader) header() *fasthttp.RequestHeader {
	if r.fastCtx != nil {
		return &r.fastCtx.Request.Header
	}
	return &fasthttp.RequestHeader{}
}

func (r *RequestHeader) Get(key string) string {
	if val := r.header().Peek(key); val != nil {
		return string(val)
	}
	return ``
}

func (r *RequestHeader) GetOrNil(key string) *string {
	if val := r.header().Peek(key); val != nil {
		str := string(val)
		return &str
	}
	return nil
}

func (r *RequestHeader) RemoteIP() string {
	if val := r.fastCtx.RemoteIP(); val != nil {
		return string(val)
	}
	return ``
}

type ResponseHeader struct {
	httpHeader *fasthttp.ResponseHeader
}

func (r *ResponseHeader) Add(key, value string) {
	r.httpHeader.Add(key, value)
}

func (r *ResponseHeader) Set(key, value string) {
	r.httpHeader.Set(key, value)
}
