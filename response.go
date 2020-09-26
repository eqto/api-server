package api

import (
	"github.com/eqto/go-json"
	log "github.com/eqto/go-logger"
)

//Response ...
type Response interface {
	Header() Header
	Status() int
	Body() []byte
}

type response struct {
	json.Object
	Response

	header  Header
	rawBody []byte //if not json put here

	status   uint16
	err      error
	errFrame []log.Frame
}

func (r *response) Status() int {
	return int(r.status)
}

func (r *response) Header() Header {
	r.header.Set(`Content-Type`, `application/json`)
	return r.header.Clone()
}

func (r *response) Success() bool {
	return r.status == StatusOK
}

func (r *response) Body() []byte {
	if r.rawBody != nil {
		return r.rawBody
	}
	js := r.Object.Clone()
	if r.errFrame != nil {
		trace := []string{}
		for _, frame := range r.errFrame {
			trace = append(trace, frame.String())
		}
		js.Put(`stacktrace`, trace)
	}
	js.Put(`status`, 0).Put(`message`, `success`)
	if r.status != 200 {
		js.Put(`status`, r.status)
		js.Put(`message`, r.err.Error())
	}
	return js.ToBytes()
}

func newResponse(status uint16) *response {
	return &response{status: status, header: Header{}, Object: json.Object{}}
}
