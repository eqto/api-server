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
		val := ctx.req.get(param)
		if val == nil {
			return nil, fmt.Errorf(errMissingParameter.Error(), param)
		}
		values = append(values, val)
	}

	var data interface{}

	switch q.qType {
	case queryTypeSelect:
		if q.qBuilder != nil { //run sql function
			rs, e := ctx.tx.Select(q.q, values...)
			if e != nil {
				return nil, fmt.Errorf(errExecutingQuery.Error(), q.q)
			}
			if rs == nil {
				rs = []db.Resultset{}
			}
			data = rs
		}
	case queryTypeGet:
		if q.qBuilder != nil { //run sql function
			rs, e := ctx.tx.Get(q.q, values...)
			if e != nil {
				return nil, fmt.Errorf(errExecutingQuery.Error(), q.q)
			}
			if rs != nil {
				data = rs
			}
		}

	case queryTypeInsert:

	}
	return data, nil
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
