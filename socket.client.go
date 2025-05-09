package cardamomo

import (
  "fmt"
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
  initialized bool
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

func (sc *SocketClient) IsInitialized() bool {
    return sc.initialized
}

func (sc *SocketClient) SetInitialized() {
    sc.initialized = true
}

func (sc *SocketClient) Listen() {
    defer func() {
      if err := recover(); err != nil {
        fmt.Println("Panic occurred on \"Listen\":", err)
      }
    }()
    
    go func() {
      ticker := time.NewTicker(30 * time.Second)
      defer ticker.Stop()

      for {
        <-ticker.C
        err := sc.Send("CardamomoPing", make(map[string]interface{}))
        if err != nil {
          fmt.Printf("Socket error: %s\n", err)

          for _, action := range sc.actions {
            if action != nil && action.action == "onDisconnect" {
              var params map[string]interface{}
              action.callback(params)
            }
          }
          
          sc.route.clients.Delete(sc.id)
          return
        }
      }
    }()

    for {
      var msg SocketClientMessage
      err := websocket.JSON.Receive(sc.WebSocket, &msg)
      if err != nil {
        fmt.Printf("Socket error: %s\n", err)

        for _, action := range sc.actions {
          if action != nil && action.action == "onDisconnect" {
            var params map[string]interface{}
            action.callback(params)
          }
        }

        sc.route.clients.Delete(sc.id)

        return
      }

      switch msg.Action {
        case "CardamomoSocketInit":
          params := make(map[string]interface{})
          params["id"] = sc.GetID()
          err := sc.Send("CardamomoSocketInit", params)
          if err != nil {
            fmt.Printf("Socket error: %s\n", err)
            
            for _, action := range sc.actions {
              if action != nil && action.action == "onDisconnect" {
                var params map[string]interface{}
                action.callback(params)
              }
            }
            
            sc.route.clients.Delete(sc.id)
          }

        case "CardamomoPing":
          err := sc.Send("CardamomoPong", make(map[string]interface{}))
          if err != nil {
            fmt.Printf("Socket error: %s\n", err)
            
            for _, action := range sc.actions {
              if action != nil && action.action == "onDisconnect" {
                var params map[string]interface{}
                action.callback(params)
              }
            }
            
            sc.route.clients.Delete(sc.id)
          }

        default:
          for _, action := range sc.actions {
            if action != nil && msg.Action == action.action {
              var params map[string]interface{}
              if err := json.Unmarshal([]byte(msg.Params), &params); err != nil {
                fmt.Printf("Error unmarshalling params: %s\n", err)
              } else {
                action.callback(params)
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
  defer func() {
    if err := recover(); err != nil {
      fmt.Println("Panic occurred on \"Send\":", err)
    }
  }()

  msg := SocketMessage{Action:action, Params: params}
  err := websocket.JSON.Send(sc.WebSocket, msg)

  if err != nil {
    fmt.Printf("Socket error: %s\n", err)

    for _, action := range sc.actions {
      if action != nil && action.action == "onDisconnect" {
        var params map[string]interface{}
        action.callback(params)
      }
    }

    sc.route.clients.Delete(sc.id)
  }
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
