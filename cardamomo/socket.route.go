package cardamomo

import (
  "fmt"
	"net/http"

  "golang.org/x/net/websocket"
)

type SocketRoute struct {
  pattern string
  callback SockFunc
  clients []*SocketClient
}

func NewSocketRoute(pattern string, callback SockFunc) SocketRoute {
  return SocketRoute{pattern: pattern, callback: callback}
}

func (sr *SocketRoute) Listen() {
  fmt.Printf("\n\nSocket listen on pattern: %s\n\n", sr.pattern)
  onConnected := func(ws *websocket.Conn) {
    fmt.Printf("\n\nClient!\n\n")
    client := NewSocketClient(ws)
    sr.clients = append(sr.clients, &client)
    client.Listen()
  }
  http.Handle(sr.pattern, websocket.Handler(onConnected))
}
