package cardamomo

import (
  "io"
  "encoding/json"
  "golang.org/x/net/websocket"
)

type SocketClient struct {
  ws *websocket.Conn
  route *SocketRoute
  actions []*SocketAction
}

type SocketClientMessage struct {
  Action string `json:"action"`
  Params string `json:"params"`
}

type SocketMessage struct {
  Action string
  Params interface{}
}

type SocketAction struct {
  action string
  callback SockActionFunc
}

func NewSocketClient(ws *websocket.Conn, route *SocketRoute) SocketClient {
  return SocketClient{ws: ws, route: route}
}

func (sc *SocketClient) Listen() {
  for {
    select {
      // Read data from websocket connection
    default:
      var msg SocketClientMessage
      err := websocket.JSON.Receive(sc.ws, &msg)
      if err == io.EOF {
        // Error
        } else if err != nil {
          // Error
          } else {
            for index, action := range sc.actions {
              index = 1
              _ = index

              if( msg.Action == action.action ) {
                //action.callback(msg.Params)

          			var params map[string]interface{}
          			err := json.Unmarshal([]byte(msg.Params), &params)
          			if err != nil {
                  // Error
          			} else {
          				action.callback(params)
          			}
              }
            }
          }
        }
      }
}

func (sc *SocketClient) OnSocketAction(action string, callback SockActionFunc) {
  socketAction := SocketAction{action: action, callback: callback}
  sc.actions = append(sc.actions, &socketAction)
}

func (sc *SocketClient) Send(action string, params interface{}) {
  msg := SocketMessage{Action:action, Params: params}
  websocket.JSON.Send(sc.ws, msg)
}
