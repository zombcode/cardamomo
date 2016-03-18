package cardamomo

import (
  "fmt"
)

type Socket struct {
  port string
  routes []*SocketRoute
}

type SockFunc func () ()

func NewSocket(port string) Socket {
  fmt.Printf("\n\nStarting TCP server at port:%s\n\n", port)

  return Socket{port: port}
}

func (s *Socket) SocketBase(pattern string, callback SockFunc) {
  s.addBase(pattern, callback);
}

func (s *Socket) addBase(pattern string, callback SockFunc) SocketRoute {
  route := NewSocketRoute(pattern, callback)
  route.Listen()
  s.routes = append(s.routes, &route)

  return route
}
