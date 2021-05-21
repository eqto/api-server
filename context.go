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
	URL() *url.URL
	Method() string
	ContentType() string
	SetCookie(key, value string, expire int)
	Write(value interface{}) error
	Status(status int, msg string) error
	StatusNotFound(msg string) error

	Request() Request
	Response() Response
	Tx() (*db.Tx, error)
	Session() Session
	SetValue(name string, value interface{})
	GetValue(name string) interface{}
}

type context struct {
	Context

	fastCtx *fasthttp.RequestCtx
	s       *Server

	property string

	req  request
	resp response
	sess *session

	vars json.Object

	tx     *db.Tx
	lockCn sync.Mutex

	values map[string]interface{}
}

func (c *context) Write(value interface{}) error {
	if c.property != `` {
		c.put(c.property, value)
	}
	return nil
}

func (c *context) Status(status int, msg string) error {
	c.resp.SetStatusCode(status)
	c.resp.message = msg
	return nil
}

func (c *context) StatusNotFound(msg string) error {
	return c.Status(StatusNotFound, msg)
}

func (c *context) SetCookie(key, value string, expire int) {
	ck := &fasthttp.Cookie{}
	ck.SetKey(key)
	ck.SetValue(value)
	ck.SetMaxAge(expire)
	ck.SetHTTPOnly(true)
	ck.SetPath(`/`)
	c.fastCtx.Response.Header.SetCookie(ck)
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

func (c *context) URL() *url.URL {
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
