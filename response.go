package api

import (
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

type Response interface {
	ContentType() string
	SetContentType(contentType string)
	SetStatusCode(statusCode int)
	Header() *ResponseHeader
	Body() []byte
	SetBody(body []byte)
}

type response struct {
	statusCode int
	statusMsg  string
	data       json.Object

	httpResp fasthttp.Response
	body     []byte
	err      error
	stop     bool
}

func (r *response) Header() *ResponseHeader {
	return &ResponseHeader{&r.httpResp.Header}
}

func (r *response) ContentType() string {
	return string(r.Header().Get(`Content-Type`))
}

func (r *response) SetContentType(contentType string) {
	r.Header().Set(`Content-Type`, contentType)
}

func (r *response) SetStatusCode(statusCode int) {
	r.statusCode = statusCode
}

func (r *response) Body() []byte {
	if r.data != nil {
		js := r.data.Clone()
		js.Put(`status`, r.statusCode).Put(`message`, r.statusMsg)
		return js.ToBytes()
	}
	return nil
}

func (r *response) SetBody(body []byte) {
	r.body = body
}
