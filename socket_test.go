package cardamomo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

// The socket bases register on http.DefaultServeMux, so every test uses its
// own pattern and serves the default mux.

var (
	socketServerOnce sync.Once
	socketServer     *httptest.Server
	socketPatternN   int64
)

func testSocketServer() *httptest.Server {
	socketServerOnce.Do(func() {
		socketServer = httptest.NewServer(http.DefaultServeMux)
	})
	return socketServer
}

func uniquePattern(name string) string {
	return fmt.Sprintf("/ws/test/%s/%d", name, atomic.AddInt64(&socketPatternN, 1))
}

func newTestSocket() *Socket {
	c := Instance("0")
	return c.OpenSocket()
}

func wsURL(pattern string) string {
	return "ws" + strings.TrimPrefix(testSocketServer().URL, "http") + pattern
}

func dial(t *testing.T, pattern string, origin string) (*websocket.Conn, error) {
	t.Helper()
	config, err := websocket.NewConfig(wsURL(pattern), origin)
	if err != nil {
		t.Fatal(err)
	}
	return websocket.DialConfig(config)
}

// rawHandshake sends an upgrade request by hand (x/net's client always sends
// an Origin) and returns the response status code.
func rawHandshake(t *testing.T, pattern string, origin string) int {
	t.Helper()
	addr := strings.TrimPrefix(testSocketServer().URL, "http://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	req := "GET " + pattern + " HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n"
	if origin != "" {
		req += "Origin: " + origin + "\r\n"
	}
	req += "\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatal(err)
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func send(t *testing.T, ws *websocket.Conn, action string, params interface{}) {
	t.Helper()
	b, _ := json.Marshal(params)
	if err := websocket.JSON.Send(ws, SocketClientMessage{Action: action, Params: string(b)}); err != nil {
		t.Fatalf("send %s: %v", action, err)
	}
}

// expectClosed waits for the server to close ws.
func expectClosed(t *testing.T, ws *websocket.Conn, within time.Duration) {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(within))
	for {
		var msg SocketMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				t.Fatalf("socket still open after %v", within)
			}
			return
		}
	}
}

// expectAlive does a CardamomoPing round trip.
func expectAlive(t *testing.T, ws *websocket.Conn) {
	t.Helper()
	send(t, ws, "CardamomoPing", map[string]interface{}{})
	ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	defer ws.SetReadDeadline(time.Time{})
	for {
		var msg SocketMessage
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			t.Fatalf("socket closed: %v", err)
		}
		if msg.Action == "CardamomoPong" {
			return
		}
	}
}

func TestSocketDefaultBehaviourUnchanged(t *testing.T) {
	s := newTestSocket()
	pattern := uniquePattern("default")
	s.OnSocketBase(pattern, func(c *SocketClient) {})

	// x/net's default: no Origin, no socket.
	if code := rawHandshake(t, pattern, ""); code != http.StatusForbidden {
		t.Fatalf("handshake without Origin = %d, want 403", code)
	}
	if code := rawHandshake(t, pattern, "http://anything.example"); code != http.StatusSwitchingProtocols {
		t.Fatalf("handshake with Origin = %d, want 101", code)
	}

	ws, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	send(t, ws, "CardamomoSocketInit", map[string]interface{}{})
	var msg SocketMessage
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		t.Fatal(err)
	}
	id, _ := msg.Params.(map[string]interface{})["id"].(string)
	if msg.Action != "CardamomoSocketInit" || len(id) != 32 {
		t.Fatalf("init reply = %+v", msg)
	}
	expectAlive(t, ws)
}

func TestSocketCheckOrigin(t *testing.T) {
	s := newTestSocket()
	pattern := uniquePattern("origin")
	var rejected int32
	s.OnSocketBaseWithOptions(pattern, SocketOptions{
		CheckOrigin: func(r *http.Request) bool {
			o := r.Header.Get("Origin")
			return o == "" || o == "https://panel.example"
		},
		OnReject: func(r *http.Request, reason string) {
			if reason == "origin" {
				atomic.AddInt32(&rejected, 1)
			}
		},
	}, func(c *SocketClient) {})

	if code := rawHandshake(t, pattern, "https://evil.example"); code != http.StatusForbidden {
		t.Fatalf("foreign Origin = %d, want 403", code)
	}
	if code := rawHandshake(t, pattern, "https://panel.example"); code != http.StatusSwitchingProtocols {
		t.Fatalf("allowed Origin = %d, want 101", code)
	}
	// The callback owns the missing-Origin decision too.
	if code := rawHandshake(t, pattern, ""); code != http.StatusSwitchingProtocols {
		t.Fatalf("missing Origin allowed by callback = %d, want 101", code)
	}
	if atomic.LoadInt32(&rejected) != 1 {
		t.Fatalf("OnReject(origin) called %d times, want 1", rejected)
	}
}

func TestSocketMaxMessageBytes(t *testing.T) {
	s := newTestSocket()
	pattern := uniquePattern("maxbytes")
	got := make(chan int, 4)
	disconnected := make(chan struct{}, 1)
	s.OnSocketBaseWithOptions(pattern, SocketOptions{MaxMessageBytes: 1024}, func(c *SocketClient) {
		c.OnSocketAction("echo", func(p map[string]interface{}) {
			s, _ := p["s"].(string)
			got <- len(s)
		})
		c.OnSocketAction("onDisconnect", func(p map[string]interface{}) {
			disconnected <- struct{}{}
		})
	})

	ws, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	send(t, ws, "echo", map[string]interface{}{"s": strings.Repeat("a", 500)})
	select {
	case n := <-got:
		if n != 500 {
			t.Fatalf("echo len = %d", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("small message not delivered")
	}

	send(t, ws, "echo", map[string]interface{}{"s": strings.Repeat("a", 4096)})
	expectClosed(t, ws, 3*time.Second)
	select {
	case <-disconnected:
	case <-time.After(3 * time.Second):
		t.Fatal("onDisconnect not called after oversized frame")
	}
	select {
	case n := <-got:
		t.Fatalf("oversized message delivered (%d bytes)", n)
	default:
	}
}

func TestSocketAuthTimeout(t *testing.T) {
	s := newTestSocket()
	pattern := uniquePattern("auth")
	var timeouts int32
	s.OnSocketBaseWithOptions(pattern, SocketOptions{
		AuthTimeout: 300 * time.Millisecond,
		OnReject: func(r *http.Request, reason string) {
			if reason == "auth-timeout" {
				atomic.AddInt32(&timeouts, 1)
			}
		},
	}, func(c *SocketClient) {
		c.OnSocketAction("auth", func(p map[string]interface{}) {
			if p["token"] == "good" {
				c.SetAuthenticated()
				c.Send("authed", JSONC{})
			}
		})
	})

	// Never authenticates: closed after the timeout, even though it pings.
	anon, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer anon.Close()
	expectAlive(t, anon)
	send(t, anon, "auth", map[string]interface{}{"token": "bad"})
	expectClosed(t, anon, 3*time.Second)

	// Authenticates in time: still open well after the timeout.
	ok, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer ok.Close()
	send(t, ok, "auth", map[string]interface{}{"token": "good"})
	time.Sleep(700 * time.Millisecond)
	expectAlive(t, ok)

	if n := atomic.LoadInt32(&timeouts); n != 1 {
		t.Fatalf("auth-timeout rejections = %d, want 1", n)
	}
}

func TestSocketMaxConnsPerIP(t *testing.T) {
	s := newTestSocket()
	pattern := uniquePattern("conns")
	opts := SocketOptions{MaxConnsPerIP: 2}
	s.OnSocketBaseWithOptions(pattern, opts, func(c *SocketClient) {})
	route := s.routes[len(s.routes)-1]

	a, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	expectAlive(t, a)
	expectAlive(t, b)

	c, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	expectClosed(t, c, 3*time.Second)

	// Closing one frees its slot.
	b.Close()
	deadline := time.Now().Add(3 * time.Second)
	for route.limiter.count("127.0.0.1") > 1 {
		if time.Now().After(deadline) {
			t.Fatalf("slot not released: %d", route.limiter.count("127.0.0.1"))
		}
		time.Sleep(10 * time.Millisecond)
	}
	d, err := dial(t, pattern, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	expectAlive(t, d)
}

func TestSocketMaxConnsCustomKey(t *testing.T) {
	s := newTestSocket()
	pattern := uniquePattern("connkey")
	s.OnSocketBaseWithOptions(pattern, SocketOptions{
		MaxConnsPerIP: 1,
		ClientKey:     func(r *http.Request) string { return r.URL.Query().Get("k") },
	}, func(c *SocketClient) {})

	a, err := dial(t, pattern+"?k=one", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := dial(t, pattern+"?k=two", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	expectAlive(t, a)
	expectAlive(t, b)

	c, err := dial(t, pattern+"?k=one", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	expectClosed(t, c, 3*time.Second)
}

func TestRandStringRunes(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		s := RandStringRunes(32)
		if len(s) != 32 {
			t.Fatalf("len = %d", len(s))
		}
		for _, r := range s {
			if !strings.ContainsRune(string(letterRunes), r) {
				t.Fatalf("unexpected rune %q", r)
			}
		}
		if seen[s] {
			t.Fatalf("duplicate id %s", s)
		}
		seen[s] = true
	}
}
