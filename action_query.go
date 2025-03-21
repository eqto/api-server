package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eqto/dbm"
	"github.com/eqto/dbm/stmt"
	"github.com/eqto/go-json"
	"github.com/pkg/errors"
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
	errExecutingQuery   = errors.New(`error executing query`)
)

type actionQuery struct {
	Action

	rawSql    string
	qType     uint8
	qProperty string
	qParams   []string

	arrayName  string
	selectStmt *stmt.Select
}

func (q *actionQuery) AssignTo(prop string) Action {
	q.qProperty = prop
	return q
}

func (q *actionQuery) property() string {
	return q.qProperty
}

func (q *actionQuery) params() []string {
	return q.qParams
}

func (q *actionQuery) executeItem(ctx *Context, values []interface{}) (interface{}, error) {
	var data interface{}
	var err error
	var selectStmt *stmt.Select

	if (q.qType == queryTypeSelect || q.qType == queryTypeGet) && q.selectStmt != nil {
		//example filter for books title contains word = 'Programming'
		// {
		//   "title": {
		//     "value": "Programming",
		//     "filter": "fulltext"
		//   }
		// }
		js := ctx.req.JSON()
		selectStmt = new(stmt.Select)

		if e := stmt.Copy(selectStmt, q.selectStmt); e != nil {
			return nil, e
		}

		if filters := js.GetJSONObject(`filters`); len(filters) > 0 {
			for key := range filters {
				js := filters.GetJSONObject(key)
				parsed := parseFilter(key, js.GetString(`value`), js.GetString(`type`), selectStmt)
				values = append(values, parsed...)
			}
		} else if filters := js.GetArray(`filters`); len(filters) > 0 {
			for _, filter := range filters {
				parsed := parseFilter(filter.GetString(`name`), filter.GetString(`value`), filter.GetString(`type`), selectStmt)
				values = append(values, parsed...)
			}
		}
		// Example:
		// {
		//   "sort": {
		// 	   "active": "created_at",
		// 	   "direction": "asc"
		//   }
		// }
		if sort := js.GetJSONObject(`sort`); sort != nil {
			if active := sort.GetString(`active`); active != `` {
				selectStmt.OrderBy(fmt.Sprintf(`%s %s`, active, sort.GetStringOr(`direction`, `asc`)))
			}
		} else if sorts := js.GetArray(`sort`); len(sorts) > 0 {
			sortStrings := []string{}
			for _, sort := range sorts {
				if active := sort.GetString(`active`); active != `` {
					sortStrings = append(sortStrings, fmt.Sprintf(`%s %s`, active, sort.GetStringOr(`direction`, `asc`)))
				}
			}
			if len(sortStrings) > 0 {
				selectStmt.OrderBy(strings.Join(sortStrings, `,`))
			}

		}
		// Example:
		// {
		//   "page": {
		// 	   "offset": 2,
		// 	   "count": 100
		//   }
		// }
		if page := js.GetJSONObject(`page`); page != nil {
			if offset := page.GetInt(`offset`); offset > 0 {
				selectStmt.Offset(offset)
			}
			if count := page.GetInt(`count`); count > 0 {
				selectStmt.Count(count)
			}
		}
	}

	tx, e := ctx.Tx()
	if e != nil {
		ctx.debugLog.logErr(errors.Wrap(e, `database connection failed`))
		return nil, errors.New(`database connection failed`)
	}

	switch q.qType {
	case queryTypeSelect:
		sql := q.rawSql
		if selectStmt != nil {
			if _, count := stmt.LimitOf(selectStmt); count == 0 {
				selectStmt.Count(1000)
			}
			sql = ctx.s.cn.Driver().StatementString(selectStmt)
		}
		data, err = tx.Select(sql, values...)
	case queryTypeGet:
		sql := q.rawSql
		if selectStmt != nil {
			sql = ctx.s.cn.Driver().StatementString(selectStmt)
		}
		res, e := tx.Get(sql, values...)
		if e != nil {
			err = e
		} else if res != nil {
			if len(res) > 1 {
				data = res
			} else {
				for key := range res {
					data = res[key]
				}
			}
		}
	case queryTypeUpdate:
		fallthrough
	case queryTypeInsert:
		fallthrough
	case queryTypeDelete:
		data, err = tx.Exec(q.rawSql, values...)
	}
	if err != nil {
		if dbm.IsErrDuplicate(err) {
			ctx.debugLog.logErr(errors.Wrap(e, `duplicate entry`))
			return nil, errors.New(`duplicate entry`)
		}
		ctx.debugLog.logErr(fmt.Errorf(`%s. Query: %s`, err, q.rawSql))
		return nil, errExecutingQuery
	}
	switch q.qType {
	case queryTypeInsert:
		if id, e := data.(*dbm.Result).LastInsertID(); e == nil {
			data = id
		}
	case queryTypeUpdate:
		fallthrough
	case queryTypeDelete:
		if rows, e := data.(*dbm.Result).RowsAffected(); e == nil {
			data = rows
		}
	}
	return data, nil
}

// return value
func parseFilter(name, value, typ string, selectStmt *stmt.Select) []interface{} {
	values := []interface{}{}

	switch strings.ToUpper(typ) {
	case `FULLTEXT`:
		selectStmt.Where(fmt.Sprintf(`MATCH(%s) AGAINST(? IN BOOLEAN MODE)`, name))
		values = append(values, value+`*`)
	case `DATE`:
		selectStmt.Where(fmt.Sprintf(`%s >= ?`, name))
		selectStmt.Where(fmt.Sprintf(`%s < ?`, name))
		time, _ := time.Parse(`2006-01-02`, value)
		values = append(values, time.Format(`2006-01-02`))
		time = time.AddDate(0, 0, 1)
		values = append(values, time.Format(`2006-01-02`))
	case `IN`:
		vals := strings.Split(value, `,`)
		names := []string{}
		for _, val := range vals {
			names = append(names, `?`)
			values = append(values, strings.TrimSpace(val))
		}
		selectStmt.Where(fmt.Sprintf(`%s IN (%s)`, name, strings.Join(names, `,`)))
	default:
		if typ == `` {
			typ = `=`
		}
		selectStmt.Where(fmt.Sprintf(`%s %s ?`, name, typ))
		values = append(values, value)
	}
	return values
}

func (q *actionQuery) populateValues(ctx *Context, item interface{}) ([]interface{}, error) {
	values := []interface{}{}
	for _, param := range q.qParams {
		if strings.HasPrefix(param, `$session.`) {
			values = append(values, ctx.sess.GetString(param[9:]))
		} else if strings.HasPrefix(param, `$`) {
			values = append(values, ctx.vars.Get(param[1:]))
		} else if strings.HasPrefix(param, q.arrayName+`[`) && strings.HasSuffix(param, `]`) {
			if js, ok := item.(json.Object); ok {
				val := js.Get(param[len(q.arrayName)+1 : len(param)-1])
				values = append(values, val)
			} else {
				values = append(values, item)
			}
		} else {
			val := ctx.req.get(param)
			if val == nil {
				return nil, fmt.Errorf(errMissingParameter.Error(), param)
			}
			values = append(values, val)
		}
	}
	return values, nil
}

func (q *actionQuery) execute(ctx *Context) error {
	if q.arrayName != `` { //execute array
		result := []interface{}{}

		js := ctx.req.JSON()
		if objs := js.GetArray(q.arrayName); objs != nil {
			for _, obj := range objs {
				values, e := q.populateValues(ctx, obj)
				if e != nil {
					return e
				}
				r, e := q.executeItem(ctx, values)
				if e != nil {
					return e
				}
				if recs, ok := r.([]dbm.Resultset); ok {
					for _, rec := range recs {
						result = append(result, rec)
					}
				} else {
					result = append(result, r)
				}
			}
		} else if arr := js.Array(q.arrayName); arr != nil {
			for _, val := range arr {
				values, e := q.populateValues(ctx, val)
				if e != nil {
					return e
				}
				r, e := q.executeItem(ctx, values)
				if e != nil {
					return e
				}
				if recs, ok := r.([]dbm.Resultset); ok {
					for _, rec := range recs {
						result = append(result, rec)
					}
				} else {
					result = append(result, r)
				}
			}
		}
		return ctx.Write(result)
	}
	values, e := q.populateValues(ctx, nil)
	if e != nil {
		return e
	}
	r, e := q.executeItem(ctx, values)
	if e != nil {
		return e
	}
	return ctx.Write(r)
}

func newQueryAction(sql, params string) (*actionQuery, error) {
	act := &actionQuery{rawSql: sql}

	str := strings.SplitN(sql, ` `, 2)
	queryType := strings.ToUpper(str[0])
	switch queryType {
	case `SELECT`:
		act.qType = queryTypeSelect
		sel := stmt.Parse(sql)
		if sel, ok := sel.(*stmt.Select); ok {
			act.selectStmt = sel
			if _, count := stmt.LimitOf(sel); count == 1 {
				act.qType = queryTypeGet
			}
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
		regex := regexp.MustCompile(`(?Uis)\s*^([a-z0-9._]+)\[([a-z0-9._]*)\]\s*$`)

		act.qParams = strings.Split(params, `,`)
		for _, val := range act.qParams {
			matches := regex.FindStringSubmatch(val)
			if len(matches) == 3 {
				if act.arrayName != `` && act.arrayName != matches[1] {
					return act, errors.New(`multiple array in single query is prohibited`)
				}
				act.arrayName = matches[1]
			}
		}
	}
	return act, nil
}
