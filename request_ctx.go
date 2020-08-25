package apims

import "net/url"

//RequestCtx ..
type RequestCtx interface {
	Header() Header
	URL() url.URL
	Session() Session
}

type requestCtx struct {
	RequestCtx

	req  *request
	sess *session
}

func (r *requestCtx) Header() Header {
	return r.req.header
}

func (r *requestCtx) URL() url.URL {
	return r.req.url
}
func (r *requestCtx) Session() Session {
	return r.sess
}

func newRequestCtx(req *request, sess *session) *requestCtx {
	return &requestCtx{req: req, sess: sess}
}
