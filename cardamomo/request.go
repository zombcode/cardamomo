package cardamomo

import (
  "net/http"
)

type Request struct {
  httprequest *http.Request
  params map[string]string
}

func (r *Request) GetParam(key string) string {
   if param, ok := r.params[key]; ok {
     return param
   } else {
     return r.httprequest.FormValue(key)
   }
}

func NewRequest(req *http.Request, route *Route) Request {
  req.ParseForm()
  return Request{httprequest: req, params: route.params}
}
