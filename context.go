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

//Ctx ..
type Ctx interface {
	URI()
	Session() Session
}
type ctx struct {
	Ctx
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
func (c *ctx) Session() Session {
	return c.sess
}

func (c *ctx) URL() *url.URL {
	return c.url
}

func (c *ctx) begin() error {
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
func (c *ctx) rollback() {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.tx != nil {
		c.tx.Rollback()
		c.tx = nil
	}
}
func (c *ctx) commit() {
	c.lockCn.Lock()
	defer c.lockCn.Unlock()
	if c.tx != nil {
		c.tx.Commit()
	}
}

func (c *ctx) put(property string, value interface{}) {
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

func (c *ctx) getRequest(key string) interface{} {
	if c.jsonReq.Has(key) {
		return c.jsonReq.Get(key)
	}
	return c.url.Query().Get(key)
}

func newCtx(s *Server, req *fasthttp.Request, resp *fasthttp.Response, cn *db.Connection) (*ctx, error) {
	ctx := &ctx{
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
