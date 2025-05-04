package cardamomo

import (
  "net/http"
  "mime/multipart"
  "time"
  "strings"
  "io"
  "os"
  "encoding/json"
	"regexp"
)

type Request struct {
  w http.ResponseWriter
  httprequest *http.Request
  jsonparams JSONC
  route *Route
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

  return Request{w: w, httprequest: req, jsonparams: jsonparams, route: route}
}

func (r *Request) OriginalRequest() *http.Request {
  return r.httprequest
}

func (r *Request) GetParam(key string, defaultValue string) string {
   reg, _ := regexp.Compile(r.route.patternRegex)
   if reg.MatchString(r.httprequest.URL.Path) {
     urlparams := reg.FindStringSubmatch(r.httprequest.URL.Path)

     params := make(map[string]string)
     index := 1
     for _, param := range r.route.paramsOrder {
       params[param] = urlparams[index]

       if param == key {
         return urlparams[index]
       }

       index += 1
     }
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

func (r *Request) SetCookie(key string, value string, path string, domain string, expire time.Time, maxage int, secure bool, httponly bool, samesite http.SameSite) {
  cookie := &http.Cookie{
		Name:     key,
		Value:    value,
		Path:     path,
		Domain:   domain,
		Expires:  expire,
		MaxAge:   maxage,
		Secure:   secure,
		HttpOnly: httponly,
		SameSite: samesite,
	}
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
  r.SetCookie(key, "", path, domain, expire, 0, false, false, http.SameSiteStrictMode)
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
