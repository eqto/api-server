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
	URI()
	Session() Session
}
type context struct {
	Context
	s    *Server
	req  *fasthttp.Request
	resp *fasthttp.Response
	sess *session
	url  *url.URL

	jsonReq, jsonResp, vars json.Object

	cn     *db.Connection
	tx     *db.Tx
	lockCn sync.Mutex
}

//Session ..
func (c *context) Session() Session {
	return c.sess
}

func (c *context) URL() *url.URL {
	return c.url
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
		if c.jsonResp == nil {
			c.jsonResp = json.Object{}
		}
		c.jsonResp.Put(property, value)
	}
}

func (c *context) getRequest(key string) interface{} {
	if c.jsonReq.Has(key) {
		return c.jsonReq.Get(key)
	}
	return c.url.Query().Get(key)
}

func newCtx(s *Server, req *fasthttp.Request, resp *fasthttp.Response, cn *db.Connection) (*context, error) {
	ctx := &context{
		s:      s,
		req:    req,
		resp:   resp,
		cn:     cn,
		sess:   &session{},
		lockCn: sync.Mutex{},
	}
	ctx.url, _ = url.Parse(string(req.RequestURI()))

	if bytes.HasPrefix(req.Header.ContentType(), []byte(`application/json`)) {
		req, e := json.Parse(req.Body())
		if e != nil {
			return nil, errors.Wrap(e, `invalid json body`)
		}
		ctx.jsonReq = req
	} else {
		ctx.jsonReq = json.Object{}
	}
	return ctx, nil
}
