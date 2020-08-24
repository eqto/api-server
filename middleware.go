package apims

//Middleware ..
type Middleware func(ctx RequestCtx) error

type middleware struct {
	f      Middleware
	isAuth bool
}
