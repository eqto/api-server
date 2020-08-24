package apims

import (
	"net/url"
	"strings"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
)

type actionCtx struct {
	tx   *db.Tx
	req  *request
	resp *response

	vars json.Object
}

func (a *actionCtx) Request() Request {
	return a.req
}
func (a *actionCtx) Response() Response {
	return a.resp
}

func (a *actionCtx) Tx() *db.Tx {
	return a.tx
}

func (a *actionCtx) put(property string, value interface{}) {
	if strings.HasPrefix(property, `$`) { //save to vars
		a.vars.Put(property[1:], value)
	} else { //save to result
		a.resp.Put(property, value)
	}
}
func (a *actionCtx) get(property string) interface{} {
	if strings.HasPrefix(property, `$`) { //get from to vars
		return a.vars.Get(property[1:])
	} else { //get from result
		return a.resp.Get(property)
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
