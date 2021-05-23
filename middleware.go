package api

//Middleware ..
type Middleware interface {
	//ForGroup set name for this middleware, this middleware only used by route that requiring the same name
	ForGroup(string) Middleware
	Secure() Middleware
}

//middlewareContainer
type middlewareContainer struct {
	Middleware
	f      func(Context) error
	secure bool
	group  string
}

func (m *middlewareContainer) ForGroup(name string) Middleware {
	m.group = name
	return m
}

func (m *middlewareContainer) Secure() Middleware {
	m.secure = true
	return m
}
