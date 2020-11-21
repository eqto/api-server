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

//ResponseHeader ...
type ResponseHeader struct {
	httpHeader *fasthttp.ResponseHeader
}

//Add ..
func (r *ResponseHeader) Add(key, value string) {
	r.httpHeader.Add(key, value)
}

//Set ..
func (r *ResponseHeader) Set(key, value string) {
	r.httpHeader.Set(key, value)
}
