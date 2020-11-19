package api

import (
	"errors"

	"github.com/eqto/go-json"
	log "github.com/eqto/go-logger"
	"github.com/valyala/fasthttp"
)

//Response ..
type Response interface {
	Body() []byte
	SetBody(body []byte)
	ContentType() string
}

type response struct {
	Response
	httpResp *fasthttp.Response
	json     json.Object
	err      error
	errFrame []log.Frame
}

func (r *response) Body() []byte {
	return r.httpResp.Body()
}

func (r *response) SetBody(body []byte) {
	r.httpResp.SetBody(body)
}

func (r *response) ContentType() string {
	return string(r.httpResp.Header.ContentType())
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
