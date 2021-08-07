package api

import (
	"errors"
	"net/url"
	"strings"
	"sync"

	"github.com/eqto/dbm"
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

//Context ..
type Context interface {
	URL() *url.URL
	Method() string
	ContentType() string
	RemoteIP() string
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

	Redirect(url string) error

	Request() Request
	Response() Response

	Tx() (*dbm.Tx, error)
	Session() Session
	SetValue(name string, value interface{})
	GetValue(name string) interface{}
}

type context struct {
	Context

	s       *Server
	fastCtx *fasthttp.RequestCtx

	property string

	req  request
	resp response

	sess *session

	vars json.Object

	stdTx  *dbm.Tx
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

func (c *context) StatusUnauthorized(msg string) error {
	return c.httpError(StatusUnauthorized, StatusUnauthorized, msg)
}

func (c *context) StatusForbidden(msg string) error {
	return c.httpError(StatusForbidden, StatusForbidden, msg)
}

func (c *context) StatusNotFound(msg string) error {
	return c.httpError(StatusNotFound, StatusNotFound, msg)
}

func (c *context) StatusServiceUnavailable(msg string) error {
	return c.httpError(StatusServiceUnavailable, StatusServiceUnavailable, msg)
}

func (c *context) StatusInternalServerError(msg string) error {
	return c.httpError(StatusInternalServerError, StatusInternalServerError, msg)
}

func (c *context) Redirect(url string) error {
	c.fastCtx.Redirect(url, fasthttp.StatusFound)
	return nil
}

func (c *context) Tx() (*dbm.Tx, error) {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	cn := c.s.Database()
	if cn != nil && c.stdTx == nil {
		tx, e := cn.Begin()
		if e != nil { //db error
			c.setErr(e)
			return nil, c.StatusServiceUnavailable(`Service unavailable`)
		}
		c.stdTx = tx
	}
	return c.stdTx, nil
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

func (c *context) RemoteIP() string {
	return string(c.fastCtx.RemoteIP().String())
}

func (c *context) closeTx() {
	if c.stdTx == nil {
		return
	}
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.resp.err != nil {
		c.stdTx.Rollback()
	} else {
		c.stdTx.Commit()
	}
	c.stdTx = nil
}

func (c *context) httpError(httpCode, statusCode int, msg string) error {
	c.resp.httpResp.SetStatusCode(httpCode)
	return c.Error(statusCode, msg)
}

func (c *context) put(property string, value interface{}) {
	if strings.HasPrefix(property, `$`) { //save to vars
		if c.vars == nil {
			c.vars = json.Object{}
		}
		if property == `$` {
			if js, e := json.ParseObject(value); e == nil {
				if c.resp.data == nil {
					c.resp.data = json.Object{}
				}
				js.CopyTo(&c.resp.data)
			}
		} else {
			c.vars.Put(property[1:], value)
		}
	} else { //save to result
		c.resp.put(property, value)
	}
}

func (c *context) setErr(err error) {
	if c.resp.err == nil {
		c.resp.err = err
	}
	if c.resp.data == nil {
		c.resp.data = json.Object{}
	}
	if c.resp.statusCode == 0 {
		c.Error(99, err.Error())
	}
}

func (c *context) logger() *logger {
	return c.s.logger
}

func newContext(s *Server, fastCtx *fasthttp.RequestCtx) (*context, error) {
	ctx := &context{
		s:       s,
		values:  make(map[string]interface{}),
		sess:    &session{logger: s.logger},
		lockCn:  sync.Mutex{},
		fastCtx: fastCtx,
	}
	ctx.resp.httpResp = &fastCtx.Response
	fastCtx.Request.CopyTo(&ctx.req.httpReq)

	return ctx, nil
}
