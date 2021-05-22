package api

import (
	"errors"
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
	Write(value interface{}) error

	//WriteBody write to body and ignoring the next action
	WriteBody(contentType string, body []byte) error
	//Status stop execution and return status and message
	Status(status int, msg string) error
	//Error stop execution, rollback database transaction, and return status and message
	Error(status int, msg string) error

	//StatusBadRequest 400
	StatusBadRequest(msg string) error
	//StatusUnauthorized 401
	StatusUnauthorized(msg string) error
	//StatusUnauthorized 403
	StatusForbidden(msg string) error
	//StatusNotFound 404
	StatusNotFound(msg string) error
	StatusServiceUnavailable(msg string) error
	StatusInternalServerError(msg string) error

	Request() Request
	Response() Response

	Tx() (*db.Tx, error)
	Session() Session
	SetValue(name string, value interface{})
	GetValue(name string) interface{}
}

type context struct {
	Context

	s *Server

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
	if !c.resp.stop && c.property != `` {
		c.put(c.property, value)
	}
	return nil
}

func (c *context) WriteBody(contentType string, body []byte) error {
	if !c.resp.stop {
		resp := c.Response()
		if contentType != `` {
			resp.SetContentType(contentType)
		}
		resp.Body()
		resp.SetBody(body)
		c.resp.stop = true
	}
	return nil
}

func (c *context) Status(code int, msg string) error {
	if !c.resp.stop {
		c.resp.statusCode = code
		c.resp.statusMsg = msg
		c.resp.stop = true
	}
	return nil
}

func (c *context) Error(code int, msg string) error {
	c.Status(code, msg)
	return errors.New(msg)
}

func (c *context) StatusBadRequest(msg string) error {
	return c.httpError(StatusBadRequest, StatusBadRequest, msg)
}

func (c *context) StatusForbidden(msg string) error {
	return c.httpError(StatusForbidden, StatusForbidden, msg)
}
func (c *context) StatusNotFound(msg string) error {
	return c.httpError(StatusNotFound, StatusNotFound, msg)
}

func (c *context) StatusInternalServerError(msg string) error {
	return c.httpError(StatusInternalServerError, StatusInternalServerError, msg)
}

func (c *context) StatusServiceUnavailable(msg string) error {
	return c.httpError(StatusServiceUnavailable, StatusServiceUnavailable, msg)
}

func (c *context) httpError(httpCode, statusCode int, msg string) error {
	c.Response().SetStatusCode(statusCode)
	return c.Error(statusCode, msg)
}

func (c *context) Tx() (*db.Tx, error) {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	cn := c.s.Database()
	if cn != nil && c.tx == nil {
		tx, e := cn.Begin()
		if e != nil { //db error
			c.setErr(e)
			return nil, c.StatusServiceUnavailable(`Service unavailable`)
		}
		c.tx = tx
	}
	return c.tx, nil
}

func (c *context) URL() *url.URL {
	return c.req.URL()
}

func (c *context) Method() string {
	return c.req.Method()
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
		if c.resp.data == nil {
			c.resp.data = json.Object{}
		}
		c.resp.data.Put(property, value)
	}
}

func (c *context) setErr(err error) {
	if c.resp.err == nil {
		c.resp.err = err
	}
}

func (c *context) logger() *logger {
	return c.s.logger
}

func newContext(s *Server, req *fasthttp.Request, resp *fasthttp.Response) (*context, error) {
	ctx := &context{
		s:      s,
		values: make(map[string]interface{}),
		sess:   &session{},
		lockCn: sync.Mutex{},
	}

	req.CopyTo(&ctx.req.httpReq)
	resp.CopyTo(&ctx.resp.httpResp)

	return ctx, nil
}
