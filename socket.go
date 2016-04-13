package cardamomo

import (
  "fmt"
  "golang.org/x/net/websocket"
  "encoding/json"
  b64 "encoding/base64"
  "time"
  "io"
  "os"
)

type Socket struct {
  server *Cardamomo
  routes []*SocketRoute
  clustered bool
  clusterParams *SocketClusterParams
  clusterMaster bool
  clusterConnections []*SocketClusterConnection
}

type SockFunc func (route *SocketClient) ()
type SockActionFunc func (params map[string]interface{}) ()

func NewSocket(c *Cardamomo) *Socket {
  fmt.Printf("\n * Starting WebSocket server...\n")

  return &Socket{server:c, clustered:false}
}

func (s *Socket) OnSocketBase(pattern string, callback SockFunc) {
  s.addBase(pattern, callback);
}

func (s *Socket) addBase(pattern string, callback SockFunc) SocketRoute {
  route := NewSocketRoute(s, pattern, callback)
  route.Listen()
  s.routes = append(s.routes, &route)

  return route
}

func (s *Socket) Send(action string, params interface{}) {
  for index, route := range s.routes {
    index = 1
    _ = index

    for index, client := range route.clients {
      index = ""
      _ = index

      client.Send(action, params)
    }
  }

  if s.clustered == true {
    for index, conn := range s.clusterConnections {
      index = 1
      _ = index

      if conn.connection != nil {
        params := JSONC{
          "action": action,
          "params": params,
          "host": JSONC{
            "ip": s.server.Config["server"]["ip"],
            "port": s.server.Config["server"]["port"],
          },
        }

        paramsStr, _ := json.Marshal(params)

        err := websocket.JSON.Send(conn.connection, SocketMessage{Action: "onSend", Params: string(paramsStr)})
        if err != nil {
          fmt.Printf("Send failed: %s\n", err.Error())
        }
      }
    }
  }
}

func (s *Socket) sendCluster(action string, params interface{}) {
  for index, route := range s.routes {
    index = 1
    _ = index

    for index, client := range route.clients {
      index = ""
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
        index = ""
        _ = index

        client.Send(action, params)
      }

      break
    }
  }

  if s.clustered == true {
    for index, conn := range s.clusterConnections {
      index = 1
      _ = index

      if conn.connection != nil {
        params := JSONC{
          "action": action,
          "params": params,
          "base": base,
          "host": JSONC{
            "ip": s.server.Config["server"]["ip"],
            "port": s.server.Config["server"]["port"],
          },
        }

        paramsStr, _ := json.Marshal(params)

        err := websocket.JSON.Send(conn.connection, SocketMessage{Action: "onSendBase", Params: string(paramsStr)})
        if err != nil {
          fmt.Printf("Send failed: %s\n", err.Error())
        }
      }
    }
  }
}

func (s *Socket) sendBaseCluster(base string, action string, params interface{}) {
  for index, route := range s.routes {
    index = 1
    _ = index

    if( route.pattern == base ) {
      for index, client := range route.clients {
        index = ""
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
        index = ""
        _ = index

        if( client.id == clientID ) {
          client.Send(action, params)

          break RoutesLoop
        }
      }
    }

  if s.clustered == true {
    for index, conn := range s.clusterConnections {
      index = 1
      _ = index

      if conn.connection != nil {
        params := JSONC{
          "action": action,
          "params": params,
          "client_id": clientID,
          "host": JSONC{
            "ip": s.server.Config["server"]["ip"],
            "port": s.server.Config["server"]["port"],
          },
        }

        paramsStr, _ := json.Marshal(params)

        err := websocket.JSON.Send(conn.connection, SocketMessage{Action: "onSendClient", Params: string(paramsStr)})
        if err != nil {
          fmt.Printf("Send failed: %s\n", err.Error())
        }
      }
    }
  }
}

func (s *Socket) sendClientCluster(clientID string, action string, params interface{}) {
  RoutesLoop:
    for index, route := range s.routes {
      index = 1
      _ = index

      for index, client := range route.clients {
        index = ""
        _ = index

        if( client.id == clientID ) {
          client.Send(action, params)

          break RoutesLoop
        }
      }
    }
}

type SocketClusterParams struct {
  Hosts []SocketClusterHost
  Password string
}

type SocketClusterHost struct {
  Host string
  Port string
  Master bool
}

type SocketClusterConnection struct {
  host *SocketClusterHost
  connection *websocket.Conn
  reconnecting bool
  config *websocket.Config
}

func (s *Socket) Cluster(params SocketClusterParams) {
  fmt.Printf("   + Preparing cluster...\n")

  s.clustered = true
  s.addBase("/cardacluster", clusterControl);

  var hosts = []SocketClusterHost{}
  for index, host := range params.Hosts {
    index = 1
    _ = index

    if host.Host != s.server.Config["server"]["ip"] || host.Port != s.server.Config["server"]["port"] {
      hosts = append(hosts, host)
    }

    if host.Master == true && host.Host == s.server.Config["server"]["ip"] && host.Port == s.server.Config["server"]["port"] {
      s.clusterMaster = true
      fmt.Printf("     - This is the master!\n")
    }

    if host.Master == true && (host.Host != s.server.Config["server"]["ip"] || host.Port != s.server.Config["server"]["port"]) {
      hosts = []SocketClusterHost{}
      hosts = append(hosts, host)
      fmt.Printf("     - Cluster master \"%s:%s\"\n", host.Host, host.Port)
      break
    }
  }

  params.Hosts = hosts

  s.clusterParams = &params

  connectToCluster(s)
}

func connectToCluster(s *Socket) {
  for index, host := range s.clusterParams.Hosts {
    index = 1
    _ = index

    fmt.Printf("     - Connecting to \"ws://%s/cardacluster\" ... \n", host.Host + ":" + host.Port)

    config, err := websocket.NewConfig(fmt.Sprintf("ws://%s/cardacluster", host.Host + ":" + host.Port), fmt.Sprintf("http://%s/", host.Host + ":" + host.Port))
    if err == nil {
      config.Header.Add("Cardamomo-Cluster-Password", b64.StdEncoding.EncodeToString([]byte(s.clusterParams.Password)))

      connection := &SocketClusterConnection{
        host: &host,
        reconnecting: false,
        config: config,
      }
      s.clusterConnections = append(s.clusterConnections, connection)

      var conn *websocket.Conn
      conn, err := websocket.DialConfig(config)
      //conn, err := websocket.Dial(fmt.Sprintf("ws://%s/cardacluster", host.Host + ":" + host.Port), "", fmt.Sprintf("http://%s/", host.Host + ":" + host.Port))
      if err == nil {
        connection.connection = conn
        go readClusterClient(connection)
      } else {
        connection.reconnecting = true
        go reconnectCluster(connection)
      }
    } else {
      os.Exit(1)
    }
  }
}

func readClusterClient(connection *SocketClusterConnection) {
  for {
    var message string
    err := websocket.Message.Receive(connection.connection, &message)
    if err == io.EOF {
      if connection.reconnecting == false {
        connection.reconnecting = true
        go reconnectCluster(connection)
      }
    } else if message == "cardamomoinvalidpassword" {
      fmt.Printf("\n\nERROR :: Invalid password for cluster in host (%s:%s)\n\n", connection.host.Host, connection.host.Port)
      os.Exit(1)
    }
  }
}

func reconnectCluster(connection *SocketClusterConnection) {
  conn, err := websocket.DialConfig(connection.config)
  //conn, err := websocket.Dial(fmt.Sprintf("ws://%s/cardacluster", connection.host.Host + ":" + connection.host.Port), "", fmt.Sprintf("http://%s/", connection.host.Host + ":" + connection.host.Port))
  if err != nil {
    time.Sleep(1 * time.Second)
    reconnectCluster(connection)
  } else {
    connection.reconnecting = false
    connection.connection = conn
  }
}

func clusterControl(client *SocketClient) {
  client.OnSocketAction("onSend", func(sparams map[string]interface{}) {
    client.route.socket.sendCluster(sparams["action"].(string), sparams["params"].(interface{}))
  })

  client.OnSocketAction("onSendBase", func(sparams map[string]interface{}) {
    client.route.socket.sendBaseCluster(sparams["base"].(string), sparams["action"].(string), sparams["params"].(interface{}))
  })

  client.OnSocketAction("onSendClient", func(sparams map[string]interface{}) {
    client.route.socket.sendClientCluster(sparams["client_id"].(string), sparams["action"].(string), sparams["params"].(interface{}))
  })
}
