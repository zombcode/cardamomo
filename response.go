package cardamomo

import (
  "io"
  "net/http"
  "encoding/json"
  "html/template"
  "strings"
  "io/ioutil"
  "log"
)

type Response struct {
  Writer http.ResponseWriter
  httprequest *http.Request
}

func NewResponse(w http.ResponseWriter, req *http.Request,) Response {
  return Response{Writer: w, httprequest: req}
}

func (r *Response) Send(m string) {
  io.WriteString(r.Writer, m)
}

type JSONC map[string]interface{}

func (r *Response) SendJSON(data interface{}) {
  result, _ := json.Marshal(data)
  r.Writer.Header().Set("Content-Type", "application/json")
  io.WriteString(r.Writer, string(result))
}

// Redirect
func (r *Response) Redirect(url string, code int) {
  http.Redirect(r.Writer, r.httprequest, url, code)
}

// Send file

func (r *Response) SendFile(path string) {
  http.ServeFile(r.Writer, r.httprequest, path)
}

// Render

var templateMap = template.FuncMap{
  "Upper": func(s string) string {
    return strings.ToUpper(s)
  },
}
var templates = template.New("").Funcs(templateMap)

func (r *Response) Render(path string, data interface{}) {

  bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicf("Unable to parse: path=%s, err=%s", path, err)
	}
	templates.New(path).Parse(string(bytes))

  err = templates.ExecuteTemplate(r.Writer, path, data)
	if err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
	}
}
