package cardamomo

import (
  "io"
  "net/http"
  "fmt"
  "encoding/json"
)

type Response struct {
  writer http.ResponseWriter
}

func NewResponse(w http.ResponseWriter) Response {
  return Response{writer: w}
}

func (r *Response) Send(m string) {
  fmt.Printf("\n RES: %s \n", m);
  io.WriteString(r.writer, m)
}

func (r *Response) SendJSON(data interface{}) {
  result, _ := json.Marshal(data)

  io.WriteString(r.writer, string(result))
}
