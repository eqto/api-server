package api

type logger struct {
	D func(...interface{})
	I func(...interface{})
	W func(...interface{})
	E func(...interface{})
}
