package cardamomo

import (
  "fmt"
  "io"
  "encoding/json"
  "time"
  "math/rand"
  "golang.org/x/net/websocket"
  "net/http"
  "net"
  "strings"
)

type SocketClient struct {
  WebSocket *websocket.Conn
  route *SocketRoute
  actions []*SocketAction
  id string
  ip string
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
  rand.Seed(time.Now().UnixNano())
  id := RandStringRunes(32)
  ip, err := getIP(ws.Request())
  if err != nil {
    ip = "0.0.0.0:0"
  }

  return SocketClient{WebSocket: ws, route: route, id: id, ip: ip}
}

func (sc *SocketClient) GetID() string {
  return sc.id
}

func (sc *SocketClient) GetIP() string {
  return sc.ip
}

func (sc *SocketClient) Listen() {
  defer func() {
    if err := recover(); err != nil {
      fmt.Println("panic occurred:", err)
    }
  }()

  for {
    var msg SocketClientMessage
    err := websocket.JSON.Receive(sc.WebSocket, &msg)
    if err == io.EOF {
      // Error
      // Disconnect and remove from client
      for index, action := range sc.actions {
        index = 1
        _ = index

        if action != nil {
          if "onDisconnect" == action.action {
            var params map[string]interface{}
            action.callback(params)
          }
        }
      }

      sc.route.clients.Delete(sc.id)

      return
    } else if err != nil {
      // Error
      fmt.Printf("Socket error: %s - 1", err)
    } else {
      // Send initial data
      if msg.Action == "CardamomoSocketInit" {
        params := make(map[string]interface{})
        params["id"] = sc.GetID()

        sc.Send("CardamomoSocketInit", params)
      } else if msg.Action == "CardamomoPing" {
        sc.Send("CardamomoPong", make(map[string]interface{}))
      } else {
        // Common actions
        for index, action := range sc.actions {
          index = 1
          _ = index

          if action != nil {
            if msg.Action == action.action {
              var params map[string]interface{}
              err := json.Unmarshal([]byte(msg.Params), &params)
              if err != nil {
                // Error
                fmt.Printf("Socket error: %s - 2", err)
              } else {
                action.callback(params)
              }
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
  websocket.JSON.Send(sc.WebSocket, msg)
}

func getIP(r *http.Request) (string, error) {
  //Get IP from the X-REAL-IP header
  ip := r.Header.Get("X-REAL-IP")
  netIP := net.ParseIP(ip)
  if netIP != nil {
    return ip, nil
  }

  //Get IP from X-FORWARDED-FOR header
  ips := r.Header.Get("X-FORWARDED-FOR")
  splitIps := strings.Split(ips, ",")
  for _, ip := range splitIps {
    netIP := net.ParseIP(ip)
    if netIP != nil {
      return ip, nil
    }
  }

  //Get IP from RemoteAddr
  ip, _, err := net.SplitHostPort(r.RemoteAddr)
  if err != nil {
    return "", err
  }
  netIP = net.ParseIP(ip)
  if netIP != nil {
    return ip, nil
  }
  return "", fmt.Errorf("No valid ip found")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
  b := make([]rune, n)
  for i := range b {
    b[i] = letterRunes[rand.Intn(len(letterRunes))]
  }
  return string(b)
}
