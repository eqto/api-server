package apims

import (
	"net/url"
	"strings"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
)

//Context ...
type Context interface {
	Request() Request
	Response() Response
	Session() Session
	Tx() *db.Tx
}

type context struct {
	tx   *db.Tx
	req  *request
	resp *response

	vars json.Object
	sess *session
}

func (c *context) Request() Request {
	return c.req
}

func (c *context) Response() Response {
	return c.resp
}

func (c *context) Session() Session {
	return c.resp
}

func (c *context) Tx() *db.Tx {
	return c.tx
}

func (c *context) put(property string, value interface{}) {
	if strings.HasPrefix(property, `$`) { //save to vars
		c.vars.Put(property[1:], value)
	} else { //save to result
		c.resp.Put(property, value)
	}
}
func (c *context) get(property string) interface{} {
	if strings.HasPrefix(property, `$`) { //get from to vars
		return c.vars.Get(property[1:])
	} else { //get from result
		return c.resp.Get(property)
	}
}

//RequestCtx ..
type RequestCtx interface {
	Header() Header
	URL() url.URL
}

type requestCtx struct {
	RequestCtx

	req *request
}

func (r *requestCtx) Header() Header {
	return r.req.header
}

func (r *requestCtx) URL() url.URL {
	return r.req.url
}

func newRequestCtx(req *request) *requestCtx {
	return &requestCtx{req: req}
}
