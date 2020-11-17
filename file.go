package api

import (
	"os"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

type file struct {
	regex   *regexp.Regexp
	handler fasthttp.RequestHandler
	path    string
}

func (f *file) match(path string) bool {
	if f.regex == nil {
		return false
	}
	return f.regex.MatchString(path)
}

func newFile(path, dest, redirectTo string) (file, error) {
	if !strings.HasPrefix(path, `^`) {
		path = `^` + path
	}
	regex, e := regexp.Compile(path)

	fs := &fasthttp.FS{
		Root:               dest,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: false,
		Compress:           false,
		AcceptByteRange:    false,
	}
	fs.PathRewrite = func(ctx *fasthttp.RequestCtx) []byte {
		url := string(ctx.RequestURI())
		if _, e := os.Stat(dest + url); e == nil {
			return []byte(url)
		}
		if !strings.HasPrefix(redirectTo, `/`) {
			redirectTo = `/` + redirectTo
		}
		return []byte(redirectTo)
	}

	f := file{
		path:    path,
		regex:   regex,
		handler: fs.NewRequestHandler(),
	}
	if e != nil {
		return f, e
	}
	return f, nil
}
