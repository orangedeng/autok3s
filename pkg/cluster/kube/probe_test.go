package kube

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func clientForURL(t *testing.T, url string) *kubernetes.Clientset {
	t.Helper()
	c, err := kubernetes.NewForConfig(&rest.Config{Host: url})
	require.NoError(t, err)
	return c
}

// newVersionServer serves a /readyz and /version endpoint like a K3s API
// server does, without any network access.
func newVersionServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Info{
			Major:      "1",
			Minor:      "30",
			GitVersion: "v1.30.11+k3s1",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestStatus(t *testing.T) {
	srv := newVersionServer(t)
	require.Equal(t, types.ClusterStatusRunning, Status(clientForURL(t, srv.URL)))
}

func TestStatusNotReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	require.Equal(t, types.ClusterStatusStopped, Status(clientForURL(t, srv.URL)))
}

func TestStatusUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()

	require.Equal(t, types.ClusterStatusStopped, Status(clientForURL(t, url)))
}

func TestVersion(t *testing.T) {
	srv := newVersionServer(t)
	require.Equal(t, "v1.30.11+k3s1", Version(clientForURL(t, srv.URL)))
}

func TestVersionUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()

	require.Equal(t, types.ClusterStatusUnknown, Version(clientForURL(t, url)))
}
