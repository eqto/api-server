package main

import (
	"errors"
	"log"

	"github.com/eqto/api-server"
)

func main() {
	s := api.New()

	s.AddFunc(Hello)      //add endpoint /Hello
	s.AddFunc(World) //add endpoint /World

	if e := s.Serve(8000); e != nil {
		log.Println(e)
	}
	log.Println(`Server stopped`)
}

//Hello endpoint http://host:port/Hello
// output:
// {
//     "data": "hello world",
//     "message": "success",
//     "status": 0
// }
func Hello(ctx api.Context) (interface{}, error) {
	return `hello world`, nil
}

//World endpoint http://host:port/World
// output:
// {
//     "data": "this is error",
//     "message": "success",
//     "status": 500
// }
func World(ctx api.Context) (interface{}, error) {
	return nil, errors.New(`this is error`)
}
