package cardamomo

import (
  "fmt"
  "encoding/json"
  "time"
  "math/rand"
  crand "crypto/rand"
  "math/big"
  "sync/atomic"
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
  // authenticated is set (atomically, 0/1) by SetAuthenticated and read by
  // the AuthTimeout timer from another goroutine.
  authenticated int32
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

// SetAuthenticated marks the client as authenticated, which stops the
// SocketOptions.AuthTimeout timer of its base from closing it. Call it once
// the client has proved who it is (a valid token in its first message, say).
// It is harmless on bases without AuthTimeout.
func (sc *SocketClient) SetAuthenticated() {
  atomic.StoreInt32(&sc.authenticated, 1)
}

// IsAuthenticated reports whether SetAuthenticated has been called.
func (sc *SocketClient) IsAuthenticated() bool {
  return atomic.LoadInt32(&sc.authenticated) == 1
}

func (sc *SocketClient) Listen() {
    defer func() {
      if err := recover(); err != nil {
        fmt.Println("Panic occurred on \"Listen\":", err)
      }
    }()

    // Guarantee the underlying socket FD is always released when Listen returns.
    // Without this, ungraceful/half-open disconnects leave sockets stuck in
    // CLOSE-WAIT and leak file descriptors until the process hits EMFILE.
    defer sc.WebSocket.Close()

    go func() {
      ticker := time.NewTicker(30 * time.Second)
      defer ticker.Stop()

      for {
        <-ticker.C
        err := sc.Send("CardamomoPing", make(map[string]interface{}))
        if err != nil {
          // The peer is gone: close the socket so the blocked Receive below
          // unblocks immediately and the client is reaped.
          sc.WebSocket.Close()
          return
        }
      }
    }()

    for {
      // Reap dead/half-open connections. Clients ping every ~10s, so no inbound
      // traffic for 120s means the peer is gone (mobile killed, network cut,
      // LB/CF dropped the conn without a FIN). This bounds the lifetime of a
      // leaked FD instead of letting Receive block forever.
      sc.WebSocket.SetReadDeadline(time.Now().Add(120 * time.Second))

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
          sc.Send("CardamomoSocketInit", params)

        case "CardamomoPing":
          sc.Send("CardamomoPong", make(map[string]interface{}))

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

func (sc *SocketClient) Send(action string, params interface{}) error {
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
    return err
  }
  return nil
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

// RandStringRunes returns n random letters. It draws from crypto/rand (client
// ids must not be guessable) and falls back to math/rand only if the system
// source fails.
func RandStringRunes(n int) string {
  b := make([]rune, n)
  max := big.NewInt(int64(len(letterRunes)))
  for i := range b {
    v, err := crand.Int(crand.Reader, max)
    if err != nil {
      b[i] = letterRunes[rand.Intn(len(letterRunes))]
      continue
    }
    b[i] = letterRunes[v.Int64()]
  }
  return string(b)
}
