package api

//Middleware ..
type Middleware func(ctx RequestCtx) error

//RespMiddleware ...
type RespMiddleware func(resp Response)
