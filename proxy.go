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

func (p *proxy) execute(s *Server, ctx *context) error {
	prepareRequest(ctx.req.httpReq)
	if e := p.client.DoTimeout(ctx.req.httpReq, ctx.resp.httpResp, 60*time.Second); e != nil {
		return e
	}
	postprocessResponse(ctx.resp.httpResp)
	return nil

	//TODO used when responsemiddleware implemented
	// httpResp := fasthttp.AcquireResponse()
	// defer fasthttp.ReleaseResponse(httpResp)
	// if e := p.client.DoTimeout(ctx.req.httpReq, httpResp, 60*time.Second); e != nil {
	// 	ctx.resp.setError(StatusBadGateway, e)
	// 	return e
	// }
	// httpResp.CopyTo(ctx.resp.httpResp)
	// return nil
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
