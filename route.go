package api

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/tuxer/go-db"

	"gitlab.com/tuxer/go-json"
)

//RouteFunc ...
type RouteFunc func(ctx Context) (interface{}, error)

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

func (r *Route) execFunc(ctx Context) (interface{}, error) {
	funcName := r.query
	if strings.HasSuffix(funcName, `()`) {
		funcName = funcName[0 : len(funcName)-2]
	}
	if f, ok := ctx.s.funcMap[funcName]; ok {
		regex := regexp.MustCompile(`(?Uis)(.*)AS\s+([\w]+)$`)
		params := []Parameter{}

		for _, param := range r.params {
			p := Parameter{}
			if matches := regex.FindStringSubmatch(param); len(matches) > 0 {
				p.set(ctx.get(matches[1]))
			} else {
				p.set(ctx.get(param))
			}
			params = append(params, p)
		}
		return f(ctx, params...)
	}
	return nil, fmt.Errorf(ErrFunctionNotFound.Error(), funcName)
}

func (r *Route) process(ctx Context) (interface{}, error) {
	req := ctx.Request()

	if r.queryType == `FUNC` {
		return r.execFunc(ctx)
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
			filterType, filterValue := ``, ``
			if valFilter == nil {
				filterValue = filter.GetString(keyFilter)
			} else {
				filterValue = valFilter.GetString(`value`)
				filterType = valFilter.GetString(`type`)
			}

			switch filterType {
			case `input`:
				qb.WhereOp(keyFilter, ` LIKE `)
				values = append(values, filterValue+`%`)
			case `number`:
				filterValue = strings.TrimSpace(filterValue)
				if strings.HasPrefix(filterValue, `<`) {
					filterValue = strings.TrimSpace(filterValue[1:])
					qb.WhereOp(keyFilter, ` < `)
				} else if strings.HasPrefix(filterValue, `>`) {
					filterValue = strings.TrimSpace(filterValue[1:])
					qb.WhereOp(keyFilter, ` > `)
				}
				values = append(values, filterValue)
			default:
				qb.Where(keyFilter)
				values = append(values, filterValue)
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
		fallthrough
	case `DELETE`:
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

//NewRoute ...
func NewRoute(query, params, output string) *Route {
	r := &Route{}
	if output != `` {
		r.output = output
	}
	r.setQuery(query)

	if params != `` {
		split := strings.Split(params, `,`)
		for key, val := range split {
			split[key] = strings.TrimSpace(val)
		}
		r.params = split
	}
	return r

}

//RouteFromJSON ...
func RouteFromJSON(cfg json.Object) *Route {
	r := NewRoute(cfg.GetString(`query`), cfg.GetString(`params`), cfg.GetStringOr(`output`, `data`))
	if t := cfg.GetString(`query_type`); t != `` {
		r.queryType = t
	}
	return r
}
