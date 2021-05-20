package api

import (
	"net/url"
	"strings"
	"sync"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

//Context ..
type Context interface {
	Session() Session
	URL() url.URL
	Method() string
	ContentType() string
	Request() Request
	Response() Response
	Tx() (*db.Tx, error)
	SetValue(name string, value interface{})
	GetValue(name string) interface{}
	Redirect(url string) error
}

type context struct {
	Context

	fastCtx *fasthttp.RequestCtx
	s       *Server

	req  request
	resp response
	sess *session

	vars json.Object

	tx     *db.Tx
	lockCn sync.Mutex

	values map[string]interface{}
}

func (c *context) Redirect(url string) error {
	return nil
}

func (c *context) Tx() (*db.Tx, error) {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	cn := c.s.Database()
	if cn != nil && c.tx == nil {
		tx, e := cn.Begin()
		if e != nil { //db error
			c.resp.setError(StatusInternalServerError, e)
			return nil, e
		}
		c.tx = tx
	}
	return c.tx, nil
}

func (c *context) URL() url.URL {
	return c.req.URL()
}

func (c *context) Method() string {
	return string(c.fastCtx.Method())
}

func (c *context) ContentType() string {
	return c.req.Header().Get(`Content-Type`)
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

func (c *context) SetValue(name string, value interface{}) {
	c.values[name] = value
}
func (c *context) GetValue(name string) interface{} {
	return c.values[name]
}

func (c *context) closeTx() {
	if c.tx == nil {
		return
	}
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.resp.err != nil {
		c.tx.Rollback()
	} else {
		c.tx.Commit()
	}
	c.tx = nil
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

func (c *context) logger() *logger {
	return c.s.logger
}

func newContext(s *Server, fastCtx *fasthttp.RequestCtx) (*context, error) {
	ctx := &context{
		s:       s,
		fastCtx: fastCtx,
		req:     request{fastCtx: fastCtx},
		resp:    response{fastCtx: fastCtx},
		values:  make(map[string]interface{}),
		sess:    &session{},
		lockCn:  sync.Mutex{},
	}

	return ctx, nil
}
