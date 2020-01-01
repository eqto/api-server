package api

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/tuxer/go-db"

	"gitlab.com/tuxer/go-json"
)

//RouteFunc ...
type RouteFunc func(ctx *Context) (interface{}, error)

//Route ...
type Route struct {
	query     string
	queryType string
	qb        *db.QueryBuilder
	params    []string
	output    string
	authType  string

	routeFunc RouteFunc
}

func (r *Route) process(ctx *Context) (interface{}, error) {
	req := ctx.Request()

	if r.queryType == `FUNC` {
		return ctx.execFunc(r.query)
	}

	values := []interface{}{}

	for _, key := range r.params {
		val := ctx.get(key)
		if val == nil {
			panic(fmt.Errorf(ErrMissingParameter.Error(), key))
		}
		values = append(values, val)
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
		rs, e := ctx.execQuery(r.query, values...)
		if e != nil {
			return nil, e
		}
		return rs.LastInsertID()

	case `UPDATE`:
		rs, e := ctx.execQuery(r.query, values...)
		if e != nil {
			return nil, e
		}
		return rs.RowsAffected()

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
			rs, e := ctx.getQuery(qb.ToSQL(), values...)
			if e != nil {
				return nil, e
			}
			if rs == nil {
				rs = db.Resultset{}
			}
			return rs, nil
		}
		rs, e := ctx.selectQuery(qb.ToSQL(), values...)
		if e != nil {
			return nil, e
		}
		if rs == nil {
			rs = []db.Resultset{}
		}
		return rs, nil
	}
	return nil, fmt.Errorf(ErrUnknownQueryType.Error(), r.queryType)
}

func (r *Route) setQuery(query string) {
	r.query = strings.TrimSpace(query)
	if r.queryType == `` {
		if strings.HasSuffix(r.query, `()`) {
			r.queryType = `FUNC`
		} else {
			str := strings.SplitN(r.query, ` `, 2)
			r.queryType = strings.ToUpper(str[0])
			if r.queryType == `SELECT` && regexp.MustCompile(`LIMIT.*\s+1$`).MatchString(strings.ToUpper(r.query)) {
				r.queryType = `GET`
			}
		}
	} else {
		r.queryType = strings.ToUpper(r.queryType)
	}
	if r.queryType != `FUNC` {
		r.qb = db.ParseQuery(r.query)
	}
}

//RouteWithFunc ...
func RouteWithFunc(f RouteFunc) *Route {
	return &Route{
		routeFunc: f,
	}
}

//RouteFromJSON ...
func RouteFromJSON(cfg json.Object) *Route {
	r := &Route{
		queryType: strings.TrimSpace(cfg.GetString(`query_type`)),
		output:    strings.TrimSpace(cfg.GetStringOr(`output`, `data`)),
		authType:  cfg.GetString(`auth`),
	}
	r.setQuery(cfg.GetString(`query`))

	if params := cfg.GetString(`params`); params != `` {
		split := strings.Split(params, `,`)
		for key, val := range split {
			split[key] = strings.TrimSpace(val)
		}
		r.params = split
	}
	return r
}
