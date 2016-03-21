package cardamomo

import (
  _"fmt"
	_"net/http"
  _"strings"
)

type Route struct {
  method string
  pattern string
  patternRegex string
  params map[string]string
  callback ReqFunc
}

func NewRoute(method string, pattern string, callback ReqFunc) Route {
  /*http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
    if( strings.ToLower(req.Method) == strings.ToLower(method) ) {
      fmt.Printf("\n %s: %s \n", req.Method, pattern);
      request := NewRequest(req)
      response := NewResponse(w)
      callback(request, response)
    }
  })*/

  params := make(map[string]string)

  return Route{method: method, pattern: pattern, patternRegex: "", params: params, callback: callback}
}
