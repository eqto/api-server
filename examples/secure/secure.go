package main

import (
	"errors"
	"log"

	"github.com/eqto/api-server"
	"github.com/eqto/go-json"
)

func main() {
	s := api.New()
	s.SetProduction() //remove stacktrace and debug log

	s.AddAuthMiddleware(Auth)
	s.AddFuncRoute(Home, true)

	if e := s.Serve(8000); e != nil {
		log.Println(e)
	}

}

// Auth ..
func Auth(ctx api.RequestCtx) error {
	js, _ := json.Parse(ctx.Body())

	username := js.GetString(`username`)
	password := js.GetString(`password`)

	if username != `admin` && password != `admin` {
		return errors.New(`you shall not pass`)
	}
	return nil
}

// Home this endpoint will executed after AuthMiddleware pass
func Home(ctx *api.Context) (interface{}, error) {
	return `welcome`, nil
}
