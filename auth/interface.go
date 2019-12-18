package auth

import (
	"net/http"

	"gitlab.com/tuxer/go-db"
)

const (
	//TypeJWT ...
	TypeJWT = `jwt`
)

//Interface ...
type Interface interface {
	Type() string
	Authenticate(*db.Tx, *http.Request) error
}
