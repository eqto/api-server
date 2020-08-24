package apims

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

	header Header
	server *Server

	status   uint16
	err      error
	errFrame []log.Frame
}

func (r *response) HTTPStatus() int {
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
	js := r.Object.Clone()
	if !r.server.isProduction && r.errFrame != nil {
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
func (r *response) setError(err error) {
	r.err = err
	if !r.server.isProduction {
		r.errFrame = log.Stacktrace(2)
	}
}
