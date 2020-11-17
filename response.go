package api

import (
	log "github.com/eqto/go-logger"
	"github.com/valyala/fasthttp"
)

//Response ..
type Response interface {
}

type response struct {
	Response
	httpResp *fasthttp.Response
	err      error
	errFrame []log.Frame
}

func (r *response) setError(status int, e error) {
	r.httpResp.SetStatusCode(status)
	r.err = e
	r.errFrame = log.Stacktrace(2)
}
