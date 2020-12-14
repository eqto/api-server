package api

//RouteAuthenticator ..
type RouteAuthenticator func(ctx Context) error

//middlewareContainer
type middlewareContainer struct {
	f      func(ctx Context) error
	secure bool
}
