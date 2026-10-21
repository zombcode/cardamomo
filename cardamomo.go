package cardamomo

import (
	"fmt"
	"runtime"
  "path"
  "net"
	"net/http"
	"strings"
	"strconv"
	"time"
)

type Cardamomo struct {
  router *Router
	socket *Socket
	compiledRoutes []*Route
  Config map[string]map[string]string
	errorHandler ErrorFunc
	tls CardamomoTLS
	// listenHost is the interface the HTTP server binds to. Empty keeps the
	// historical behaviour: every interface (":"+port).
	listenHost string
	// anchoredRoutes wraps every route regex in ^(?:...)$ so a pattern must
	// match the whole path instead of any substring of it. Off by default.
	anchoredRoutes bool
	routesPrepared bool
}

type CardamomoTLS struct {
	enabled bool
	cert string
	key string
}

type ErrorFunc func (code string, req Request, res Response) ()

func Instance(port string) Cardamomo {
  config := make(map[string]map[string]string)
  config["server"] = make(map[string]string)
  config["server"]["port"] = port
	config["development"] = make(map[string]string)
	config["development"]["debug"] = "false"
	config["production"] = make(map[string]string)
	config["production"]["debug"] = "false"

  // Get server IP
  ip := GetHostIP()
  config["server"]["ip"] = ip

  r := NewRouter("/")

  return Cardamomo{router: r, Config: config}
}

func (c *Cardamomo) SetDevDebugMode(debug bool) {
	c.Config["development"]["debug"] = strconv.FormatBool(debug)
}

func (c *Cardamomo) SetProdDebugMode(debug bool) {
	c.Config["production"]["debug"] = strconv.FormatBool(debug)
}

func (c *Cardamomo) SetErrorHandler(callback ErrorFunc) {
	c.errorHandler = callback
}

// SetListenHost sets the interface the HTTP server binds to, e.g. "127.0.0.1"
// to accept connections only from a local reverse proxy. An empty host (the
// default) binds every interface, as cardamomo always did. It must be called
// before Run. Config["server"]["ip"] is not affected: it stays the detected
// LAN address used in logs and cluster identity.
func (c *Cardamomo) SetListenHost(host string) {
	c.listenHost = host
}

// ListenAddr is the address handed to http.Server: ":port" by default, or
// "host:port" after SetListenHost.
func (c *Cardamomo) ListenAddr() string {
	if c.listenHost == "" {
		return ":" + c.Config["server"]["port"]
	}
	return net.JoinHostPort(c.listenHost, c.Config["server"]["port"])
}

// SetAnchoredRoutes makes parametrised routes match the whole request path
// (the compiled regex becomes ^(?:pattern)$). Without it, which is the
// historical behaviour, a route matches any path that merely contains it, so
// "/users/:id" also answers "/foo/users/1" and the first registered route
// wins. It must be called before Run.
func (c *Cardamomo) SetAnchoredRoutes(anchored bool) {
	c.anchoredRoutes = anchored
}

func (c *Cardamomo) SetTLS(cert, key string) {
	c.tls.enabled = true
	c.tls.cert = cert
	c.tls.key = key
}

// HTTP Server

func (c *Cardamomo) Run() error {
	return c.RunAndCallback(nil)
}

func (c *Cardamomo) RunAndCallback(handler http.Handler) error {
  // Run server
	_, filename, _, ok := runtime.Caller(0)
  if !ok {
      panic("No caller information")
  }
	http.Handle("/cardamomo/", http.StripPrefix("/cardamomo/", http.FileServer(http.Dir(path.Dir(filename) + "/static"))))

	http.HandleFunc("/", c.HandleFunc)

	c.prepareRoutes()

	// Build the server explicitly so idle keep-alive connections are reaped.
	// Without IdleTimeout, idle HTTP keep-alives hold a socket FD open forever
	// and pile up alongside leaked WebSocket sockets. WebSocket upgrades are
	// hijacked before these timeouts apply, so they are unaffected.
	server := &http.Server{
		Addr:              c.ListenAddr(),
		Handler:           handler,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	if c.tls.enabled == true {
		// Start HTTPS server
		fmt.Printf("\n * Starting HTTPS server at: https://%s:%s (listening on %s)\n", c.Config["server"]["ip"], c.Config["server"]["port"], c.ListenAddr())
	  return server.ListenAndServeTLS(c.tls.cert, c.tls.key)
	}

	// Start HTTP server
	fmt.Printf("\n * Starting HTTP server at: http://%s:%s (listening on %s)\n", c.Config["server"]["ip"], c.Config["server"]["port"], c.ListenAddr())
  return server.ListenAndServe()
}

func (c *Cardamomo) HandleFunc(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/favicon.ico" {
		return
	}

	currentRoute := c.matchRoute(req.Method, req.URL.Path)

	if currentRoute != nil {
		if strings.ToLower(req.Method) == strings.ToLower(currentRoute.method) {
			if c.Config["production"]["debug"] == "true" {
				fmt.Printf("\n %s: %s => %s \n", req.Method, currentRoute.pattern, req.URL.Path)
			}
			request := NewRequest(w, req, currentRoute)
			response := NewResponse(w, req)
			currentRoute.callback(request, response)
		} else {
			if c.Config["production"]["debug"] == "true" {
				fmt.Printf("\n HTTP ERROR: 404 - 1")
			}

			if c.errorHandler != nil {
				request := NewRequest(w, req, nil)
				response := NewResponse(w, req)
				c.errorHandler("404", request, response)
			}
		}
	} else {
		if c.Config["production"]["debug"] == "true" {
			fmt.Printf("\n HTTP ERROR: 404 - 2")
		}

		if c.errorHandler != nil {
			request := NewRequest(w, req, nil)
			response := NewResponse(w, req)
			c.errorHandler("404", request, response)
		}
	}
}

func compileRoutes(c *Cardamomo, router *Router) {
	for index, router := range router.routers {
		index = 1
		_ = index

		fmt.Printf("\n   + Base: %s\n\n", router.pattern)
		for index, route := range router.routes {
			index = 1
			_ = index

			fmt.Printf("     - Pattern: %s %s ✓\n", route.pattern, strings.ToUpper(route.method))
			c.compiledRoutes = append(c.compiledRoutes, route)
		}

		compileRoutes(c, router)
	}
}

func (c *Cardamomo) Base(pattern string, callback BaseFunc) {
  c.router.Base(pattern, callback)
}

func (c *Cardamomo) Get(pattern string, callback ReqFunc) {
  c.router.Get(pattern, callback)
}

func (c *Cardamomo) Head(pattern string, callback ReqFunc) {
  c.router.Head(pattern, callback)
}

func (c *Cardamomo) Post(pattern string, callback ReqFunc) {
  c.router.Post(pattern, callback)
}

func (c *Cardamomo) Put(pattern string, callback ReqFunc) {
  c.router.Put(pattern, callback)
}

func (c *Cardamomo) Delete(pattern string, callback ReqFunc) {
  c.router.Delete(pattern, callback)
}

func (c *Cardamomo) Connect(pattern string, callback ReqFunc) {
  c.router.Connect(pattern, callback)
}

func (c *Cardamomo) Options(pattern string, callback ReqFunc) {
  c.router.Options(pattern, callback)
}

// Socket

func (c *Cardamomo) OpenSocket() *Socket {
  return NewSocket(c)
}

func (c *Cardamomo) OpenSecureSocket(cert, key string) *Socket {
  return NewSecureSocket(c, cert, key)
}

func (c *Cardamomo) GetSocket() *Socket {
  return c.socket
}

// Utils
var hostIP string = ""
func GetHostIP() string {
  if hostIP == "" {
    tt, err := net.Interfaces()
    if err != nil {
      return ""
    }
    for _, t := range tt {
      aa, err := t.Addrs()
      if err != nil {
        return ""
      }
      for _, a := range aa {
        ipnet, ok := a.(*net.IPNet)
        if !ok {
          continue
        }
        v4 := ipnet.IP.To4()
        if v4 == nil || v4[0] == 127 { // loopback address
          continue
        }
        return v4.String()
      }
    }
    return ""
    } else {
      return hostIP
    }
  }

/*if hostIP == "" {
  var ip net.IP
  ifaces, err := net.Interfaces()
  if err == nil {
    for _, i := range ifaces {
      addrs, err := i.Addrs()
      if err == nil {
        for _, addr := range addrs {
          switch v := addr.(type) {
          case *net.IPNet:
            ip = v.IP
          }
        }
      }
    }
  }

  hostIP = ip.String()

  return hostIP
} else {
  return hostIP
}*/
