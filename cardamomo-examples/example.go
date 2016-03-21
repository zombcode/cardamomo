package main

import (
	"fmt"
	"../cardamomo"
)

type Box struct {
  Size BoxSize
  Color  string
  Open   bool
}

type BoxSize struct {
  Width  int
  Height int
}

func main() {

	// HTTP

  c := cardamomo.Instance("8000")
  c.Get("/", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello world!");
  })

  c.Get("/routeget1", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello route get 1!");
  })

  c.Get("/routejson", func(req cardamomo.Request, res cardamomo.Response) {
    boxsize := BoxSize {
      Width:  10,
    	Height: 20,
    }

    box := Box {
    	Size: boxsize,
    	Color:  "blue",
    	Open:   false,
    }

    res.SendJSON(box)
  })

  c.Post("/routepost1", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello route post 1!");
  })

  c.Base("/base1", func(router cardamomo.Router) {
    router.Get("/routeget1", func(req cardamomo.Request, res cardamomo.Response) {
      res.Send("Hello route base1/routeget1!");
    })
    router.Post("/routepost1", func(req cardamomo.Request, res cardamomo.Response) {
      res.Send("Hello route base1/routepost1!");
    })
  })

	// Sockets

	socket := c.OpenSocket()
	socket.OnSocketBase("/base1", func(client *cardamomo.SocketClient) {
		fmt.Printf("\n\nBase 1 new client!\n\n")

		client.OnSocketAction("action1", func(sparams map[string]interface{}) {
			fmt.Printf("\n\nAction 1!\n\n")

			fmt.Printf("\n\nParam: %s\n\n", sparams["param_1"])
			fmt.Printf("\n\nParam: %s\n\n", sparams["param_2"].(map[string]interface{})["inner_1"])
			fmt.Printf("\n\nParam: %b\n\n", sparams["param_2"].(map[string]interface{})["inner_2"].([]interface{})[1])

			client.Send("action1", sparams)
		})
	})

  c.Run()
}
