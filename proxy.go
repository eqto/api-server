package api

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type proxy struct {
	path  string
	dest  string
	regex *regexp.Regexp
}

func (p *proxy) match(path string) bool {
	if p.regex == nil {
		return false
	}
	return p.regex.MatchString(path)
}

func (p *proxy) execute(s *Server, ctx *fasthttp.RequestCtx) (Response, error) {
	u, e := url.Parse(string(ctx.RequestURI()))
	if e != nil {
		return newResponseError(StatusInternalServerError, e)
	}

	dest := p.dest
	if strings.HasSuffix(dest, `/`) {
		dest = dest[:len(dest)-1]
	}
	dest = dest + u.Path
	if len(u.RawQuery) > 0 {
		dest = dest + `?` + u.RawQuery
	}
	s.logD(`Proxy dest:`, dest)

	httpReq := fasthttp.AcquireRequest()
	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(httpReq)
	defer fasthttp.ReleaseResponse(httpResp)
	ctx.Request.Header.CopyTo(&httpReq.Header)

	httpReq.SetRequestURI(dest)
	httpReq.Header.SetMethod(string(ctx.Method()))
	httpReq.SetBody(ctx.Request.Body())

	client := &fasthttp.Client{}

	if e := client.DoTimeout(httpReq, httpResp, 120*time.Second); e != nil {
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
	return p, nil
}
