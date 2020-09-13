package api

import (
	"net/url"

	"github.com/eqto/go-db"
)

//RequestCtx ..
type RequestCtx interface {
	Header() Header
	Body() []byte
	URL() url.URL
	Session() Session
	Tx() *db.Tx
}

type requestCtx struct {
	RequestCtx

	req  *request
	sess *session
	cn   *db.Connection
	tx   *db.Tx
}

func (r *requestCtx) Header() Header {
	return r.req.header
}

func (r *requestCtx) Body() []byte {
	return r.req.body
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

func (r *requestCtx) begin() error {
	if r.cn != nil {
		tx, e := r.cn.Begin()
		if e != nil { //db error
			return e
		}
		r.tx = tx
	}
	return nil
}
func (r *requestCtx) rollback() {
	if r.tx != nil {
		r.tx.Rollback()
	}
}
func (r *requestCtx) commit() {
	if r.tx != nil {
		r.tx.Commit()
	}
}

func newRequestCtx(cn *db.Connection, req *request, sess *session) *requestCtx {
	return &requestCtx{cn: cn, req: req, sess: sess}
}
