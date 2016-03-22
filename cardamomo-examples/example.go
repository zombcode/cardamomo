package main

import (
	"fmt"
	"time"
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

	c.Get("/routeget2/:param1/and/:param2", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello route get 1 with param1 = " + req.GetParam("param1", "") + " and param2 = " + req.GetParam("param2", "") + "!");
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

  c.Base("/base1", func(router *cardamomo.Router) {
    router.Get("/routeget1", func(req cardamomo.Request, res cardamomo.Response) {
      res.Send("Hello route base1/routeget1!");
    })
    router.Post("/routepost1", func(req cardamomo.Request, res cardamomo.Response) {
      res.Send("Hello route base1/routepost1!");
    })
		router.Base("/base2", func(router *cardamomo.Router) {
			router.Get("/routeget1", func(req cardamomo.Request, res cardamomo.Response) {
	      res.Send("Hello route base1/base2/routeget1!");
	    })
		});
  })

	// Cookies

	c.Get("/setcookie/:key/:value", func(req cardamomo.Request, res cardamomo.Response) {
		key := req.GetParam("key", "")
		value := req.GetParam("value", "")

		expire := time.Now().AddDate(0, 0, 1) // Expires in one day!
		req.SetCookie(key, value, "/", "localhost", expire, 86400, false, false) // key, value, path, domain, expiration, max-age, httponly, secure

    res.Send("Added cookie \"" + key + "\"=\"" + value + "\"");
  })

	c.Get("/getcookie/:key", func(req cardamomo.Request, res cardamomo.Response) {
		key := req.GetParam("key", "")

		cookie := req.GetCookie(key, "empty cookie!"); // key, defaultValue

    res.Send("The value for cookie \"" + key + "\" is \"" + cookie + "\"");
  })

	// Sockets

	socket := c.OpenSocket()
	socket.OnSocketBase("/base1", func(client *cardamomo.SocketClient) {
		fmt.Printf("\n\nBase 1 new client!\n\n")

		client.OnSocketAction("action1", func(sparams map[string]interface{}) {
			fmt.Printf("\n\nAction 1!\n\n")

			fmt.Printf("\n\nParam: %s\n\n", sparams["param_1"])
			fmt.Printf("\n\nParam: %s\n\n", sparams["param_2"].(map[string]interface{})["inner_1"])
			fmt.Printf("\n\nParam: %d\n\n", int(sparams["param_2"].(map[string]interface{})["inner_2"].([]interface{})[1].(float64)))

			client.Send("action1", sparams)
		})
	})

  c.Run()
}
