package main

import (
	"errors"
	"log"

	"github.com/eqto/api-server"
)

func main() {
	s := api.New()

	s.AddFunc(Hello)      //add endpoint /Hello
	s.AddFunc(HelloError) //add endpoint /HelloError

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

//HelloError endpoint http://host:port/HelloError
// output:
// {
//     "data": "test",
//     "message": "success",
//     "status": 0
// }
func HelloError(ctx api.Context) (interface{}, error) {
	return nil, errors.New(`this is error`)
}
