package apims

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	errExecutingQuery   = errors.New(`error executing query: %s`)
)

type queryAction struct {
	Action

	rawQuery  string
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

func (q *queryAction) executeItem(ctx *context, values []interface{}) (interface{}, error) {
	var data interface{}
	var err error

	switch q.qType {
	case queryTypeSelect:
		if q.qBuilder != nil { //run sql function
			data, err = ctx.tx.Select(q.qBuilder.ToSQL(), values...)
			if data == nil {
				data = []db.Resultset{}
			}
		}
	case queryTypeGet:
		if q.qBuilder != nil { //run sql function
			res, e := ctx.tx.Get(q.qBuilder.ToSQL(), values...)
			if e != nil {
				err = e
			} else if res != nil {
				data = res
			}
		}
	case queryTypeUpdate:
		data, err = ctx.tx.Exec(q.rawQuery, values...)
	case queryTypeInsert:

	}
	if err != nil {
		return nil, fmt.Errorf(errExecutingQuery.Error(), q.rawQuery)
	}
	return data, nil
}

func (q *queryAction) populateValues(ctx *context, item json.Object) ([]interface{}, error) {
	values := []interface{}{}
	for _, param := range q.qParams {
		if strings.HasPrefix(param, q.arrayName+`[`) && strings.HasSuffix(param, `]`) {
			val := item.Get(param[len(q.arrayName)+1 : len(param)-1])
			values = append(values, val)
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

func (q *queryAction) execute(ctx *context) (interface{}, error) {
	if q.arrayName != `` { //execute array
		array := ctx.req.jsonBody.GetArray(q.arrayName)

		result := []interface{}{}
		for _, obj := range array {
			values, e := q.populateValues(ctx, obj)
			if e != nil {
				return nil, e
			}
			r, e := q.executeItem(ctx, values)
			if e != nil {
				return nil, e
			}
			result = append(result, r)
		}
		return result, nil
	}
	values, e := q.populateValues(ctx, nil)
	if e != nil {
		return nil, e
	}
	return q.executeItem(ctx, values)
}

func newQueryAction(query, property, params string) (*queryAction, error) {
	act := &queryAction{rawQuery: query, qProperty: property}

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
