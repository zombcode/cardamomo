package cardamomo

import (
  "net/http"
  "mime/multipart"
  "time"
  "strings"
  "io"
  "os"
  "encoding/json"
)

type Request struct {
  w http.ResponseWriter
  httprequest *http.Request
  params map[string]string
  jsonparams JSONC
}

func NewRequest(w http.ResponseWriter, req *http.Request, route *Route) Request {
  req.ParseForm()

  // JSON params
  jsonparams := JSONC{}

  if(strings.Contains(req.Header.Get("Content-Type"), "application/json")) {
    decoder := json.NewDecoder(req.Body)
    var t JSONC
    err := decoder.Decode(&t)
    if err == nil {
      jsonparams = t
    }
  }

  // Multipart params
  if(strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")) {
    req.ParseMultipartForm(32 << 20)
  }

  // Route params
  params := make(map[string]string)
  if( route != nil ) {
    params = route.params
  }

  return Request{w: w, httprequest: req, params: params, jsonparams: jsonparams}
}

func (r *Request) OriginalRequest() *http.Request {
  return r.httprequest
}

func (r *Request) GetParam(key string, defaultValue string) string {
   if param, ok := r.params[key]; ok {
     return param
   }

   param := r.httprequest.FormValue(key)
   if param != "" {
     return param
   }

   if paraminterface, ok := r.jsonparams[key]; ok {
     param, err := json.Marshal(paraminterface)
     if(err == nil) {
       return string(param)
     }
   }

   return defaultValue
}

func (r *Request) GetParamJSON(key string, defaultValue interface{}) interface{} {
   if param, ok := r.jsonparams[key]; ok {
    return param
   }

   return defaultValue
}

func (r * Request) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
  return r.httprequest.FormFile(key)
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

func (r *Request) MoveUploadedFile(fileKey string, destinationPath string) error {
  f1, _, err := r.GetFile(fileKey)
  if err != nil {
    return err
  }
  defer f1.Close()

  f2, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE, 0666)
  if err != nil {
    return err
  }
  defer f2.Close()

  n, err := io.Copy(f2, f1)
  _ = n

  if err != nil {
    return err
  }

  return err
}
