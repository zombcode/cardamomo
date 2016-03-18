package main

import (
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

	socket := c.OpenSocket("8001")
	socket.SocketBase("/base1", func(){

	})

  c.Run()
}
