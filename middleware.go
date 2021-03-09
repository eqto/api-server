package api

//Middleware ..
type Middleware interface {
	//For set name for this middleware, this middleware only used by route that requiring the same name
	For(string) Middleware
	Secure() Middleware
}

//RouteAuthenticator ..
type RouteAuthenticator func(ctx Context) error

//middlewareContainer
type middlewareContainer struct {
	Middleware
	f      func(Context) error
	secure bool
	name   string
}

func (m *middlewareContainer) For(name string) Middleware {
	m.name = name
	return m
}

func (m *middlewareContainer) Secure() Middleware {
	m.secure = true
	return m
}
