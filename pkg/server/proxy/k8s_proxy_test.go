package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: c
  cluster:
    server: https://127.0.0.1:6443
    insecure-skip-tls-verify: true
contexts:
- name: c
  context:
    cluster: c
    user: u
current-context: c
users:
- name: u
  user: {}
`

func TestGetTransportForConfigCachesByKubeconfig(t *testing.T) {
	first, err := getTransportForConfig(testKubeconfig)
	require.NoError(t, err)

	second, err := getTransportForConfig(testKubeconfig)
	require.NoError(t, err)
	assert.Same(t, first, second, "identical kubeconfig must reuse the cached transport")

	_, err = getTransportForConfig(strings.Replace(testKubeconfig, "6443", "6444", 1))
	require.NoError(t, err)
}

func TestGetTransportForConfigInvalidKubeconfig(t *testing.T) {
	_, err := getTransportForConfig("::not-a-kubeconfig::")
	require.Error(t, err)
}

func TestK8sProxyHandlerRejectsInvalidConfig(t *testing.T) {
	handler := NewK8sProxy()

	req := httptest.NewRequest(http.MethodGet, "/k8s/proxy/api/v1/nodes", nil)
	// the value is not a kubeconfig and not valid base64, so it must fail to
	// parse as a kubeconfig instead of being proxied anywhere.
	req.Header.Set("config", "not-a-kubeconfig{{{")
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.NotEmpty(t, rw.Body.String())
}
