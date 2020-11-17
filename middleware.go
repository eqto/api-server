package api

//RouteAuthenticator ..
type RouteAuthenticator func(ctx Context) error

//Middleware ..
type Middleware func(ctx Context) error

//RespMiddleware ...
type RespMiddleware func(req Request, resp Response)
