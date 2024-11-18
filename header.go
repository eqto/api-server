package api

import (
	"time"

	"github.com/valyala/fasthttp"
)

type RequestHeader struct {
	httpHeader *fasthttp.RequestHeader
}

func (r *RequestHeader) Get(key string) string {
	if val := r.httpHeader.Peek(key); val != nil {
		return string(val)
	}
	return ``
}

func (r *RequestHeader) Cookie(key string) string {
	cookie := r.httpHeader.Cookie(key)
	if cookie == nil {
		return ``
	}
	return string(cookie)
}

func (r *RequestHeader) Bytes() []byte {
	return r.httpHeader.Header()
}

type ResponseHeader struct {
	httpHeader *fasthttp.ResponseHeader
}

func (r *ResponseHeader) SetCookie(key, value string, expireIn time.Duration) {
	ck := &fasthttp.Cookie{}
	ck.SetKey(key)
	ck.SetValue(value)

	if expireIn > 0 {
		sec := int(expireIn.Seconds())
		if sec > 0 {
			ck.SetMaxAge(sec)
		}
	}
	ck.SetHTTPOnly(true)
	ck.SetPath(`/`)
	r.httpHeader.SetCookie(ck)
}

func (r *ResponseHeader) Add(key, value string) {
	r.httpHeader.Add(key, value)
}

func (r *ResponseHeader) Set(key, value string) {
	r.httpHeader.Set(key, value)
}

func (r *ResponseHeader) Del(key string) {
	r.httpHeader.Del(key)
}

func (r *ResponseHeader) Get(key string) string {
	if val := r.httpHeader.Peek(key); val != nil {
		return string(val)
	}
	return ``

}
