package cardamomo

import (
  "net/http"
  "time"
)

type Request struct {
  w http.ResponseWriter
  httprequest *http.Request
  params map[string]string
}

func NewRequest(w http.ResponseWriter, req *http.Request, route *Route) Request {
  req.ParseForm()

  if( route != nil ) {
    return Request{w: w, httprequest: req, params: route.params}
  }

  return Request{w: w, httprequest: req}
}

func (r *Request) GetParam(key string, defaultValue string) string {
   if param, ok := r.params[key]; ok {
     return param
   }

   param := r.httprequest.FormValue(key)
   if param != "" {
     return param
   }

   return defaultValue
}

func (r *Request) SetCookie(key string, value string, path string, domain string, expire time.Time, maxage int, secure bool, httponly bool) {
  cookie := &http.Cookie{key, value, path, domain, expire, expire.Format(time.UnixDate), maxage, secure, httponly, key + "=" + value, []string{key + "=" + value}}
  http.SetCookie(r.w, cookie)
}

func (r *Request) GetCookie(key string, defaultValue string) string {
  cookie, err := r.httprequest.Cookie(key)
  if err == nil {
    return cookie.Value
  }

  return defaultValue
}

func (r *Request) DeleteCookie(key string, path string, domain string) {
  expire := time.Now().AddDate(0, 0, -1) // Expires yesterday!
  r.SetCookie(key, "", path, domain, expire, 0, false, false)
}
