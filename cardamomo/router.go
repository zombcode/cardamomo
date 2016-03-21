package cardamomo

import (
)

type Router struct {
  pattern string
  routes []*Route
  routers []*Router
}

type BaseFunc func (router *Router) ()
type ReqFunc func (req Request, res Response) ()

func NewRouter(pattern string) *Router {
  return &Router{pattern: pattern}
}

func (r *Router) Base(pattern string, callback BaseFunc) {
  router := r.addBase(pattern)
  callback(router)
}

func (r *Router) addBase(pattern string) *Router {
  if( r.pattern != "/" ) {
    pattern = r.pattern + pattern
  }

  router := NewRouter(pattern)
  r.routers = append(r.routers, router)

  return router
}

func (r *Router) Get(pattern string, callback ReqFunc) {
  r.addRoute("get", pattern, callback)
}

func (r *Router) Post(pattern string, callback ReqFunc) {
  r.addRoute("post", pattern, callback)
}

func (r *Router) addRoute(method string, pattern string, callback ReqFunc) {
  if( r.pattern != "/" ) {
    pattern = r.pattern + pattern
  }

  route := NewRoute(method, pattern, callback)
  r.routes = append(r.routes, &route)
}
