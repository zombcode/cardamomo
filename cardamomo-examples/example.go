package main

import (
  "fmt"
	"../cardamomo"
)

func main() {
  c := cardamomo.Instance()
  c.Get("/", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello world!");
  })

  c.Get("/routeget1", func(req cardamomo.Request, res cardamomo.Response) {
    res.Send("Hello route get 1!");
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

  c.Run()
}
