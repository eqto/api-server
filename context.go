package api

import (
	"bytes"
	"net/url"
	"strings"
	"sync"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

//Context ..
type Context interface {
	Session() Session
	Request() Request
	Response() Response
	Tx() *db.Tx
}

type context struct {
	Context
	s    *Server
	req  request
	resp response
	sess *session

	vars json.Object

	cn     *db.Connection
	tx     *db.Tx
	lockCn sync.Mutex

	next bool
}

//Session ..
func (c *context) Session() Session {
	return c.sess
}

func (c *context) Request() Request {
	return &c.req
}

func (c *context) Response() Response {
	return &c.resp
}

func (c *context) Tx() *db.Tx {
	return c.tx
}

func (c *context) begin() error {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.cn != nil {
		tx, e := c.cn.Begin()
		if e != nil { //db error
			return e
		}
		c.tx = tx
	}
	return nil
}
func (c *context) rollback() {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.tx != nil {
		c.tx.Rollback()
		c.tx = nil
	}
}
func (c *context) commit() {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.tx != nil {
		c.tx.Commit()
	}
}

func (c *context) put(property string, value interface{}) {
	if strings.HasPrefix(property, `$`) { //save to vars
		if c.vars == nil {
			c.vars = json.Object{}
		}
		c.vars.Put(property[1:], value)
	} else { //save to result
		if c.resp.json == nil {
			c.resp.json = json.Object{}
		}
		c.resp.json.Put(property, value)
	}
}

func newContext(s *Server, req *fasthttp.Request, resp *fasthttp.Response, cn *db.Connection) (*context, error) {
	ctx := &context{
		s:      s,
		req:    request{httpReq: req},
		resp:   response{httpResp: resp},
		cn:     cn,
		sess:   &session{},
		lockCn: sync.Mutex{},
	}
	url, e := url.Parse(string(req.RequestURI()))
	if url == nil {
		return nil, errors.Wrap(e, `invalid url `+string(req.RequestURI()))
	}
	ctx.req.url = url

	if bytes.HasPrefix(req.Header.ContentType(), []byte(`application/json`)) {
		body := req.Body()
		if len(body) == 0 {
			body = []byte(`{}`)
		}
		req, e := json.Parse(body)
		if e != nil {
			return nil, errors.Wrap(e, `invalid json body`)
		}
		ctx.req.json = req
		ctx.resp.json = json.Object{}
	} else {
		ctx.req.json = json.Object{}
	}
	return ctx, nil
}
