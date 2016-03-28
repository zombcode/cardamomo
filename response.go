package cardamomo

import (
  "io"
  "net/http"
  "encoding/json"
)

type Response struct {
  Writer http.ResponseWriter
}

func NewResponse(w http.ResponseWriter) Response {
  return Response{Writer: w}
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

}
