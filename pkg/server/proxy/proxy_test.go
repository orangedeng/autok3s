package proxy

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteURL(t *testing.T) {
	tr := &Transport{Scheme: "http", Host: "proxy.local", PathPrepend: "/pfx"}
	source := &url.URL{Scheme: "http", Host: "origin.local", Path: "/page"}

	cases := []struct {
		name      string
		targetURL string
		want      string
	}{
		{"root relative path gets the proxy prefix", "/foo", "/pfx/foo"},
		{"same host absolute url is rewritten", "http://origin.local/foo", "/pfx/foo"},
		{"different host is untouched", "http://other.local/foo", "http://other.local/foo"},
		{"relative url is untouched", "page.html", "page.html"},
		{"trailing slash is preserved", "/foo/", "/pfx/foo/"},
		{"already prefixed path is not double prefixed", "/pfx/foo", "/pfx/foo"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, tr.rewriteURL(c.targetURL, source, "origin.local"))
		})
	}
}

func TestRewriteHTML(t *testing.T) {
	rewriter := func(s string) string { return "P:" + s }
	input := `<html><body><a href="/x">a</a><img src="/y.png"><span title="/z">s</span></body></html>`

	var out strings.Builder
	require.NoError(t, rewriteHTML(strings.NewReader(input), &out, rewriter))

	body := out.String()
	assert.Contains(t, body, `href="P:/x"`)
	assert.Contains(t, body, `src="P:/y.png"`)
	assert.NotContains(t, body, `P:/z`, "non url attributes must not be rewritten")
}

func htmlResponse(body string, headers map[string]string) *http.Response {
	h := http.Header{}
	h.Set("Content-Type", "text/html")
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRewriteResponsePlain(t *testing.T) {
	tr := &Transport{Scheme: "http", Host: "proxy.local", PathPrepend: "/pfx"}
	req := httptest.NewRequest(http.MethodGet, "/pfx/page", nil)

	resp, err := tr.rewriteResponse(req, htmlResponse(`<a href="/foo/">x</a>`, nil))
	require.NoError(t, err)

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(out), `href="/pfx/foo/"`)
	assert.Equal(t, int64(len(out)), resp.ContentLength, "content length must match the rewritten body")
}

func TestRewriteResponseGzip(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(`<a href="/foo/">x</a>`))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	tr := &Transport{Scheme: "http", Host: "proxy.local", PathPrepend: "/pfx"}
	req := httptest.NewRequest(http.MethodGet, "/pfx/page", nil)

	resp, err := tr.rewriteResponse(req, htmlResponse(buf.String(), map[string]string{"Content-Encoding": "gzip"}))
	require.NoError(t, err)
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	zr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	out, err := io.ReadAll(zr)
	require.NoError(t, err)
	assert.Contains(t, string(out), `href="/pfx/foo/"`)
}

func TestRewriteResponseDeflate(t *testing.T) {
	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.BestCompression)
	require.NoError(t, err)
	_, err = fw.Write([]byte(`<a href="/foo/">x</a>`))
	require.NoError(t, err)
	require.NoError(t, fw.Close())

	tr := &Transport{Scheme: "http", Host: "proxy.local", PathPrepend: "/pfx"}
	req := httptest.NewRequest(http.MethodGet, "/pfx/page", nil)

	resp, err := tr.rewriteResponse(req, htmlResponse(buf.String(), map[string]string{"Content-Encoding": "deflate"}))
	require.NoError(t, err)

	out, err := io.ReadAll(flate.NewReader(resp.Body))
	require.NoError(t, err)
	assert.Contains(t, string(out), `href="/pfx/foo/"`)
}

func TestRewriteResponseUnknownEncoding(t *testing.T) {
	// encodings the transport cannot decode are passed through untouched.
	tr := &Transport{Scheme: "http", Host: "proxy.local", PathPrepend: "/pfx"}
	req := httptest.NewRequest(http.MethodGet, "/pfx/page", nil)
	original := htmlResponse(`<a href="/foo/">x</a>`, map[string]string{"Content-Encoding": "br"})

	resp, err := tr.rewriteResponse(req, original)
	require.NoError(t, err)
	assert.Same(t, original, resp)

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `<a href="/foo/">x</a>`, string(out))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTransportRoundTripSetsForwardedHeaders(t *testing.T) {
	var seen http.Header
	tr := &Transport{
		Scheme:      "http",
		Host:        "origin.local",
		PathPrepend: "/pfx",
		RoundTripper: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			seen = r.Header.Clone()
			return htmlResponse("", nil), nil
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/pfx/page", nil)
	_, err := tr.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "/pfx/pfx/page", seen.Get("X-Forwarded-Uri"))
	assert.Equal(t, "origin.local/pfx", seen.Get("X-Forwarded-Host"))
	assert.Equal(t, "http", seen.Get("X-Forwarded-Proto"))
}

func TestTransportRoundTripRewritesRedirectLocation(t *testing.T) {
	tr := &Transport{
		Scheme:      "http",
		Host:        "origin.local",
		PathPrepend: "/pfx",
		RoundTripper: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			resp := htmlResponse("", nil)
			resp.StatusCode = http.StatusFound
			resp.Header.Set("Location", "/other/page")
			return resp, nil
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/pfx/page", nil)
	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	// the rewriter returns the path only, prefixed with the proxy prefix.
	assert.Equal(t, "/pfx/other/page", resp.Header.Get("Location"))
}

func TestTransportRoundTripPassesNonHTMLThrough(t *testing.T) {
	tr := &Transport{
		Scheme:      "http",
		Host:        "origin.local",
		PathPrepend: "/pfx",
		RoundTripper: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"href":"/untouched"}`)),
			}, nil
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "/pfx/data", nil)
	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"href":"/untouched"}`, string(out))
}

func TestRemoteHandlerRedirectsWhenLocationPathEmpty(t *testing.T) {
	handler := &RemoteHandler{Location: &url.URL{Scheme: "http", Host: "example.com"}}

	req := httptest.NewRequest(http.MethodGet, "/meta/proxy?k=v", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusMovedPermanently, rw.Code)
	assert.Equal(t, "/meta/proxy/?k=v", rw.Header().Get("Location"))
}
