package api

import (
	log "github.com/eqto/go-logger"
	"github.com/valyala/fasthttp"
)

//Response ..
type Response interface {
	Body() []byte
	SetBody(body []byte)
}

type response struct {
	Response
	httpResp *fasthttp.Response
	err      error
	errFrame []log.Frame
}

func (r response) Body() []byte {
	return r.httpResp.Body()
}

func (r response) SetBody(body []byte) {
	r.httpResp.SetBody(body)
}

func (r response) setError(status int, e error) {
	if r.err == nil {
		r.httpResp.SetStatusCode(status)
		r.err = e
		r.errFrame = log.Stacktrace(2)
	}
}
