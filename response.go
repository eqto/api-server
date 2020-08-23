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

	status uint16
	header Header
	server *Server

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
	js := r.Object.Clone()
	if !r.server.isProduction && r.err != nil {
		trace := []string{}
		for _, frame := range r.errFrame {
			trace = append(trace, frame.String())
		}
		js.Put(`debug.message`, r.err.Error())
		js.Put(`debug.stacktrace`, trace)
	}
	return js.ToBytes()
}

func (r *response) setError(err error) {
	r.err = err
	r.errFrame = log.Stacktrace(2)
}
