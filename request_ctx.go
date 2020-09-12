package api

import (
	"net/url"

	"github.com/eqto/go-db"
)

//RequestCtx ..
type RequestCtx interface {
	Header() Header
	URL() url.URL
	Session() Session
	Tx() *db.Tx
}

type requestCtx struct {
	RequestCtx

	req  *request
	sess *session
	tx   *db.Tx
}

func (r *requestCtx) Header() Header {
	return r.req.header
}

func (r *requestCtx) URL() url.URL {
	return r.req.url
}

func (r *requestCtx) Tx() *db.Tx {
	return r.tx
}

func (r *requestCtx) Session() Session {
	return r.sess
}

func newRequestCtx(req *request, sess *session) *requestCtx {
	return &requestCtx{req: req, sess: sess}
}
