package api

import (
	"time"

	"github.com/valyala/fasthttp"
)

// RequestHeader ...
type RequestHeader struct {
	httpHeader *fasthttp.RequestHeader
}

// Get ..
func (r *RequestHeader) Get(key string) string {
	if val := r.httpHeader.Peek(key); val != nil {
		return string(val)
	}
	return ``
}

func (r *RequestHeader) Bytes() []byte {
	return r.httpHeader.Header()
}

// ResponseHeader ...
type ResponseHeader struct {
	httpHeader *fasthttp.ResponseHeader
}

func (r *ResponseHeader) SetCookie(key, value string, expireIn time.Duration) {
	ck := &fasthttp.Cookie{}
	ck.SetKey(key)
	ck.SetValue(value)
	if expireIn > 0 {
		sec := int(expireIn / 1 * time.Second)
		if sec > 0 {
			ck.SetMaxAge(sec)
		}
	}
	ck.SetHTTPOnly(true)
	ck.SetPath(`/`)
	r.httpHeader.SetCookie(ck)
}

// Add ..
func (r *ResponseHeader) Add(key, value string) {
	r.httpHeader.Add(key, value)
}

// Set ..
func (r *ResponseHeader) Set(key, value string) {
	r.httpHeader.Set(key, value)
}

// Get ..
func (r *ResponseHeader) Get(key string) string {
	if val := r.httpHeader.Peek(key); val != nil {
		return string(val)
	}
	return ``

}
