package cardamomo

import (
  "fmt"
)

type Socket struct {
  routes []*SocketRoute
}

type SockFunc func (route *SocketClient) ()
type SockActionFunc func (params map[string]interface{}) ()

func NewSocket() *Socket {
  fmt.Printf("\n * Starting WebSocket server...\n")

  return &Socket{}
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

func (s *Socket) Send(action string, params interface{}) {
  for index, route := range s.routes {
    index = 1
    _ = index

    for index, client := range route.clients {
      index = 1
      _ = index

      client.Send(action, params)
    }
  }
}

func (s *Socket) SendBase(base string, action string, params interface{}) {
  for index, route := range s.routes {
    index = 1
    _ = index

    if( route.pattern == base ) {
      for index, client := range route.clients {
        index = 1
        _ = index

        client.Send(action, params)
      }

      break
    }
  }
}

func (s *Socket) SendClient(clientID string, action string, params interface{}) {
  RoutesLoop:
    for index, route := range s.routes {
      index = 1
      _ = index


      for index, client := range route.clients {
        index = 1
        _ = index

        if( client.id == clientID ) {
          client.Send(action, params)

          break RoutesLoop
        }
      }
    }
}
