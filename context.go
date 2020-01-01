package api

import (
	"fmt"
	"strings"

	"gitlab.com/tuxer/go-db"
	"gitlab.com/tuxer/go-json"
)

//Context ...
type Context struct {
	// context.Context
	s      *Server
	req    *Request
	resp   *Response
	tx     *db.Tx
	result json.Object
}

//Request ...
func (c *Context) Request() *Request {
	if c.req == nil {
		c.req = new(Request)
	}
	return c.req
}

//Response ...
func (c *Context) Response() *Response {
	if c.resp == nil {
		c.resp = new(Response)
	}
	return c.resp
}

func (c *Context) execQuery(query string, params ...interface{}) (*db.Result, error) {
	return c.tx.Exec(query, params...)
}

func (c *Context) getQuery(query string, params ...interface{}) (db.Resultset, error) {
	return c.tx.Get(query, params...)
}

func (c *Context) selectQuery(query string, params ...interface{}) ([]db.Resultset, error) {
	return c.tx.Select(query, params...)
}

func (c *Context) get(key string) interface{} {
	if strings.HasPrefix(key, `$`) {
		return c.result.Get(key)
	}
	return c.req.Get(key)
}

func (c *Context) execFunc(name string) (interface{}, error) {
	if strings.HasSuffix(name, `()`) {
		name = name[0 : len(name)-2]
	}
	if f, ok := c.s.funcMap[name]; ok {
		return f(c)
	}
	return nil, fmt.Errorf(ErrFunctionNotFound.Error(), name)
}

func (c *Context) put(output string, value interface{}) {
	current := c.get(output)

	value = json.NormalizeValue(value)

	if value, ok := value.(map[string]interface{}); ok {
		if current, ok := current.(map[string]interface{}); ok {
			for key, val := range current {
				value[key] = val
			}
		}
	}

	if strings.HasPrefix(output, `$`) {
		if c.result == nil {
			c.result = make(json.Object)
		}
		c.result.Put(output, value)
	} else {
		c.resp.Put(output, value)
	}
}
