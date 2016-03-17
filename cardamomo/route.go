package cardamomo

import (
  "fmt"
	"net/http"
  "strings"
)

type Route struct {
  method string
  pattern string
  callback ReqFunc
}

func NewRoute(method string, pattern string, callback ReqFunc) Route {
  http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
    if( strings.ToLower(req.Method) == strings.ToLower(method) ) {
      fmt.Printf("\n %s: %s \n", req.Method, pattern);
      response := NewResponse(w)
      callback(response)
    }
  })

  return Route{method: method, pattern: pattern, callback: callback}
}
