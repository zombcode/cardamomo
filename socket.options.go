package cardamomo

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// SocketOptions hardens one socket base. The zero value changes nothing: the
// base behaves exactly as one registered with OnSocketBase.
type SocketOptions struct {
	// MaxMessageBytes caps the payload of one incoming frame. A client that
	// sends more is disconnected (onDisconnect fires as for any other read
	// error). Zero keeps the x/net default (websocket.DefaultMaxPayloadBytes,
	// 32 MB).
	MaxMessageBytes int

	// CheckOrigin decides whether the handshake is accepted; returning false
	// answers 403. It replaces x/net's default check, which accepts any
	// parseable Origin and refuses a request without one: a callback that
	// wants the same for a missing Origin must return false itself. Nil keeps
	// the default.
	CheckOrigin func(r *http.Request) bool

	// AuthTimeout closes the socket if SocketClient.SetAuthenticated has not
	// been called that long after the connection was accepted. Zero disables
	// it: a socket may stay anonymous for as long as it keeps pinging.
	AuthTimeout time.Duration

	// MaxConnsPerIP caps the simultaneous sockets of this base per client
	// key. A connection over the cap is closed right after the handshake,
	// before the base callback runs. Zero means no cap.
	MaxConnsPerIP int

	// ClientKey returns the key MaxConnsPerIP counts by. Nil uses the same IP
	// as SocketClient.GetIP, which trusts X-Real-IP and X-Forwarded-For as
	// sent: behind a proxy that does not set them to the real client (a CDN
	// in front of nginx, say), every client shares the proxy's address.
	ClientKey func(r *http.Request) string

	// OnReject, if set, is told about every connection these options refuse:
	// reason is "origin", "max-conns" or "auth-timeout".
	OnReject func(r *http.Request, reason string)
}

func (o SocketOptions) reject(r *http.Request, reason string) {
	if o.OnReject != nil {
		o.OnReject(r, reason)
	}
}

var errOriginRefused = errors.New("cardamomo: origin refused")

// handler builds the http.Handler for one base. Without CheckOrigin it is the
// same websocket.Handler cardamomo always registered.
func (o SocketOptions) handler(onConnected func(ws *websocket.Conn)) http.Handler {
	if o.CheckOrigin == nil {
		return websocket.Handler(onConnected)
	}
	return websocket.Server{
		Handler: onConnected,
		Handshake: func(config *websocket.Config, req *http.Request) error {
			if !o.CheckOrigin(req) {
				o.reject(req, "origin")
				return errOriginRefused
			}
			// Keep config.Origin filled when it parses, as the default check does.
			if origin, err := websocket.Origin(config, req); err == nil {
				config.Origin = origin
			}
			return nil
		},
	}
}

// connLimiter counts open sockets per client key.
type connLimiter struct {
	mu     sync.Mutex
	max    int
	counts map[string]int
}

func newConnLimiter(max int) *connLimiter {
	if max <= 0 {
		return nil
	}
	return &connLimiter{max: max, counts: make(map[string]int)}
}

func (l *connLimiter) acquire(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.counts[key] >= l.max {
		return false
	}
	l.counts[key]++
	return true
}

func (l *connLimiter) release(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.counts[key] <= 1 {
		delete(l.counts, key)
		return
	}
	l.counts[key]--
}

func (l *connLimiter) count(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.counts[key]
}
