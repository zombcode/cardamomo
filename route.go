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
  paramsOrder []string
  callback ReqFunc
}

func NewRoute(method string, pattern string, callback ReqFunc) Route {
  params := make(map[string]string)
  paramsOrder := make([]string, 0)

  return Route{method: method, pattern: pattern, patternRegex: "", params: params, paramsOrder: paramsOrder, callback: callback}
}
