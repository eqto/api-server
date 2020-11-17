package api

//RouteAuthenticator ..
type RouteAuthenticator func(ctx Ctx) error

//Middleware ..
type Middleware func(ctx Context) error

//RespMiddleware ...
type RespMiddleware func(req Request, resp Response)
