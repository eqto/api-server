package api

import log "github.com/eqto/go-logger"

//ResponseError used to generate response from error
func ResponseError(status int, err error) (Response, error) {
	return newErrorResponse(status, err)
}

func newErrorResponse(status int, err error) (*response, error) {
	resp := newResponse(status)
	resp.err = err
	resp.errFrame = log.Stacktrace(2)
	return resp, err
}
