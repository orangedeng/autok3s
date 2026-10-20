package websocket

import (
	"net/http"
	"net/url"
)

// CheckSameOrigin rejects websocket upgrades whose Origin header points at
// another site. Browsers always set Origin on an upgrade request, so a
// mismatch means a foreign page is trying to ride the single-user API;
// non-browser clients omit the header and pass through.
func CheckSameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Host == r.Host
}
