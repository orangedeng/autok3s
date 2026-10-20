package cmd

import "testing"

func TestIsLoopbackHost(t *testing.T) {
	for _, c := range []struct {
		host string
		want bool
	}{
		// an empty host binds every interface, not just loopback.
		{"", false},
		{"127.0.0.1", true},
		// the whole 127.0.0.0/8 range is loopback.
		{"127.8.8.8", true},
		{"::1", true},
		{"localhost", true},
		{"LOCALHOST", true},
		{"0.0.0.0", false},
		{"192.168.1.10", false},
		// a hostname that is not "localhost" cannot be resolved here and is
		// treated as non-loopback so the warning fires.
		{"example.com", false},
	} {
		if got := isLoopbackHost(c.host); got != c.want {
			t.Errorf("isLoopbackHost(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}
