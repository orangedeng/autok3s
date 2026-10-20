package websocket

import (
	"net/http"
	"testing"
)

func TestCheckSameOrigin(t *testing.T) {
	req := func(host, origin string) *http.Request {
		r := &http.Request{Host: host, Header: http.Header{}}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		return r
	}

	for _, c := range []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		// non-browser clients do not send Origin.
		{"no origin header", "127.0.0.1:8080", "", true},
		{"same host and port", "127.0.0.1:8080", "http://127.0.0.1:8080", true},
		{"same host no port", "localhost", "http://localhost", true},
		{"different host", "127.0.0.1:8080", "http://evil.example.com", false},
		{"different port", "127.0.0.1:8080", "http://127.0.0.1:5173", false},
		{"localhost vs loopback ip", "127.0.0.1:8080", "http://localhost:8080", false},
		// "null" origin is sent by sandboxed pages and must not pass.
		{"null origin", "127.0.0.1:8080", "null", false},
		{"malformed origin", "127.0.0.1:8080", "http://[::1", false},
	} {
		if got := CheckSameOrigin(req(c.host, c.origin)); got != c.want {
			t.Errorf("%s: CheckSameOrigin() = %v, want %v", c.name, got, c.want)
		}
	}
}
