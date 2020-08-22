package apims

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/eqto/go-db"
)

const (
	queryTypeSelect = 1 + iota
	queryTypeGet
	queryTypeInsert
	queryTypeUpdate
	queryTypeDelete
)

var (
	errMissingParameter = errors.New(`error missing required parameter: %s`)
	errExecutingQuery   = errors.New(`error executing query: %s`)
)

type queryAction struct {
	Action

	q         string
	qType     uint8
	qProperty string
	qParams   []string
	qBuilder  *db.QueryBuilder

	arrayName string
}

//Property ...
func (q *queryAction) property() string {
	return q.qProperty
}

//Params ..
func (q *queryAction) params() []string {
	return q.qParams
}

func (q *queryAction) execute(ctx *context) (interface{}, error) {
	values := []interface{}{}

	for _, param := range q.qParams {
		val := ctx.get(param)
		if val == nil {
			return nil, fmt.Errorf(errMissingParameter.Error(), param)
		}
		values = append(values, val)
	}

	switch q.qType {
	case queryTypeInsert:

	default: //select
		if q.qBuilder != nil { //run sql function
			rs, e := ctx.tx.Select(q.q, values...)
			if e != nil {
				return nil, fmt.Errorf(errExecutingQuery.Error(), q.q)
			}
			if rs == nil {
				rs = []db.Resultset{}
			}
			if q.qType == queryTypeGet {
				return rs[0], nil
			}
			return rs, nil
		}
	}
	// 	if q.queryType == `SELECT` || q.queryType == `GET` {
	// 		if q.queryBuilder == nil { //run sql function
	// 			rs, e := ctx.selectQuery(q.query, values...)

	// 			if e != nil {
	// 				return nil, e
	// 			}
	// 			if rs == nil {
	// 				rs = []db.Resultset{}
	// 			}
	// 			if regexSQLFunc.MatchString(q.query) && len(rs) >= 1 {
	// 				return rs[0], nil
	// 			}
	// 			return rs, nil
	// 		}
	// 		qb := *q.queryBuilder
	// 		if filter := req.GetJSONObject(`filter`); filter != nil {
	// 			for keyFilter := range filter {
	// 				valFilter := filter.GetJSONObject(keyFilter)
	// 				filterType, filterValue := ``, ``
	// 				if valFilter == nil {
	// 					filterValue = filter.GetString(keyFilter)
	// 				} else {
	// 					filterValue = valFilter.GetString(`value`)
	// 					filterType = valFilter.GetString(`type`)
	// 				}

	// 				filterValue = strings.TrimSpace(filterValue)
	// 				switch filterType {
	// 				case `fulltext`:
	// 					qb.WhereOp(keyFilter, filterType)
	// 					values = append(values, filterValue+`*`)
	// 				case `input`:
	// 					qb.Where(keyFilter)
	// 					values = append(values, filterValue)
	// 				case `number`:
	// 					if strings.HasPrefix(filterValue, `<`) {
	// 						filterValue = strings.TrimSpace(filterValue[1:])
	// 						qb.WhereOp(keyFilter, ` < `)
	// 					} else if strings.HasPrefix(filterValue, `>`) {
	// 						filterValue = strings.TrimSpace(filterValue[1:])
	// 						qb.WhereOp(keyFilter, ` > `)
	// 					} else {
	// 						qb.Where(keyFilter)
	// 					}
	// 					values = append(values, filterValue)
	// 				case `date`:
	// 					date := strings.SplitN(filterValue, ` - `, 2)
	// 					if len(date) == 2 {
	// 						if start, e := time.Parse(`2006-01-02`, date[0]); e == nil {
	// 							if end, e := time.Parse(`2006-01-02`, date[1]); e == nil {
	// 								end = end.Add(24 * time.Hour)

	// 								qb.WhereOp(keyFilter, `>=`)
	// 								values = append(values, start.Format(`2006-01-02 15:04:05`))

	// 								qb.WhereOp(keyFilter, `<`)
	// 								values = append(values, end.Format(`2006-01-02 15:04:05`))
	// 							}
	// 						}
	// 					}
	// 				default:
	// 					qb.Where(keyFilter)
	// 					values = append(values, filterValue)
	// 				}
	// 			}
	// 		}
	// 		page := req.GetInt(`page`)

	// 		maxRows := req.GetInt(`max_rows`)
	// 		if maxRows == 0 {
	// 			maxRows = qb.LimitLength()
	// 			if maxRows == 0 {
	// 				maxRows = 100
	// 			}
	// 		}
	// 		if page >= 1 {
	// 			qb.Limit((page-1)*maxRows, maxRows)
	// 		}
	// 		active, direction := req.GetString(`sort.active`), req.GetString(`sort.direction`)

	// 		if active != `` && direction != `` {
	// 			qb.Order(active, direction)
	// 		}

	// 		if q.queryType == `GET` {
	// 			qb.Limit(qb.LimitStart(), 1)
	// 			rs, e := ctx.getQuery(qb.ToSQL(), values...)
	// 			if e != nil {
	// 				return nil, e
	// 			}
	// 			if rs == nil {
	// 				rs = db.Resultset{}
	// 			}
	// 			return rs, nil
	// 		}
	// 		rs, e := ctx.selectQuery(qb.ToSQL(), values...)

	// 		if e != nil {
	// 			return nil, e
	// 		}
	// 		if rs == nil {
	// 			rs = []db.Resultset{}
	// 		}
	// 		return rs, nil
	// 	}

	// 	switch q.queryType {
	// 	case `INSERT`:
	// 		rs, e := ctx.execQuery(q.query, values...)
	// 		if e != nil {
	// 			return nil, e
	// 		}
	// 		return rs.LastInsertID()

	// 	case `UPDATE`:
	// 		fallthrough
	// 	case `DELETE`:
	// 		rs, e := ctx.execQuery(q.query, values...)
	// 		if e != nil {
	// 			return nil, e
	// 		}
	// 		return rs.RowsAffected()
	// 	}
	// 	return nil, fmt.Errorf(ErrUnknownQueryType.Error(), q.queryType)
	return nil, nil
}

func newQueryAction(query, property, params string) (*queryAction, error) {
	act := &queryAction{q: query, qProperty: property}

	str := strings.SplitN(query, ` `, 2)
	queryType := strings.ToUpper(str[0])
	switch queryType {
	case `SELECT`:
		act.qBuilder = db.ParseQuery(query)
		act.qType = queryTypeSelect
		if regexp.MustCompile(`LIMIT.*\s+1$`).MatchString(strings.ToUpper(query)) {
			act.qType = queryTypeGet
		}
	case `INSERT`:
		act.qType = queryTypeInsert
	case `UPDATE`:
		act.qType = queryTypeUpdate
	case `DELETE`:
		act.qType = queryTypeDelete
	}

	params = strings.ReplaceAll(strings.TrimSpace(params), ` `, ``)
	if params != `` {
		regex := regexp.MustCompile(`(?Uis)\s*^([a-z0-9._]+)\[([a-z0-9._]+)\]\s*$`)

		act.qParams = strings.Split(params, `,`)
		for _, val := range act.qParams {
			matches := regex.FindStringSubmatch(val)
			if len(matches) == 3 {
				if act.arrayName != `` && act.arrayName != matches[1] {
					return nil, errors.New(`multiple array in single query is prohibited`)
				}
				act.arrayName = matches[1]
				// q.params = append(q.params, QueryParam{name: matches[1], indexName: matches[2]})
			} else {
				// q.params = append(q.params, QueryParam{name: val})
			}
		}
	}
	return act, nil
}
