package apims

//Response ...
type Response interface {
	Header() Header
	Status() int
	Body() []byte
}

type response struct {
	Response

	status uint16
	header Header
	body   []byte
}

func (r *response) Status() int {
	return int(r.status)
}

func (r *response) Body() []byte {
	return r.body
}

func (r *response) Header() Header {
	return r.header.Clone()
}

func (r *response) Success() bool {
	return r.status == StatusOK
}

func newResponse(status uint16) *response {
	return &response{status: status, header: Header{}, body: []byte{}}
}
