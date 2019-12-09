package api

import (
	"regexp"
	"strings"

	"gitlab.com/tuxer/go-db"

	"gitlab.com/tuxer/go-json"
)

//Parameter ...
type Parameter struct {
	query     string
	queryType string
	qb        *db.QueryBuilder
	params    []string
	output    string

	children []Parameter
}

func (p *Parameter) isArray() bool {
	return p.children != nil
}

func (p *Parameter) putOutput(req *Request, resp *json.Object, value interface{}) {
	if strings.HasPrefix(p.output, `$`) {
		req.Put(p.output, value)
	} else {
		resp.Put(p.output, value)
	}
}

func newParameter(js json.Object) *Parameter {
	p := &Parameter{
		query:     strings.TrimSpace(js.GetString(`query`)),
		queryType: strings.TrimSpace(js.GetString(`query_type`)),
		output:    strings.TrimSpace(js.GetStringOr(`output`, `data`)),
	}
	if p.queryType == `` {
		str := strings.SplitN(p.query, ` `, 2)
		p.queryType = strings.ToUpper(str[0])
		if p.queryType == `SELECT` && regexp.MustCompile(`LIMIT.*\s+1$`).MatchString(strings.ToUpper(p.query)) {
			p.queryType = `GET`
		}
	}

	p.queryType = strings.ToUpper(p.queryType)
	p.qb = db.ParseQuery(p.query)

	if params := js.GetString(`params`); params != `` {
		split := strings.Split(params, `,`)
		for key, val := range split {
			split[key] = strings.TrimSpace(val)
		}
		p.params = split
	}

	return p
}
