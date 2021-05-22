package api

import (
	"regexp"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type proxy struct {
	path   string
	dest   string
	regex  *regexp.Regexp
	client *fasthttp.HostClient
}

func (p *proxy) match(path string) bool {
	if p.regex == nil {
		return false
	}
	return p.regex.MatchString(path)
}

func prepareRequest(req *fasthttp.Request) {
	req.Header.Del("Connection")
}

func postprocessResponse(resp *fasthttp.Response) {
	resp.Header.Del("Connection")
}

func (p *proxy) execute(s *Server, fastCtx *fasthttp.RequestCtx) error {
	req, resp := &fastCtx.Request, &fastCtx.Response

	prepareRequest(req)
	if e := p.client.DoTimeout(req, resp, 60*time.Second); e != nil {
		return e
	}
	postprocessResponse(resp)
	return nil
}

func newProxy(path, dest string) (proxy, error) {
	if !strings.HasPrefix(path, `^`) {
		path = `^` + path
	}
	regex, e := regexp.Compile(path)

	p := proxy{path: path, dest: dest}
	if e != nil {
		return p, e
	}
	p.regex = regex
	p.client = &fasthttp.HostClient{Addr: dest}
	return p, nil
}
