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
