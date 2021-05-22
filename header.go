package api

import (
	"github.com/valyala/fasthttp"
)

//RequestHeader ...
type RequestHeader struct {
	httpHeader *fasthttp.RequestHeader
}

//Get ..
func (r *RequestHeader) Get(key string) string {
	if val := r.httpHeader.Peek(key); val != nil {
		return string(val)
	}
	return ``
}

func (r *RequestHeader) Bytes() []byte {
	return r.httpHeader.Header()
}

//ResponseHeader ...
type ResponseHeader struct {
	httpHeader *fasthttp.ResponseHeader
}

func (r *ResponseHeader) SetCookie(key, value string, expire int) {
	ck := &fasthttp.Cookie{}
	ck.SetKey(key)
	ck.SetValue(value)
	ck.SetMaxAge(expire)
	ck.SetHTTPOnly(true)
	ck.SetPath(`/`)
	r.httpHeader.SetCookie(ck)
}

//Add ..
func (r *ResponseHeader) Add(key, value string) {
	r.httpHeader.Add(key, value)
}

//Set ..
func (r *ResponseHeader) Set(key, value string) {
	r.httpHeader.Set(key, value)
}

//Get ..
func (r *ResponseHeader) Get(key string) string {
	if val := r.httpHeader.Peek(key); val != nil {
		return string(val)
	}
	return ``

}
