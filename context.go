package api

import (
	"net/url"
	"strings"
	"sync"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

type Context interface {
	Session() Session
	Request() Request
	Response() Response
	Tx() *db.Tx
	SetValue(name string, value interface{})
	GetValue(name string) interface{}
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

	next   bool
	values map[string]interface{}
}

func (c *context) Session() Session {
	return c.sess
}

func (c *context) Request() Request {
	return &c.req
}

func (c *context) Response() Response {
	return &c.resp
}

func (c *context) SetValue(name string, value interface{}) {
	c.values[name] = value
}
func (c *context) GetValue(name string) interface{} {
	return c.values[name]
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
		c.resp.JSON().Put(property, value)
	}
}

func newContext(s *Server, fastCtx *fasthttp.RequestCtx, cn *db.Connection) (*context, error) {
	httpReq := &fastCtx.Request
	ctx := &context{
		s:      s,
		req:    request{fastCtx: fastCtx},
		resp:   response{fastResp: &fastCtx.Response},
		cn:     cn,
		values: make(map[string]interface{}),
		sess:   &session{},
		lockCn: sync.Mutex{},
	}
	url, e := url.Parse(string(httpReq.URI().FullURI()))
	if url == nil {
		return nil, errors.Wrap(e, `invalid url `+string(httpReq.RequestURI()))
	}
	ctx.req.url = url
	ctx.req.body = httpReq.Body()

	return ctx, nil
}
