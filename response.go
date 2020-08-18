package api

import (
	"net/http"
	"time"

	"gitlab.com/tuxer/go-json"
)

//Response ...
type Response struct {
	json.Object

	header json.Object

	cookies []http.Cookie
}

//SetHeader ...
func (r *Response) SetHeader(key, value string) {
	if r.header == nil {
		r.header = make(json.Object)
	}
	r.header.Put(key, value)
}

//AddCookie ...
func (r *Response) AddCookie(name, value string, expires time.Time) {
	cookie := http.Cookie{Name: name, Value: value, Expires: expires, HttpOnly: true}
	r.cookies = append(r.cookies, cookie)
}
