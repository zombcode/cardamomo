package cardamomo

import (
	"fmt"
	"runtime"
  "path"
	"net/http"
	"regexp"
	"strings"
)

type Cardamomo struct {
  router *Router
	socket *Socket
	compiledRoutes []*Route
  Config map[string]map[string]string
}

func Instance(port string) Cardamomo {
  config := make(map[string]map[string]string)
  config["server"] = make(map[string]string)
  config["server"]["port"] = port

  r := NewRouter("/")

  return Cardamomo{router: r, Config: config}
}

// HTTP Server

func (c *Cardamomo) Run() {
	_, filename, _, ok := runtime.Caller(0)
  if !ok {
      panic("No caller information")
  }
	http.Handle("/cardamomo/", http.StripPrefix("/cardamomo/", http.FileServer(http.Dir(path.Dir(filename) + "/static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var currentRoute *Route

		for index, route := range c.compiledRoutes {
			index = 1
			_ = index

			if(route.patternRegex != "") {
				fmt.Printf("Checking: \"%s\" for \"%s\"\n", route.patternRegex, req.URL.Path)

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
					fmt.Printf("There are params: \"%s\"\n", route.params)

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
	      fmt.Printf("\n %s: %s \n", req.Method, currentRoute.pattern)
	      request := NewRequest(req, currentRoute)
	      response := NewResponse(w)
	      currentRoute.callback(request, response)
	    }
		}
  })

	// Compile routes
	fmt.Printf("\n * Compiling routes...\n")

	fmt.Printf("\nBase: %s\n\n", c.router.pattern)
	for index, route := range c.router.routes {
		index = 1
		_ = index

		fmt.Printf("Pattern: %s\n", route.pattern)
		c.compiledRoutes = append(c.compiledRoutes, route)
	}
	compileRoutes(c, c.router)

	for index, route := range c.compiledRoutes {
		index = 1
		_ = index

		r, _ := regexp.Compile("/:([a-zA-Z0-9]+)")
		if(r.MatchString(route.pattern)) {
			fmt.Printf("\nCompile route params for: %s\n", route.pattern)

			params := r.FindAllString(route.pattern, -1)

			route.patternRegex = route.pattern
			for index, param := range params {
				index = 1
				_ = index

				route.params[strings.Replace(param, "/:", "", -1)] = ""
				route.patternRegex = strings.Replace(route.patternRegex, param, "/([a-zA-Z0-9]+)", -1)
			}

			fmt.Printf("Compiled to: %s\n", route.patternRegex)
		}
	}

	// Start HTTP server
	fmt.Printf("\n * Starting HTTP server at: http://localhost:%s\n", c.Config["server"]["port"])
  http.ListenAndServe(":" + c.Config["server"]["port"], nil)
}

func compileRoutes(c *Cardamomo, router *Router) {
	for index, router := range router.routers {
		index = 1
		_ = index

		fmt.Printf("\nBase: %s\n\n", router.pattern)
		for index, route := range router.routes {
			index = 1
			_ = index

			fmt.Printf("Pattern: %s\n", route.pattern)
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
  return NewSocket()
}

func (c *Cardamomo) GetSocket() *Socket {
	return c.socket
}
