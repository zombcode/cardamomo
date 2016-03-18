package cardamomo

import (
	"fmt"
	"net/http"
)

type Cardamomo struct {
  Router
	Socket
  Config map[string]map[string]string
}

func Instance(port string) Cardamomo {
  config := make(map[string]map[string]string)
  config["server"] = make(map[string]string)
  config["server"]["port"] = port

  r := NewRouter("/")

  return Cardamomo{Router: r, Config: config}
}

// HTTP Server

func (c *Cardamomo) Run() {
	fmt.Printf("\n\nStarting HTTP server at: http://localhost:%s\n\n", c.Config["server"]["port"])
  http.ListenAndServe(":" + c.Config["server"]["port"], nil)
}

// Socket

func (c *Cardamomo) OpenSocket(port string) Socket {
  return NewSocket(port)
}

func (c *Cardamomo) GetSocket() Socket {
	return c.Socket
}
