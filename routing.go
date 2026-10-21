package cardamomo

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	routeParamRegex  = regexp.MustCompile("/:([a-zA-Z0-9]+)")
	routeInlineRegex = regexp.MustCompile("{{(.*)}}")
)

// routeParamClass is what a ":name" segment expands to. It includes "/", so a
// parameter can span several segments.
const routeParamClass = "/([a-zA-Z0-9!@#$&()_\\-`.+,/\"]+)"

// prepareRoutes flattens the router tree and compiles every route regex once.
// Run calls it before serving; MatchRoute calls it lazily. Later calls are
// no-ops, so routes registered after the first call are ignored, as they
// always were once the server was running.
func (c *Cardamomo) prepareRoutes() {
	if c.routesPrepared {
		return
	}
	c.routesPrepared = true

	fmt.Printf("\n * Compiling routes...\n")

	fmt.Printf("\n   + Base: %s\n\n", c.router.pattern)
	for _, route := range c.router.routes {
		fmt.Printf("     - Pattern: %s %s ✓\n", route.pattern, strings.ToUpper(route.method))
		c.compiledRoutes = append(c.compiledRoutes, route)
	}
	compileRoutes(c, c.router)

	fmt.Printf("\n * Compiling routes params...\n")

	for _, route := range c.compiledRoutes {
		if routeParamRegex.MatchString(route.pattern) {
			params := routeParamRegex.FindAllString(route.pattern, -1)

			route.patternRegex = route.pattern
			for _, param := range params {
				name := strings.Replace(param, "/:", "", -1)
				route.params[name] = ""
				route.paramsOrder = append(route.paramsOrder, name)
				route.patternRegex = strings.Replace(route.patternRegex, param, routeParamClass, -1)
			}

			fmt.Printf("\n   + Compiling route params for: %s ✓\n", route.pattern)
		}
	}

	fmt.Printf("\n * Compiling routes regex...\n")

	for _, route := range c.compiledRoutes {
		if routeInlineRegex.MatchString(route.pattern) {
			params := routeInlineRegex.FindAllString(route.pattern, -1)

			for _, param := range params {
				paramRegex := strings.Replace(param, "{{", "", -1)
				paramRegex = strings.Replace(paramRegex, "}}", "", -1)

				route.patternRegex = strings.Replace(route.patternRegex, param, paramRegex, -1)
			}

			fmt.Printf("\n   + Compiling route regex for: %s ✓\n", route.pattern)
		}
	}

	// Compile each regex once here instead of on every request (and again in
	// every GetParam call).
	for _, route := range c.compiledRoutes {
		if route.patternRegex == "" {
			continue
		}
		expr := route.patternRegex
		if c.anchoredRoutes {
			expr = "^(?:" + expr + ")$"
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			// The route can never match; before, it panicked on every request
			// that reached it.
			fmt.Printf("\n   ! Invalid route regex for %s %s: %s\n", strings.ToUpper(route.method), route.pattern, err)
			continue
		}
		route.regex = re
	}
}

// matchRoute returns the first registered route for method and path, or nil.
func (c *Cardamomo) matchRoute(method string, path string) *Route {
	debug := c.Config["development"]["debug"] == "true"

	for _, route := range c.compiledRoutes {
		if route.patternRegex != "" && strings.EqualFold(route.method, method) {
			if debug {
				fmt.Printf("Checking: \"%s\" for \"%s\"\n", route.patternRegex, path)
			}

			if route.regex != nil && route.regex.MatchString(path) {
				return route
			}
		} else {
			if debug {
				fmt.Printf("Checking: \"%s\" for \"%s\" \"%s:%s\"\n", route.pattern, path, strings.ToUpper(route.method), method)
			}

			if strings.EqualFold(route.method, method) && path == route.pattern {
				return route
			}
		}
	}

	return nil
}

// MatchRoute reports which registered pattern would serve method and path,
// without serving it. It is meant for tests that check a route table (for
// instance, that no route shadows another). It compiles the routes on first
// use, so call it only once every route is registered.
func (c *Cardamomo) MatchRoute(method string, path string) (pattern string, ok bool) {
	c.prepareRoutes()
	route := c.matchRoute(method, path)
	if route == nil {
		return "", false
	}
	return route.pattern, true
}

// RouteInfo describes one registered HTTP route.
type RouteInfo struct {
	Method  string // upper case: GET, POST...
	Pattern string // as registered, base included: /api/users/:id
}

// Routes lists every registered HTTP route in matching order (the order in
// which HandleFunc tries them). Like MatchRoute, it is meant for tests of a
// route table and compiles the routes on first use.
func (c *Cardamomo) Routes() []RouteInfo {
	c.prepareRoutes()
	out := make([]RouteInfo, 0, len(c.compiledRoutes))
	for _, route := range c.compiledRoutes {
		out = append(out, RouteInfo{Method: strings.ToUpper(route.method), Pattern: route.pattern})
	}
	return out
}
