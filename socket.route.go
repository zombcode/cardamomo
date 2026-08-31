package cardamomo

import (
  "fmt"
	"net/http"
  b64 "encoding/base64"
  "golang.org/x/net/websocket"
  "sync"
)

type SocketRoute struct {
  socket *Socket
  pattern string
  callback SockFunc
  clients *sync.Map
}

func NewSocketRoute(s *Socket, pattern string, callback SockFunc) SocketRoute {
  return SocketRoute{socket: s, pattern: pattern, callback: callback, clients: &sync.Map{}}
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

    sr.clients.Store(client.id, &client)
    sr.callback(&client)

    client.Listen()
  }
  http.Handle(sr.pattern, websocket.Handler(onConnected))
}
