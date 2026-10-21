package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstructRegex(t *testing.T) {
	// the % placeholder matches exactly one DNS label.
	r := constructRegex("ec2.%.amazonaws.com")

	assert.True(t, r.MatchString("ec2.us-west-1.amazonaws.com"))
	assert.False(t, r.MatchString("ec2.amazonaws.com"), "placeholder must not be optional")
	assert.False(t, r.MatchString("s3.us-west-1.amazonaws.com"), "literal labels must match exactly")
	assert.False(t, r.MatchString("xec2.us-west-1.amazonaws.com"), "match must be anchored")
	assert.False(t, r.MatchString("ec2.us-west-1.amazonaws.com.cn"), "match must be anchored at the end")

	// regex metacharacters in whitelist entries are quoted.
	quoted := constructRegex("a.b.c")
	assert.False(t, quoted.MatchString("axbxc"), "dots must be literal")
	assert.True(t, quoted.MatchString("a.b.c"))
}

func TestIsAllowed(t *testing.T) {
	t.Setenv("AUTOK3S_WHITELIST_DOMAIN", "example.com,*.foo.com,ec2.%.amazonaws.com")

	p := &proxy{prefix: "/proxy/"}

	cases := []struct {
		host    string
		allowed bool
	}{
		{"example.com", true},
		{"sub.example.com", false},
		{"notexample.com", false},
		{"a.foo.com", true},
		{"a.b.foo.com", true},
		{"foo.com", false},
		{"evilfoo.com", false},
		{"ec2.us-west-1.amazonaws.com", true},
		{"ec2.amazonaws.com", false},
		{"evil.com", false},
		{"", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.allowed, p.isAllowed(c.host), "host %q", c.host)
	}
}

func TestEmptyWhitelistAllowsNothing(t *testing.T) {
	t.Setenv("AUTOK3S_WHITELIST_DOMAIN", "")

	p := &proxy{prefix: "/proxy/"}
	assert.False(t, p.isAllowed("example.com"))
}

func TestProxyRewritesRequest(t *testing.T) {
	t.Setenv("AUTOK3S_WHITELIST_DOMAIN", "example.com")

	p := &proxy{prefix: "/proxy/"}

	cases := []struct {
		name     string
		target   string
		wantURL  string
		wantHost string
		wantErr  bool
	}{
		{
			name:     "scheme-less target defaults to https",
			target:   "/meta/proxy/example.com/path?a=b",
			wantURL:  "https://example.com/path?a=b",
			wantHost: "example.com",
		},
		{
			name:     "single slash https target is fixed up",
			target:   "/meta/proxy/https:/example.com/path",
			wantURL:  "https://example.com/path",
			wantHost: "example.com",
		},
		{
			name:     "single slash http target is fixed up",
			target:   "/meta/proxy/http:/example.com/path",
			wantURL:  "http://example.com/path",
			wantHost: "example.com",
		},
		{
			// the UI only issues single-slash targets (http:/host); a double
			// slash target gets a second scheme prepended, fails the host
			// check and is rejected instead of proxied.
			name:    "double slash scheme target is rejected",
			target:  "/meta/proxy/https://example.com/path",
			wantErr: true,
		},
		{
			name:    "host not in whitelist is rejected",
			target:  "/meta/proxy/https://evil.com/path",
			wantErr: true,
		},
		{
			name:    "empty host is rejected",
			target:  "/meta/proxy/",
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.target, nil)
			err := p.proxy(req)
			if c.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.wantURL, req.URL.String())
			assert.Equal(t, c.wantHost, req.Host)
		})
	}
}

func TestProxySetsForwardedProtoForTLS(t *testing.T) {
	t.Setenv("AUTOK3S_WHITELIST_DOMAIN", "example.com")

	p := &proxy{prefix: "/proxy/"}
	req := httptest.NewRequest(http.MethodGet, "/meta/proxy/example.com/path", nil)
	req.TLS = &tls.ConnectionState{}

	require.NoError(t, p.proxy(req))
	assert.Equal(t, "https", req.Header.Get(forwardProto))
}

func TestUnescapePath(t *testing.T) {
	const escaped = "https%3A%2F%2Fexample.com%2Fpath"

	t.Setenv("AUTOK3S_DEV_MODE", "")
	unchanged, err := unescapePath(escaped)
	require.NoError(t, err)
	assert.Equal(t, escaped, unchanged)

	t.Setenv("AUTOK3S_DEV_MODE", "true")
	unescaped, err := unescapePath(escaped)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/path", unescaped)

	_, err = unescapePath("%zz")
	require.Error(t, err)
}
