package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type rewriteMap struct {
	regex       *regexp.Regexp
	replacePath string
}

type Proxy struct {
	client     *fasthttp.HostClient
	rewriteMap []rewriteMap
}

func (p *Proxy) translate(path string) (string, bool) {
	for _, rw := range p.rewriteMap {
		if rw.regex != nil {
			matches := rw.regex.FindStringSubmatch(path)
			if matches == nil {
				return ``, false
			}
			newPath := rw.replacePath
			for key, val := range matches {
				newPath = strings.ReplaceAll(newPath, fmt.Sprintf(`$%d`, key), val)
			}
			return newPath, true
		}
	}
	return ``, false
}

func prepareRequest(req *fasthttp.Request) {
	req.Header.Del("Connection")
}

func postprocessResponse(resp *fasthttp.Response) {
	resp.Header.Del("Connection")
}

func (p *Proxy) execute(s *Server, fastCtx *fasthttp.RequestCtx, newPath string) error {
	req, resp := &fastCtx.Request, &fastCtx.Response
	prepareRequest(req)
	query := req.URI().QueryString()
	req.SetRequestURI(newPath)
	if query != nil {
		req.URI().SetQueryString(string(query))
	}
	if e := p.client.DoTimeout(req, resp, 60*time.Second); e != nil {
		return e
	}
	postprocessResponse(resp)
	return nil
}

func (p *Proxy) Rewrite(regexPath, replacePath string) (*Proxy, error) {
	regex, e := regexp.Compile(regexPath)
	if e != nil {
		return p, e
	}
	p.rewriteMap = append(p.rewriteMap, rewriteMap{regex, replacePath})
	return p, nil
}
