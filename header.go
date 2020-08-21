package apims

import (
	"net/http"
	"net/textproto"
)

//Header copy from net/http
type Header http.Header

//Add ...
func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

//Set ...
func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

//Get ...
func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Clone ...
func (h Header) Clone() Header {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(Header, len(h))
	for k, vv := range h {
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
	return h2
}
