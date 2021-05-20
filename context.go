package api

import (
	"strings"
	"sync"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
	"github.com/valyala/fasthttp"
)

//Context ..
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
	req    request
	resp   response
	sess   *session
	logger *logger

	vars json.Object

	cn     *db.Connection
	tx     *db.Tx
	lockCn sync.Mutex

	values map[string]interface{}
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

func newContext(fastCtx *fasthttp.RequestCtx) (*context, error) {
	ctx := &context{
		req:    request{fastCtx: fastCtx},
		resp:   response{fastCtx: fastCtx},
		values: make(map[string]interface{}),
		sess:   &session{},
		lockCn: sync.Mutex{},
	}

	return ctx, nil
}
