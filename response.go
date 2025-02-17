package api

import (
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

type Response struct {
	statusCode int
	statusMsg  *string
	data       json.Object

	httpResp *fasthttp.Response
	err      error
	stop     bool
	writer   *streamWriter
}

func (r *Response) GetError() error {
	return r.err
}
func (r *Response) Header() *ResponseHeader {
	return &ResponseHeader{&r.httpResp.Header}
}

func (r *Response) ContentType() string {
	return string(r.Header().Get(`Content-Type`))
}

func (r *Response) SetContentType(contentType string) {
	r.Header().Set(`Content-Type`, contentType)
}

func (r *Response) SetStatusCode(statusCode int) {
	r.statusCode = statusCode
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) StatusMessage() string {
	if r.statusMsg == nil {
		return ``
	}
	return *r.statusMsg
}

func (r *Response) statusMessage() *string {
	return r.statusMsg
}

func (r *Response) Put(key string, value interface{}) {
	if r.data == nil {
		r.data = json.Object{}
	}
	r.data.Put(key, value)
}

func (r *Response) Data() json.Object {
	if r.data == nil {
		return nil
	}
	return r.data.Clone()
}

func (r *Response) Body() []byte {
	return r.httpResp.Body()
}

func (r *Response) streamWriter() Writer {
	if r.writer == nil {
		sw := &streamWriter{}
		r.httpResp.SetBodyStreamWriter(sw.write)
		r.writer = sw
	}
	return r.writer
}

func (r *Response) setBody(body []byte) {
	r.httpResp.SetBody(body)
}

func (r *Response) put(key string, value interface{}) {
	if r.data == nil {
		r.data = json.Object{}
	}
	r.data.Put(key, value)
}
