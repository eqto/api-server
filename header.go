package api

import (
	"github.com/valyala/fasthttp"
)

//RequestHeader ...
type RequestHeader struct {
	httpHeader *fasthttp.RequestHeader
}
