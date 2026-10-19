package kube

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func writeTestKubeconfig(t *testing.T, context string) string {
	t.Helper()
	cfg := api.NewConfig()
	cfg.CurrentContext = context
	cfg.Clusters[context] = api.NewCluster()
	cfg.Clusters[context].Server = "https://127.0.0.1:6443"
	cfg.AuthInfos[context] = api.NewAuthInfo()
	cfg.Contexts[context] = api.NewContext()
	cfg.Contexts[context].AuthInfo = context
	cfg.Contexts[context].Cluster = context

	path := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, clientcmd.WriteToFile(*cfg, path))
	return path
}

func TestClient(t *testing.T) {
	path := writeTestKubeconfig(t, "test-ctx")

	c, err := Client("test-ctx", path)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestClientUnknownContext(t *testing.T) {
	path := writeTestKubeconfig(t, "test-ctx")

	_, err := Client("absent-ctx", path)
	require.Error(t, err)
}

func TestClientMissingKubeconfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent")

	_, err := Client("test-ctx", path)
	require.Error(t, err)
}
