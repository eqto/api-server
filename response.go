package api

import (
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

type Response interface {
	ContentType() string
	StatusCode() int
	StatusMessage() string
	SetStatusCode(statusCode int)
	Header() *ResponseHeader
	Put(key string, value interface{})
	Data() json.Object
	Body() []byte
	SetContentType(contentType string)

	setBody(body []byte)
	statusMessage() *string
}

type response struct {
	statusCode int
	statusMsg  *string
	data       json.Object

	httpResp *fasthttp.Response
	err      error
	stop     bool
	writer   *streamWriter
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

func (r *response) StatusCode() int {
	return r.statusCode
}

func (r *response) StatusMessage() string {
	if r.statusMsg == nil {
		return ``
	}
	return *r.statusMsg
}

func (r *response) statusMessage() *string {
	return r.statusMsg
}

func (r *response) Put(key string, value interface{}) {
	if r.data == nil {
		r.data = json.Object{}
	}
	r.data.Put(key, value)
}

func (r *response) Data() json.Object {
	if r.data == nil {
		return nil
	}
	return r.data.Clone()
}

func (r *response) Body() []byte {
	return r.httpResp.Body()
}

func (r *response) streamWriter() Writer {
	if r.writer == nil {
		sw := &streamWriter{}
		r.httpResp.SetBodyStreamWriter(sw.write)
		r.writer = sw
	}
	return r.writer
}

func (r *response) setBody(body []byte) {
	r.httpResp.SetBody(body)
}

func (r *response) put(key string, value interface{}) {
	if r.data == nil {
		r.data = json.Object{}
	}
	r.data.Put(key, value)
}
