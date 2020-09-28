package api

import (
	"errors"

	log "github.com/eqto/go-logger"
)

//ResponseError used for generate response from error
func ResponseError(status int, err error) (Response, error) {
	if err == nil {
		err = errors.New(`Unknown error`)
	}
	return newResponseError(status, err)
}

func newResponseError(status int, err error) (*response, error) {
	resp := newResponse(status)
	resp.err = err
	resp.errFrame = log.Stacktrace(2)
	return resp, err
}
