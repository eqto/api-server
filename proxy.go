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

func (p *proxy) execute(s *Server, ctx *fasthttp.RequestCtx) (Response, error) {
	httpReq := &ctx.Request
	if s.respMiddleware == nil || len(s.respMiddleware) == 0 {
		httpResp := &ctx.Response
		prepareRequest(httpReq)
		if e := p.client.DoTimeout(httpReq, httpResp, 60*time.Second); e != nil {
			return nil, nil
		}
		postprocessResponse(httpResp)
		return nil, nil
	}
	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)
	if e := p.client.DoTimeout(httpReq, httpResp, 60*time.Second); e != nil {
		return newResponseError(StatusBadGateway, e)
	}
	resp := newResponse(StatusOK)
	resp.status = httpResp.StatusCode()
	header := resp.Header()
	httpResp.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})
	resp.header = header
	resp.rawBody = httpResp.Body()
	return resp, nil
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
