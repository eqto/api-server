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
	SetStatus(status int, msg string)
	SetStatusCode(status int)
	SetStatusMessage(msg string)
}

type response struct {
	Response
	httpResp *fasthttp.Response
	json     json.Object
	err      error
	errFrame []log.Frame
	status   int
	message  string
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

func (r *response) SetStatus(status int, msg string) {
	r.SetStatusCode(status)
	r.SetStatusMessage(msg)
}

func (r *response) SetStatusCode(status int) {
	r.status = status
}

func (r *response) SetStatusMessage(msg string) {
	r.message = msg
}

func (r *response) getStatus() int {
	if r.err != nil {
		if r.status > 0 {
			return r.status
		} else if code := r.httpResp.StatusCode(); code != 200 {
			return code
		} else {
			return 999
		}
	}
	return r.status
}

func (r *response) getMessage() string {
	if r.message != `` {
		return r.message
	}
	if r.err != nil {
		return r.err.Error()
	}
	return ``
}

func (r *response) JSON() *json.Object {
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
