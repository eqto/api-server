package apims

import (
	uri "net/url"

	"github.com/eqto/go-json"
)

//Request ...
type Request interface {
	GetJSONBody() json.Object
}

type request struct {
	body     []byte
	url      *uri.URL
	jsonBody json.Object
}

func (r *request) GetJSONBody() json.Object {
	if r.jsonBody != nil {
		return r.jsonBody
	}
	return json.Object{}
}

func (r *request) URL() *uri.URL {
	return r.url
}

func parseRequest(url, contentType, body []byte) (*request, error) {
	u, e := uri.Parse(string(url))
	if e != nil {
		return nil, e
	}
	req := &request{url: u, body: body}
	if string(contentType) == `application/json` {
		js, e := json.Parse(body)
		if e != nil {
			return nil, e
		}
		req.jsonBody = js
	}
	return req, nil
}
