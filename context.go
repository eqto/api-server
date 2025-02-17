package api

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/eqto/dbm"
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

// type Context interface {
// 	URL() *url.URL
// 	Method() string
// 	ContentType() string
// 	RemoteIP() string

// 	Write(value interface{}) error
// 	//WriteBody write to body and ignoring the next action
// 	WriteBody(contentType string, body []byte) error
// 	//WriteStream
// 	WriteStream(filename, contentType string, writeFunc func(Writer)) error

// 	//Status stop execution and return status and message
// 	Status(status int, msg string) error
// 	//Error stop execution, rollback database transaction, and return status and message
// 	Error(status int, msg string) error

// 	//StatusBadRequest 400
// 	StatusBadRequest(msg string) error
// 	//StatusUnauthorized 401
// 	StatusUnauthorized(msg string) error
// 	//StatusUnauthorized 403
// 	StatusForbidden(msg string) error
// 	//StatusNotFound 404
// 	StatusNotFound(msg string) error
// 	StatusServiceUnavailable(msg string) error
// 	StatusInternalServerError(msg string) error

// 	Redirect(url string) error

// 	Request() Request
// 	RequiredParams(names string) (json.Object, error)
// 	Response() Response

// 	Database() (*dbm.Connection, error)
// 	Tx() (*dbm.Tx, error)
// 	Session() Session
// 	SetValue(name string, value interface{})
// 	GetValue(name string) interface{}
// }

type Context struct {
	s       *Server
	fastCtx *fasthttp.RequestCtx

	property string

	req  request
	resp *Response

	sess *session

	vars json.Object

	stdTx  *dbm.Tx
	lockCn sync.Mutex

	values map[string]interface{}
}

func (c *Context) Write(value interface{}) error {
	if c.property != `` {
		c.put(c.property, value)
	}
	return nil
}

func (c *Context) WriteStream(filename, contentType string, fn func(Writer)) error {
	c.resp.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment;filename="%s"`, filename))
	c.resp.Header().Set(`Content-Type`, contentType)
	sw := c.resp.streamWriter()
	go func() {
		defer sw.Close()
		fn(sw)
	}()
	return nil
}

func (c *Context) WriteBody(contentType string, body []byte) error {
	if !c.resp.stop {
		resp := c.Response()
		if contentType != `` {
			resp.SetContentType(contentType)
		}
		resp.setBody(body)
		c.resp.stop = true
	}
	return nil
}

func (c *Context) Status(code int, msg string) error {
	if !c.resp.stop {
		c.resp.statusCode = code
		c.resp.statusMsg = &msg
		c.resp.stop = true
	}
	return nil
}

func (c *Context) Error(code int, msg string) error {
	c.Status(code, msg)
	return errors.New(msg)
}

func (c *Context) StatusBadRequest(msg string) error {
	return c.httpError(StatusBadRequest, StatusBadRequest, msg)
}

func (c *Context) StatusUnauthorized(msg string) error {
	return c.httpError(StatusUnauthorized, StatusUnauthorized, msg)
}

func (c *Context) StatusForbidden(msg string) error {
	return c.httpError(StatusForbidden, StatusForbidden, msg)
}

func (c *Context) StatusNotFound(msg string) error {
	return c.httpError(StatusNotFound, StatusNotFound, msg)
}

func (c *Context) StatusServiceUnavailable(msg string) error {
	return c.httpError(StatusServiceUnavailable, StatusServiceUnavailable, msg)
}

func (c *Context) StatusInternalServerError(msg string) error {
	return c.httpError(StatusInternalServerError, StatusInternalServerError, msg)
}

func (c *Context) Redirect(url string) error {
	c.fastCtx.Redirect(url, fasthttp.StatusFound)
	return nil
}

func (c *Context) Database() (*dbm.Connection, error) {
	if c.s.cn == nil {
		return nil, errors.New(`database not available`)
	}
	return c.s.cn, nil
}

func (c *Context) Tx() (*dbm.Tx, error) {
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

func (c *Context) URL() *url.URL {
	return c.req.URL()
}

func (c *Context) Method() string {
	return c.req.Method()
}

func (c *Context) ContentType() string {
	return c.req.Header().Get(`Content-Type`)
}

// Session ..
func (c *Context) Session() Session {
	return c.sess
}

func (c *Context) Request() Request {
	return &c.req
}
func (c *Context) RequiredParams(names string) (json.Object, error) {
	split := strings.Split(names, `,`)
	js := c.req.JSON().Clone()

	jsResp := json.Object{}
	for _, s := range split {
		name := strings.TrimSpace(s)
		if js.Has(name) {
			return nil, errors.New(`parameter not found: ` + name)
		}
		jsResp.Put(name, js.Get(name))
	}
	return jsResp, nil
}

func (c *Context) Response() *Response {
	return c.resp
}

func (c *Context) SetValue(name string, value interface{}) {
	c.values[name] = value
}
func (c *Context) GetValue(name string) interface{} {
	return c.values[name]
}

func (c *Context) RemoteIP() string {
	return string(c.fastCtx.RemoteIP().String())
}

func (c *Context) closeTx() {
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

func (c *Context) httpError(httpCode, statusCode int, msg string) error {
	c.resp.httpResp.SetStatusCode(httpCode)
	return c.Error(statusCode, msg)
}

func (c *Context) put(property string, value interface{}) {
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

func (c *Context) setErr(err error) {
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

func (c *Context) logger() *logger {
	return c.s.logger
}

func newContext(s *Server, fastCtx *fasthttp.RequestCtx) (*Context, error) {
	ctx := &Context{
		s:       s,
		values:  make(map[string]interface{}),
		sess:    &session{logger: s.logger},
		lockCn:  sync.Mutex{},
		fastCtx: fastCtx,
	}
	ctx.resp.httpResp = &fastCtx.Response
	ctx.req.fastCtx = fastCtx

	return ctx, nil
}
