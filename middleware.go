package api

//RouteAuthenticator ..
type RouteAuthenticator func(ctx RequestCtx) error

//Middleware ..
type Middleware func(ctx RequestCtx) error

//RespMiddleware ...
type RespMiddleware func(req Request, resp Response)
