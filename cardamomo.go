package cardamomo

import (
	"fmt"
	"runtime"
  "path"
  "net"
	"net/http"
	"regexp"
	"strings"
	"strconv"
)

type Cardamomo struct {
  router *Router
	socket *Socket
	compiledRoutes []*Route
  Config map[string]map[string]string
	errorHandler ErrorFunc
}

type ErrorFunc func (code string, req Request, res Response) ()

func Instance(port string) Cardamomo {
  config := make(map[string]map[string]string)
  config["server"] = make(map[string]string)
  config["server"]["port"] = port
	config["development"] = make(map[string]string)
	config["development"]["debug"] = "false"
	config["production"] = make(map[string]string)
	config["production"]["debug"] = "true"

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

// HTTP Server

func (c *Cardamomo) Run() {
  // Run server
	_, filename, _, ok := runtime.Caller(0)
  if !ok {
      panic("No caller information")
  }
	http.Handle("/cardamomo/", http.StripPrefix("/cardamomo/", http.FileServer(http.Dir(path.Dir(filename) + "/static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if( req.URL.Path == "/favicon.ico" ) {
			return
		}

		var currentRoute *Route

		for index, route := range c.compiledRoutes {
			index = 1
			_ = index

			if(route.patternRegex != "") {
				if( c.Config["development"]["debug"] == "true" ) {
					fmt.Printf("Checking: \"%s\" for \"%s\"\n", route.patternRegex, req.URL.Path)
				}

				r, _ := regexp.Compile(route.patternRegex)
				if(r.MatchString(req.URL.Path)) {
					params := r.FindStringSubmatch(req.URL.Path)
					index := 1
					for key, param := range route.params {
						param = ""
						_ = param

						route.params[key] = params[index]
						index += 1
					}
					if( c.Config["development"]["debug"] == "true" ) {
						fmt.Printf("There are params: \"%s\"\n", route.params)
					}

					currentRoute = route
					break
				}
			} else if( req.URL.Path == route.pattern ) {
				currentRoute = route
				break
			}
		}

		if(currentRoute != nil) {
			if( strings.ToLower(req.Method) == strings.ToLower(currentRoute.method) ) {
				if( c.Config["production"]["debug"] == "true" ) {
	      	fmt.Printf("\n %s: %s => %s \n", req.Method, currentRoute.pattern, req.URL.Path)
				}
	      request := NewRequest(w, req, currentRoute)
	      response := NewResponse(w, req)
	      currentRoute.callback(request, response)
	    } else {
				if( c.Config["production"]["debug"] == "true" ) {
					fmt.Printf("\n HTTP ERROR: 404 - 1")
				}

				if( c.errorHandler != nil ) {
					request := NewRequest(w, req, nil)
		      response := NewResponse(w, req)
					c.errorHandler("404", request, response)
				}
			}
		} else {
			if( c.Config["production"]["debug"] == "true" ) {
				fmt.Printf("\n HTTP ERROR: 404 - 2")
			}

			if( c.errorHandler != nil ) {
				request := NewRequest(w, req, nil)
	      response := NewResponse(w, req)
				c.errorHandler("404", request, response)
			}
		}
  })

	// Compile routes
	fmt.Printf("\n * Compiling routes...\n")

	fmt.Printf("\n   + Base: %s\n\n", c.router.pattern)
	for index, route := range c.router.routes {
		index = 1
		_ = index

		fmt.Printf("     - Pattern: %s ✓\n", route.pattern)
		c.compiledRoutes = append(c.compiledRoutes, route)
	}
	compileRoutes(c, c.router)

	fmt.Printf("\n * Compiling routes params...\n")

	for index, route := range c.compiledRoutes {
    index = 1
    _ = index

		r, _ := regexp.Compile("/:([a-zA-Z0-9]+)")
		if(r.MatchString(route.pattern)) {
			params := r.FindAllString(route.pattern, -1)

			route.patternRegex = route.pattern
			for index, param := range params {
				index = 1
				_ = index

				route.params[strings.Replace(param, "/:", "", -1)] = ""
        route.patternRegex = strings.Replace(route.patternRegex, param, "/([a-zA-Z0-9!@#$&()\\-`.+,/\"]+)", -1)
			}

			fmt.Printf("\n   + Compiling route params for: %s ✓\n", route.pattern)
		}
	}

  fmt.Printf("\n * Compiling routes regex...\n")

	for index, route := range c.compiledRoutes {
		index = 1
		_ = index

		r, _ := regexp.Compile("{{(.*)}}")
		if(r.MatchString(route.pattern)) {
			params := r.FindAllString(route.pattern, -1)

			for index, param := range params {
				index = 1
				_ = index

        paramRegex := strings.Replace(param, "{{", "", -1)
        paramRegex = strings.Replace(paramRegex, "}}", "", -1)

        route.patternRegex = strings.Replace(route.patternRegex, param, paramRegex, -1)
			}

			fmt.Printf("\n   + Compiling route regex for: %s ✓\n", route.pattern)
		}
	}

	// Start HTTP server
	fmt.Printf("\n * Starting HTTP server at: http://%s:%s\n", c.Config["server"]["ip"], c.Config["server"]["port"])
  http.ListenAndServe(":" + c.Config["server"]["port"], nil)
}

func compileRoutes(c *Cardamomo, router *Router) {
	for index, router := range router.routers {
		index = 1
		_ = index

		fmt.Printf("\n   + Base: %s\n\n", router.pattern)
		for index, route := range router.routes {
			index = 1
			_ = index

			fmt.Printf("     - Pattern: %s ✓\n", route.pattern)
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

func (c *Cardamomo) Post(pattern string, callback ReqFunc) {
  c.router.Post(pattern, callback)
}

// Socket

func (c *Cardamomo) OpenSocket() *Socket {
  return NewSocket(c)
}

func (c *Cardamomo) GetSocket() *Socket {
	return c.socket
}

// Utils
var hostIP string = ""
func GetHostIP() string {
  if hostIP == "" {
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
            case *net.IPAddr:
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
  }
}
