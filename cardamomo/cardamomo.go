package cardamomo

import (
	"fmt"
	"net/http"
)

type Cardamomo struct {
  Router
  *Response
  Config map[string]map[string]string
}

func Instance() Cardamomo {
  config := make(map[string]map[string]string)
  config["server"] = make(map[string]string)
  config["server"]["port"] = "8000"

  fmt.Printf("\n\nStarting http server at: http://localhost:%s\n\n", config["server"]["port"])

  r := NewRouter("/")

  return Cardamomo{Router: r, Config: config}
}

func (c *Cardamomo) Run() {
  http.ListenAndServe(":" + c.Config["server"]["port"], nil)
}
