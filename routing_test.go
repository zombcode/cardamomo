package cardamomo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestApp(anchored bool) *Cardamomo {
	c := Instance("3400")
	c.SetAnchoredRoutes(anchored)
	return &c
}

func TestListenAddr(t *testing.T) {
	c := Instance("3400")
	if got := c.ListenAddr(); got != ":3400" {
		t.Fatalf("default ListenAddr = %q, want :3400", got)
	}
	c.SetListenHost("127.0.0.1")
	if got := c.ListenAddr(); got != "127.0.0.1:3400" {
		t.Fatalf("ListenAddr = %q, want 127.0.0.1:3400", got)
	}
	c.SetListenHost("::1")
	if got := c.ListenAddr(); got != "[::1]:3400" {
		t.Fatalf("ListenAddr = %q, want [::1]:3400", got)
	}
	c.SetListenHost("")
	if got := c.ListenAddr(); got != ":3400" {
		t.Fatalf("reset ListenAddr = %q, want :3400", got)
	}
}

func noop(req Request, res Response) {}

// registerSample mirrors the shape of a real table: a parametrised route that
// overlaps a longer one registered after it.
func registerSample(c *Cardamomo) {
	c.Base("/api", func(r *Router) {
		r.Get("/clients/:id/nirvana/users/:uid", noop)
		r.Get("/clients/:id", noop)
		r.Get("/users/:id", noop)
		r.Get("/users", noop)
		r.Post("/users/:id", noop)
		r.Get("/files/{{[0-9]+}}/:name", noop)
	})
	c.Base("/pay", func(r *Router) {
		r.Get("/confirm/:token", noop)
	})
}

func TestMatchRouteLegacyUnanchored(t *testing.T) {
	c := newTestApp(false)
	registerSample(c)

	cases := []struct {
		method, path, want string
	}{
		{"GET", "/api/users", "/api/users"},
		{"GET", "/api/users/42", "/api/users/:id"},
		{"POST", "/api/users/42", "/api/users/:id"},
		{"GET", "/api/clients/7/nirvana/users/9", "/api/clients/:id/nirvana/users/:uid"},
		// The param class includes "/": the shorter route swallows longer paths.
		{"GET", "/api/clients/7/other/thing", "/api/clients/:id"},
		// Unanchored: any path that contains the pattern matches.
		{"GET", "/evil/api/users/1", "/api/users/:id"},
		{"GET", "/api/x/pay/confirm/abc", "/pay/confirm/:token"},
		{"GET", "/api/users/a b", "/api/users/:id"},
	}
	for _, tc := range cases {
		got, ok := c.MatchRoute(tc.method, tc.path)
		if !ok || got != tc.want {
			t.Errorf("%s %s -> %q (%v), want %q", tc.method, tc.path, got, ok, tc.want)
		}
	}
	if _, ok := c.MatchRoute("DELETE", "/api/users/42"); ok {
		t.Errorf("DELETE matched a route that only has GET/POST")
	}
}

func TestMatchRouteAnchored(t *testing.T) {
	c := newTestApp(true)
	registerSample(c)

	match := []struct {
		method, path, want string
	}{
		{"GET", "/api/users", "/api/users"},
		{"GET", "/api/users/42", "/api/users/:id"},
		{"GET", "/api/clients/7/nirvana/users/9", "/api/clients/:id/nirvana/users/:uid"},
		{"GET", "/api/clients/7/other/thing", "/api/clients/:id"},
		{"GET", "/api/files/12/report.pdf", "/api/files/{{[0-9]+}}/:name"},
		{"GET", "/pay/confirm/abc", "/pay/confirm/:token"},
	}
	for _, tc := range match {
		got, ok := c.MatchRoute(tc.method, tc.path)
		if !ok || got != tc.want {
			t.Errorf("%s %s -> %q (%v), want %q", tc.method, tc.path, got, ok, tc.want)
		}
	}

	miss := []string{"/evil/api/users/1", "/api/x/pay/confirm/abc", "/api/users/a b", "/api/files/x12/report.pdf"}
	for _, path := range miss {
		if got, ok := c.MatchRoute("GET", path); ok {
			t.Errorf("anchored GET %s matched %q, want no match", path, got)
		}
	}
}

func TestRoutesCompiledOnce(t *testing.T) {
	c := newTestApp(true)
	registerSample(c)
	c.prepareRoutes()

	before := map[*Route]interface{}{}
	for _, r := range c.compiledRoutes {
		if r.patternRegex != "" && r.regex == nil {
			t.Fatalf("route %s has no compiled regex", r.pattern)
		}
		before[r] = r.regex
	}
	n := len(c.compiledRoutes)

	c.prepareRoutes()
	c.MatchRoute("GET", "/api/users/1")
	if len(c.compiledRoutes) != n {
		t.Fatalf("prepareRoutes appended routes twice: %d -> %d", n, len(c.compiledRoutes))
	}
	for _, r := range c.compiledRoutes {
		if before[r] != interface{}(r.regex) {
			t.Fatalf("route %s regex was recompiled", r.pattern)
		}
	}
}

func TestInvalidRouteRegexIsSkipped(t *testing.T) {
	c := newTestApp(false)
	c.Get("/bad/{{[}}/:id", noop)
	c.Get("/bad/:id", noop)
	got, ok := c.MatchRoute("GET", "/bad/1")
	if !ok || got != "/bad/:id" {
		t.Fatalf("got %q (%v), want /bad/:id", got, ok)
	}
}

func TestHandleFuncParams(t *testing.T) {
	for _, anchored := range []bool{false, true} {
		c := newTestApp(anchored)
		c.Get("/api/clients/:id/users/:uid", func(req Request, res Response) {
			res.Send(req.GetParam("id", "-") + "|" + req.GetParam("uid", "-") + "|" + req.GetParam("q", "-"))
		})
		c.Get("/api/static", func(req Request, res Response) {
			res.Send("static:" + req.GetParam("id", "none"))
		})
		var notFound string
		c.SetErrorHandler(func(code string, req Request, res Response) {
			notFound = code
			// GetParam without a route must not panic.
			res.Send("err:" + code + ":" + req.GetParam("id", "none"))
		})
		c.prepareRoutes()

		srv := httptest.NewServer(http.HandlerFunc(c.HandleFunc))

		get := func(path string) string {
			resp, err := http.Get(srv.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			return string(b)
		}

		if got := get("/api/clients/c1/users/u2?q=hello"); got != "c1|u2|hello" {
			t.Errorf("anchored=%v params = %q", anchored, got)
		}
		if got := get("/api/static?id=7"); got != "static:7" {
			t.Errorf("anchored=%v static = %q", anchored, got)
		}
		if got := get("/nope"); !strings.HasPrefix(got, "err:404:none") || notFound != "404" {
			t.Errorf("anchored=%v 404 = %q", anchored, got)
		}
		srv.Close()
	}
}

func TestGetParamOnHandBuiltRoute(t *testing.T) {
	// A Route that never went through prepareRoutes still resolves params.
	route := NewRoute("get", "/a/:id", noop)
	route.patternRegex = "/a" + routeParamClass
	route.paramsOrder = []string{"id"}
	req := httptest.NewRequest("GET", "/a/xyz", nil)
	r := NewRequest(httptest.NewRecorder(), req, &route)
	if got := r.GetParam("id", ""); got != "xyz" {
		t.Fatalf("GetParam = %q, want xyz", got)
	}
}

func TestRoutesListsInMatchingOrder(t *testing.T) {
	c := newTestApp(false)
	registerSample(c)
	got := c.Routes()
	want := []RouteInfo{
		{"GET", "/api/clients/:id/nirvana/users/:uid"},
		{"GET", "/api/clients/:id"},
		{"GET", "/api/users/:id"},
		{"GET", "/api/users"},
		{"POST", "/api/users/:id"},
		{"GET", "/api/files/{{[0-9]+}}/:name"},
		{"GET", "/pay/confirm/:token"},
	}
	if len(got) != len(want) {
		t.Fatalf("Routes() = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Routes()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
