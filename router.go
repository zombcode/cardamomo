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

func (r *Router) Head(pattern string, callback ReqFunc) {
  r.addRoute("head", pattern, callback)
}

func (r *Router) Post(pattern string, callback ReqFunc) {
  r.addRoute("post", pattern, callback)
}

func (r *Router) Put(pattern string, callback ReqFunc) {
  r.addRoute("put", pattern, callback)
}

func (r *Router) Delete(pattern string, callback ReqFunc) {
  r.addRoute("delete", pattern, callback)
}

func (r *Router) Connect(pattern string, callback ReqFunc) {
  r.addRoute("connect", pattern, callback)
}

func (r *Router) Options(pattern string, callback ReqFunc) {
  r.addRoute("options", pattern, callback)
}

func (r *Router) addRoute(method string, pattern string, callback ReqFunc) {
  if( r.pattern != "/" ) {
    pattern = r.pattern + pattern
  }

  route := NewRoute(method, pattern, callback)
  r.routes = append(r.routes, &route)
}
