package kube

import (
	"path/filepath"
	"testing"

	"github.com/orangedeng/autok3s/pkg/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/client-go/tools/clientcmd"
)

// k3sMasterConfig is the kubeconfig shape written by a fresh K3s master:
// everything is named "default" and the server points to localhost.
const k3sMasterConfig = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: default
contexts:
- context:
    cluster: default
    user: default
  name: default
current-context: default
users:
- name: default
  user:
    token: token
`

func TestSaveConfig(t *testing.T) {
	origCfgPath, origFileManager := common.CfgPath, common.FileManager
	common.CfgPath = t.TempDir()
	common.FileManager = &common.ConfigFileManager{}
	t.Cleanup(func() {
		common.CfgPath = origCfgPath
		common.FileManager = origFileManager
	})

	require.NoError(t, SaveConfig(k3sMasterConfig, "1.2.3.4", "ctx-a"))

	loaded, err := clientcmd.LoadFromFile(filepath.Join(common.CfgPath, common.KubeCfgFile))
	require.NoError(t, err)
	require.NotNil(t, loaded.Contexts["ctx-a"])
	// the localhost server address must be rewritten to the public IP and the
	// default names to the cluster's context
	assert.Equal(t, "https://1.2.3.4:6443", loaded.Clusters["ctx-a"].Server)

	// saving the same context again replaces it instead of failing: the file
	// manager removes the context before merging the new one.
	require.NoError(t, SaveConfig(k3sMasterConfig, "5.6.7.8", "ctx-a"))
	loaded, err = clientcmd.LoadFromFile(filepath.Join(common.CfgPath, common.KubeCfgFile))
	require.NoError(t, err)
	assert.Equal(t, "https://5.6.7.8:6443", loaded.Clusters["ctx-a"].Server)
}
