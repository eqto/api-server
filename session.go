package apims

import "github.com/eqto/go-json"

type Session interface {
}

type session struct {
	Session
	json.Object
}
