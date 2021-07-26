package api

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eqto/go-db"
	"github.com/eqto/go-json"
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

	rawQuery  string
	qType     uint8
	qProperty string
	qParams   []string

	arrayName string
	builder   *db.QueryBuilder
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

func (q *actionQuery) executeItem(ctx *context, values []interface{}) (interface{}, error) {
	var data interface{}
	var err error

	var builder *db.QueryBuilder
	if (q.qType == queryTypeSelect || q.qType == queryTypeGet) && q.builder != nil {
		builder = q.builder.Clone()
		//example filter for books title contains word = 'Programming'
		// {
		//   "title": {
		//     "value": "Programming",
		//     "filter": "fulltext"
		//   }
		// }
		js := ctx.req.JSON()

		if filters := js.GetJSONObject(`filters`); filters != nil && len(filters) > 0 {
			for key := range filters {
				js := filters.GetJSONObject(key)
				value := js.GetString(`value`)
				filter := strings.ToUpper(js.GetString(`filter`))
				switch filter {
				case `LIKE`:
					fallthrough
				case `FULLTEXT`:
					fallthrough
				case `>`:
					fallthrough
				case `>=`:
					fallthrough
				case `<`:
					fallthrough
				case `<=`:
					builder.WhereOp(key, filter)
					values = append(values, value)
				case `date`:
					builder.WhereOp(key, `>=`)
					values = append(values, value)

					time, _ := time.Parse(`2006-01-02`, value)
					time = time.AddDate(0, 0, 1)
					builder.WhereOp(key, `<`)
					values = append(values, time.Format(`2006-01-02`))
				default:
					builder.Where(key)
					values = append(values, value)
				}

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
				builder.Order(active, sort.GetStringOr(`direction`, `asc`))
			}
		}
		// Example:
		// {
		//   "page": {
		// 	   "location": 2,
		// 	   "length": 100
		//   }
		// }
		if page := js.GetJSONObject(`page`); page != nil {
			length := page.GetInt(`length`)
			location := page.GetInt(`location`)
			if length > 0 && location > 0 {
				location = (location - 1) * length
				builder.Limit(location, length)
			}
		}
	}

	tx, e := ctx.Tx()
	if e != nil {
		ctx.logger().E(e)
		return nil, errors.New(`Database connection failed`)
	}

	switch q.qType {
	case queryTypeSelect:
		sql := q.rawQuery
		if builder != nil {
			if builder.LimitLength() == 0 {
				builder.Limit(builder.LimitStart(), 1000)
			}
			sql = builder.ToSQL()
		}
		data, err = tx.Select(sql, values...)
	case queryTypeGet:
		sql := q.rawQuery
		if builder != nil {
			sql = builder.ToSQL()
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
		data, err = tx.Exec(q.rawQuery, values...)
	}
	if err != nil {
		if db.ErrorDuplicate(err) {
			return nil, errors.New(`duplicate entry`)
		}
		ctx.logger().E(err)
		return nil, errExecutingQuery
	}
	switch q.qType {
	case queryTypeInsert:
		if id, e := data.(*db.Result).LastInsertID(); e == nil {
			data = id
		}
	case queryTypeUpdate:
		fallthrough
	case queryTypeDelete:
		if rows, e := data.(*db.Result).RowsAffected(); e == nil {
			data = rows
		}
	}
	return data, nil
}

func (q *actionQuery) populateValues(ctx *context, item interface{}) ([]interface{}, error) {
	values := []interface{}{}
	for _, param := range q.qParams {
		if strings.HasPrefix(param, `$session.`) {
			val := ctx.sess.GetString(param[9:])
			values = append(values, val)
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

func (q *actionQuery) execute(ctx *context) error {
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
				if recs, ok := r.([]db.Resultset); ok {
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
				if recs, ok := r.([]db.Resultset); ok {
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

func newQueryAction(query, params string) (*actionQuery, error) {
	act := &actionQuery{rawQuery: query}

	str := strings.SplitN(query, ` `, 2)
	queryType := strings.ToUpper(str[0])
	switch queryType {
	case `SELECT`:
		act.builder = db.ParseQuery(query)
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
