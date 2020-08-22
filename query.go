package apims

import (
	"github.com/eqto/go-db"
)

//QueryParam ...
type QueryParam struct {
	name      string
	indexName string
}

//Query ...
type Query struct {
	query        string
	queryType    string
	packageName  string
	functionName string
	property     string
	queryBuilder *db.QueryBuilder
	params       []QueryParam
	arrayName    string
}

// func (q *Query) execFunc(ctx Context, values []interface{}) (interface{}, error) {
// 	packageName := q.packageName
// 	if packageName == `` {
// 		packageName = `system`
// 	}
// 	path := `libs/lib_` + packageName + `.api`
// 	var pl *plugin.Plugin

// 	if p := plugin.Get(path); p == nil {
// 		p, e := plugin.Load(path)
// 		if e != nil {
// 			return nil, e
// 		}
// 		m := p.Request()
// 		db := ctx.svr.db
// 		m.Add(ctx.svr.appID, db.hostname, db.port, db.username, db.password, db.name)
// 		m.Send(`init`)
// 		pl = p
// 	} else {
// 		p, e := plugin.Load(path)
// 		if e != nil {
// 			return nil, e
// 		}
// 		pl = p
// 	}

// 	m := pl.Request()
// 	for i := 0; i < len(values); i++ {
// 		m.Add(values[i])
// 	}
// 	m.Send(q.functionName)
// 	return nil, nil
// }

// func (q *Query) execQuery(ctx Context, values []interface{}) (interface{}, error) {
// 	req := ctx.Request()
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
// }
