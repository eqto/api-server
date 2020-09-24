package main

import (
	"errors"
	"log"

	"github.com/eqto/api-server"
)

func main() {
	s := api.New()
	// uncomment this line to change
	// s.NormalizeFunc(true)

	s.AddFunc(Hello)      //add endpoint /Hello
	s.AddFunc(HelloWorld) //add endpoint /HelloWorld or /hello_world

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

//HelloWorld endpoint http://host:port/HelloWorld if normalize false, or http://host:port/hello_world if normalize true
// output:
// {
//     "data": "this is error",
//     "message": "success",
//     "status": 500
// }
func HelloWorld(ctx api.Context) (interface{}, error) {
	return nil, errors.New(`this is error`)
}
