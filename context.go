package api

import (
	"context"

	"gitlab.com/tuxer/go-db"
)

//Context ...
type Context struct {
	context.Context

	req  *Request
	resp *Response
	tx   *db.Tx
}

//Request ...
func (c *Context) Request() *Request {
	return c.req
}

//Response ...
func (c *Context) Response() *Response {
	return c.resp
}

//Tx ...
func (c *Context) Tx() *db.Tx {
	return c.tx
}
