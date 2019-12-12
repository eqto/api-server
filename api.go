package api

//Start ...
// func Start() (*Server, error) {
// 	srv := new(Server)
// 	// e := srv.Start()
// 	return srv, e
// }

type (
	// MiddlewareFunc func(HandlerFunc) HandlerFunc

	//MiddlewareFunc ...
	MiddlewareFunc func(Context) error
	//HandlerFunc ...
	// HandlerFunc func(Context) error
)
