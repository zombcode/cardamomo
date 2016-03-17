package main

import (
	"../cardamomo"
)

func main() {
  c := cardamomo.Instance()
  c.Get("/", func(res cardamomo.Response) {
    res.Send("Hello world!");
  })

  c.Get("/route2", func(res cardamomo.Response) {
    res.Send("Hello route 2!");
  })

  c.Post("/route3", func(res cardamomo.Response) {
    res.Send("Hello route 3!");
  })

  c.Base("/base1", func(router cardamomo.Router) {
    router.Get("/route1", func(res cardamomo.Response) {
      res.Send("Hello route base1/route1!");
    })
  })

  c.Run()
}
