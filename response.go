package api

import (
	"errors"

	"github.com/eqto/go-json"
	log "github.com/eqto/go-logger"
	"github.com/valyala/fasthttp"
)

//Response ..
type Response interface {
	Header() *ResponseHeader
	Body() []byte
	SetBody(body []byte)
	ContentType() string
	SetStatus(status int, message string)
}

type response struct {
	Response
	httpResp *fasthttp.Response
	json     json.Object
	err      error
	errFrame []log.Frame
}

func (r *response) Header() *ResponseHeader {
	return &ResponseHeader{&r.httpResp.Header}
}

func (r *response) Body() []byte {
	return r.httpResp.Body()
}

func (r *response) SetBody(body []byte) {
	r.httpResp.SetBody(body)
	r.json = nil
}

func (r *response) ContentType() string {
	return string(r.httpResp.Header.ContentType())
}

func (r *response) SetStatus(status int, message string) {
	if r.json == nil {
		r.json = json.Object{}
	}
	r.json.Put(`status`, status)
	r.json.Put(`message`, message)
}

func (r *response) mustJSON() *json.Object {
	if r.json == nil {
		r.json = json.Object{}
	}
	r.httpResp.ResetBody()
	return &r.json
}

func (r *response) setError(status int, e error) {
	if r.err == nil {
		r.httpResp.SetStatusCode(status)
		if e == nil {
			e = errors.New(``)
		}
		r.err = e
		r.errFrame = log.Stacktrace(2)
	}
}
