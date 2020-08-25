package apims

import (
	"strings"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
)

//Context ...
type Context interface {
	Request() Request
	Response() Response
	Session() Session
	Tx() *db.Tx
}

type context struct {
	tx   *db.Tx
	req  *request
	resp *response
	sess *session

	vars json.Object
}

func (c *context) Request() Request {
	return c.req
}

func (c *context) Response() Response {
	return c.resp
}

func (c *context) Session() Session {
	return &session{val: c.sess.val.Clone()}
}

func (c *context) Tx() *db.Tx {
	return c.tx
}

func (c *context) put(property string, value interface{}) {
	if strings.HasPrefix(property, `$`) { //save to vars
		c.vars.Put(property[1:], value)
	} else { //save to result
		c.resp.Put(property, value)
	}
}
func (c *context) get(property string) interface{} {
	if strings.HasPrefix(property, `$`) { //get from to vars
		return c.vars.Get(property[1:])
	} else { //get from result
		return c.resp.Get(property)
	}
}
