package api

import (
	"bufio"
	"bytes"
	"errors"
	"net/textproto"
	uri "net/url"

	"github.com/eqto/go-json"
)

//Request ...
type Request interface {
	Header() Header
	JSONBody() json.Object
}

type request struct {
	method   []byte
	header   Header
	body     []byte
	url      uri.URL
	jsonBody json.Object
}

func (r *request) Header() Header {
	return r.header
}

func (r *request) JSONBody() json.Object {
	if r.jsonBody != nil {
		return r.jsonBody
	}
	return json.Object{}
}

func (r *request) get(key string) interface{} {
	if r.jsonBody != nil {
		if r.jsonBody.Has(key) {
			return r.jsonBody.Get(key)
			// } else {
			// 	return nil
		}
		return nil
	}
	return r.url.Query().Get(key)
}

func (r *request) URL() uri.URL {
	return r.url
}

func parseRequest(method, url, header, body []byte) (*request, error) {
	u, e := uri.Parse(string(url))
	if e != nil {
		return nil, e
	}
	req := &request{method: method, url: *u}
	if string(method) == MethodPost {
		req.body = body
	}
	tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(header)))
	mimeReader, e := tp.ReadMIMEHeader()
	req.header = Header(mimeReader)

	if string(method) == MethodPost {
		contentType := req.header.Get(`Content-Type`)

		if contentType != `application/json` && body != nil && len(body) > 0 {
			return nil, errors.New(`POST method only support Content-Type: application/json`)
		}
		if body != nil && len(body) > 0 {
			js, e := json.Parse(body)
			if e != nil {
				return nil, e
			}
			if js == nil {
				js = json.Object{}
			}
			req.jsonBody = js
		}
	}

	return req, nil
}
