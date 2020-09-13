package main

import (
	"errors"
	"log"

	"github.com/eqto/api-server"
)

func main() {
	s := api.New()

	s.AddFunc(Hello)
	s.AddFunc(HelloError)

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
func Hello(f api.Context) (interface{}, error) {
	return `hello world`, nil
}

//HelloError endpoint http://host:port/HelloError
// output:
// {
//     "data": "test",
//     "message": "success",
//     "status": 0
// }
func HelloError(f api.Context) (interface{}, error) {
	return nil, errors.New(`this is error`)
}
