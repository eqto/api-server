package api

import (
	"regexp"
	"strings"

	"gitlab.com/tuxer/go-db"

	"gitlab.com/tuxer/go-json"
)

//Parameter ...
type Route struct {
	query     string
	queryType string
	qb        *db.QueryBuilder
	params    []string
	output    string
	secure    bool

	children []Route
}

func (r *Route) isArray() bool {
	return r.children != nil
}

func (r *Route) putOutput(req *Request, resp *json.Object, value interface{}) {
	if strings.HasPrefix(r.output, `$`) {
		req.Put(r.output, value)
	} else {
		resp.Put(r.output, value)
	}
}

func newRoute(js json.Object) *Route {
	r := &Route{
		query:     strings.TrimSpace(js.GetString(`query`)),
		queryType: strings.TrimSpace(js.GetString(`query_type`)),
		output:    strings.TrimSpace(js.GetStringOr(`output`, `data`)),
		secure:    js.GetBoolean(`secure`),
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

	if params := js.GetString(`params`); params != `` {
		split := strings.Split(params, `,`)
		for key, val := range split {
			split[key] = strings.TrimSpace(val)
		}
		r.params = split
	}

	return r
}
