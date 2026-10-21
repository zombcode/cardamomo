package cardamomo

import (
  "fmt"
	"net/http"
  b64 "encoding/base64"
  "golang.org/x/net/websocket"
  "sync"
  "time"
)

type SocketRoute struct {
  socket *Socket
  pattern string
  callback SockFunc
  clients *sync.Map
  options SocketOptions
  limiter *connLimiter
}

func NewSocketRoute(s *Socket, pattern string, callback SockFunc) SocketRoute {
  return SocketRoute{socket: s, pattern: pattern, callback: callback, clients: &sync.Map{}}
}

// NewSocketRouteWithOptions is NewSocketRoute with the hardening options of
// SocketOptions applied to every connection of the base.
func NewSocketRouteWithOptions(s *Socket, pattern string, callback SockFunc, options SocketOptions) SocketRoute {
  route := NewSocketRoute(s, pattern, callback)
  route.options = options
  route.limiter = newConnLimiter(options.MaxConnsPerIP)
  return route
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

    if sr.options.MaxMessageBytes > 0 {
      ws.MaxPayloadBytes = sr.options.MaxMessageBytes
    }

    client := NewSocketClient(ws, sr)

    if sr.limiter != nil {
      key := client.ip
      if sr.options.ClientKey != nil {
        key = sr.options.ClientKey(ws.Request())
      }
      if !sr.limiter.acquire(key) {
        sr.options.reject(ws.Request(), "max-conns")
        ws.Close()
        return
      }
      defer sr.limiter.release(key)
    }

    if sr.options.AuthTimeout > 0 {
      // Closing the socket unblocks Receive in Listen, which then runs the
      // usual disconnect path. No deadline is set on the connection.
      timer := time.AfterFunc(sr.options.AuthTimeout, func() {
        if !client.IsAuthenticated() {
          sr.options.reject(ws.Request(), "auth-timeout")
          ws.Close()
        }
      })
      defer timer.Stop()
    }

    sr.clients.Store(client.id, &client)
    sr.callback(&client)

    client.Listen()
  }
  http.Handle(sr.pattern, sr.options.handler(onConnected))
}
