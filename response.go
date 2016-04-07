package cardamomo

import (
  "io"
  "net/http"
  "encoding/json"
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
  http.ServeFile(r.Writer, r.httprequest, view)
}
