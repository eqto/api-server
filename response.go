package apims

import "github.com/eqto/go-json"

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
	return r.Object.ToBytes()
}

func newResponse(status uint16) *response {
	return &response{status: status, header: Header{}, Object: json.Object{}}
}
