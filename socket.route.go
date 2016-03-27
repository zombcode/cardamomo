package cardamomo

import (
  "fmt"
	"net/http"

  "golang.org/x/net/websocket"
)

type SocketRoute struct {
  pattern string
  callback SockFunc
  clients map[string]*SocketClient
}

func NewSocketRoute(pattern string, callback SockFunc) SocketRoute {
  clients := make(map[string]*SocketClient)

  return SocketRoute{pattern: pattern, callback: callback, clients: clients}
}

func (sr *SocketRoute) Listen() {
  fmt.Printf("   - Socket listen on pattern: %s\n", sr.pattern)
  onConnected := func(ws *websocket.Conn) {
    client := NewSocketClient(ws, sr)
    //sr.clients = append(sr.clients, &client)
    sr.clients[client.id] = &client
    sr.callback(&client)
    client.Listen()
  }
  http.Handle(sr.pattern, websocket.Handler(onConnected))
}
