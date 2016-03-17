package cardamomo

import (
  "io"
  "net/http"
  "fmt"
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
