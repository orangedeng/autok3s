package ssh

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupClusterDB points common.DefaultDB at a throwaway database holding the
// given cluster states.
func setupClusterDB(t *testing.T, clusters ...*common.ClusterState) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&common.ClusterState{}))
	for _, c := range clusters {
		require.NoError(t, db.Create(c).Error)
	}

	prev := common.DefaultDB
	common.DefaultDB = &common.Store{DB: db}
	t.Cleanup(func() { common.DefaultDB = prev })
}

func clusterState(contextName string, masters, workers []types.Node) *common.ClusterState {
	return &common.ClusterState{
		Metadata:    types.Metadata{Name: contextName, Provider: "native", ContextName: contextName},
		Status:      common.StatusRunning,
		MasterNodes: mustJSON(masters),
		WorkerNodes: mustJSON(workers),
	}
}

func mustJSON(v interface{}) []byte {
	out, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return out
}

func TestNewDialerClusterMissing(t *testing.T) {
	setupClusterDB(t)

	_, err := NewDialer("missing", "i-1", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster missing is not exist")
}

func TestNewDialerNodeNotFound(t *testing.T) {
	setupClusterDB(t, clusterState("c1",
		[]types.Node{{InstanceID: "i-master", PublicIPAddress: []string{"192.0.2.10"}}},
		[]types.Node{{InstanceID: "i-worker", PublicIPAddress: []string{"192.0.2.11"}}},
	))

	_, err := NewDialer("c1", "i-other", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "node i-other is not found for cluster [c1]")
}

func TestNewDialerMalformedNodes(t *testing.T) {
	state := &common.ClusterState{
		Metadata:    types.Metadata{Name: "c1", Provider: "native", ContextName: "c1"},
		MasterNodes: []byte("not-json"),
	}
	setupClusterDB(t, state)

	_, err := NewDialer("c1", "i-1", nil)
	require.Error(t, err)
}

func TestNewMasterDialerClusterMissing(t *testing.T) {
	setupClusterDB(t)

	_, err := NewMasterDialer("missing", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster missing is not exist")
}

func TestNewMasterDialerNoMasters(t *testing.T) {
	setupClusterDB(t, clusterState("c1", []types.Node{}, []types.Node{}))

	_, err := NewMasterDialer("c1", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster c1 has no master node")
}

func TestNewMasterDialerMalformedMasters(t *testing.T) {
	state := &common.ClusterState{
		Metadata:    types.Metadata{Name: "c1", Provider: "native", ContextName: "c1"},
		MasterNodes: []byte("not-json"),
	}
	setupClusterDB(t, state)

	_, err := NewMasterDialer("c1", nil)
	require.Error(t, err)
}
