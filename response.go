package cardamomo

import (
  "io"
  "net/http"
  "encoding/json"
  "html/template"
  "strings"
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

func (r *Response) Render(view string, data interface{}) {
  //http.ServeFile(r.Writer, r.httprequest, view)
  renderTemplate(r.Writer, view, data)
}

var templateMap = template.FuncMap{
	"Upper": func(s string) string {
		return strings.ToUpper(s)
	},
}
var templates = template.New("").Funcs(templateMap)

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
