package cardamomo

import (
  "net/http"
)

type Request struct {
  httprequest *http.Request
}

func (r *Request) GetParam(key string) string {
   return r.httprequest.FormValue(key)
}

func NewRequest(req *http.Request) Request {
  req.ParseForm()
  return Request{httprequest: req}
}
