package cardamomo

import (
  "fmt"
)

type Socket struct {
  routes []*SocketRoute
}

type SockFunc func (route *SocketClient) ()
type SockActionFunc func (params map[string]interface{}) ()

func NewSocket() Socket {
  fmt.Printf("\n\nStarting WebSocket server\n\n")

  return Socket{}
}

func (s *Socket) OnSocketBase(pattern string, callback SockFunc) {
  s.addBase(pattern, callback);
}

func (s *Socket) addBase(pattern string, callback SockFunc) SocketRoute {
  route := NewSocketRoute(pattern, callback)
  route.Listen()
  s.routes = append(s.routes, &route)

  return route
}
