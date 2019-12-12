package api

import (
	"net/http"
)

//Context ...
type Context interface {
	Request() *http.Request
}

type context struct {
	Context

	r *http.Request
}

func (c *context) Request() *http.Request {
	return c.r
}
