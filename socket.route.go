package cardamomo

import (
  "fmt"
	"net/http"
  b64 "encoding/base64"
  "golang.org/x/net/websocket"
)

type SocketRoute struct {
  socket *Socket
  pattern string
  callback SockFunc
  clients map[string]*SocketClient
}

func NewSocketRoute(s *Socket, pattern string, callback SockFunc) SocketRoute {
  clients := make(map[string]*SocketClient)

  return SocketRoute{socket: s, pattern: pattern, callback: callback, clients: clients}
}

func (sr *SocketRoute) Listen() {
  fmt.Printf("   - Socket listen on pattern: %s\n", sr.pattern)
  onConnected := func(ws *websocket.Conn) {
    if sr.socket.clustered == true && sr.pattern == "/cardacluster" {
      receivedPassword, _ := b64.StdEncoding.DecodeString(ws.Request().Header.Get("Cardamomo-Cluster-Password"))
      if( string(receivedPassword) != sr.socket.clusterParams.Password ) {
        websocket.Message.Send(ws, "cardamomoinvalidpassword")
        ws.Close()
        return
      }
    }

    client := NewSocketClient(ws, sr)

    lock.RLock()
    defer lock.RUnlock()
    //sr.clients = append(sr.clients, &client)
    sr.clients[client.id] = &client
    sr.callback(&client)
    client.Listen()
  }
  http.Handle(sr.pattern, websocket.Handler(onConnected))
}
