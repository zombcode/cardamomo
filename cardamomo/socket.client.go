package cardamomo

import (
  "io"
  "golang.org/x/net/websocket"
)

type SocketClient struct {
  ws *websocket.Conn
}

func NewSocketClient(ws *websocket.Conn) SocketClient {
  return SocketClient{ws: ws}
}

func (sc *SocketClient) Listen() {
  for {
		select {
  		// read data from websocket connection
  		default:
  			var msg Message
  			err := websocket.JSON.Receive(sc.ws, &msg)
  			fmt.Printf("\n\nMESSAGE: %s\n\n", msg)
		}
	}
}
