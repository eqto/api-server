package api

import (
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

//Connection ...
func (c *Context) Connection() *db.Tx {
	return c.tx
}

func (c *Context) Server() *Server {
	return c.s
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
func (c *Context) getString(key string) string {
	if strings.HasPrefix(key, `$`) {
		return c.result.GetString(key)
	}
	return c.req.GetString(key)
}
func (c *Context) getInt(key string) int {
	if strings.HasPrefix(key, `$`) {
		return c.result.GetInt(key)
	}
	return c.req.GetInt(key)
}

func (c *Context) put(output string, value interface{}) {
	if output == `` {
		output = `data`
	}
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
