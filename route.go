package api

import (
	"reflect"
	"regexp"
	"strings"

	"gitlab.com/tuxer/go-db"

	"gitlab.com/tuxer/go-json"
)

//RouteFunc ...
type RouteFunc func(ctx Context) error

//RouteConfig ...
type RouteConfig struct {
	query     string
	queryType string
	qb        *db.QueryBuilder
	params    []string
	output    string
	authType  string

	routeFunc RouteFunc
}

func (r *RouteConfig) putOutput(req *Request, resp *Response, value interface{}) {
	current := req.Get(r.output)

	//3 lines below to convert any value to recognized value
	js := make(json.Object)
	js.Put(r.output, value)
	new := js.Get(r.output)

	if reflect.TypeOf(current) == reflect.TypeOf(new) {
		if n, ok := new.(map[string]interface{}); ok {
			for key, val := range n {
				current.(map[string]interface{})[key] = val
			}
			new = current
		}
	}
	if strings.HasPrefix(r.output, `$`) {
		req.Put(r.output, new)
	} else {
		resp.Put(r.output, new)
	}
}

func (r *RouteConfig) process(ctx Context) error {
	req, resp, tx := ctx.Request(), ctx.Response(), ctx.Tx()

	values := []interface{}{}

	for _, val := range r.params {
		values = append(values, req.MustString(val))
	}
	qb := *r.qb

	if filter := req.GetJSONObject(`filter`); filter != nil {
		for keyFilter := range filter {
			valFilter := filter.GetJSONObject(keyFilter)
			value := valFilter.GetString(`value`)

			switch valFilter.GetString(`type`) {
			case `input`:
				qb.WhereOp(keyFilter, ` LIKE `)
				values = append(values, value+`%`)
			case `number`:
				value = strings.TrimSpace(value)
				if strings.HasPrefix(value, `<`) {
					value = strings.TrimSpace(value[1:])
					qb.WhereOp(keyFilter, ` < `)
				} else if strings.HasPrefix(value, `>`) {
					value = strings.TrimSpace(value[1:])
					qb.WhereOp(keyFilter, ` > `)
				}
				values = append(values, value)
			default:
				qb.Where(keyFilter)
				values = append(values, value)
			}
		}
	}
	switch r.queryType {
	case `INSERT`:
		rs, e := tx.Exec(r.query, values...)
		if e != nil {
			return e
		}
		id, e := rs.LastInsertID()
		if e != nil {
			return e
		}
		r.putOutput(req, resp, id)

	case `UPDATE`:
		rs, e := tx.Exec(r.query, values...)
		if e != nil {
			return e
		}
		rows, e := rs.RowsAffected()
		if e != nil {
			return e
		}
		r.putOutput(req, resp, rows)

	case `GET`:
		fallthrough
	case `SELECT`:
		page := req.GetInt(`page`)

		maxRows := req.GetInt(`max_rows`)
		if maxRows == 0 {
			maxRows = qb.LimitLength()
			if maxRows == 0 {
				maxRows = 100
			}
		}
		if page >= 1 {
			qb.Limit((page-1)*maxRows, maxRows)
		}
		active, direction := req.GetString(`sort.active`), req.GetString(`sort.direction`)

		if active != `` && direction != `` {
			qb.Order(active, direction)
		}

		if r.queryType == `GET` {
			qb.Limit(qb.LimitStart(), 1)
			rs, e := tx.Get(qb.ToSQL(), values...)
			if e != nil {
				return e
			}
			r.putOutput(req, resp, rs)
		} else {
			rs, e := tx.Select(qb.ToSQL(), values...)
			if e != nil {
				return e
			}
			r.putOutput(req, resp, rs)
		}
	}
	return nil
}

func newRouteConfig(cfg json.Object) *RouteConfig {
	r := &RouteConfig{
		query:     strings.TrimSpace(cfg.GetString(`query`)),
		queryType: strings.TrimSpace(cfg.GetString(`query_type`)),
		output:    strings.TrimSpace(cfg.GetStringOr(`output`, `data`)),
		authType:  cfg.GetString(`auth`),
	}
	if r.queryType == `` {
		str := strings.SplitN(r.query, ` `, 2)
		r.queryType = strings.ToUpper(str[0])
		if r.queryType == `SELECT` && regexp.MustCompile(`LIMIT.*\s+1$`).MatchString(strings.ToUpper(r.query)) {
			r.queryType = `GET`
		}
	}

	r.queryType = strings.ToUpper(r.queryType)
	r.qb = db.ParseQuery(r.query)

	if params := cfg.GetString(`params`); params != `` {
		split := strings.Split(params, `,`)
		for key, val := range split {
			split[key] = strings.TrimSpace(val)
		}
		r.params = split
	}
	return r
}
