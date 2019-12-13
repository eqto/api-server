package api

import (
	"gitlab.com/tuxer/go-json"
)

//Request ...
type Response struct {
	json.Object
}

// func parseRequest(r *http.Request) *Request {
// 	req := new(Request)
// 	switch r.Method {
// 	case http.MethodPost:
// 		if body, e := ioutil.ReadAll(r.Body); e == nil {
// 			req.Object = json.Parse(body)
// 		}
// 	case http.MethodGet:
// 		js := make(json.Object)
// 		for key := range r.URL.Query() {
// 			js.Put(key, r.URL.Query().Get(key))
// 		}
// 		req.Object = js
// 	}
// 	return req
// }
