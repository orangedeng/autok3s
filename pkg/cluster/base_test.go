package cluster

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/providers"
	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// listTestProvider implements only the methods describeClusterState relies on.
// The remaining providers.Provider methods stay unimplemented so unexpected
// calls fail loudly instead of silently returning zero values.
type listTestProvider struct {
	providers.Provider
	name     string
	meta     *types.Metadata
	block    time.Duration
	exist    bool
	existErr error
}

func (p *listTestProvider) GetProviderName() string {
	return p.name
}

func (p *listTestProvider) GenerateClusterName() string {
	return "list-test." + p.name
}

func (p *listTestProvider) SetMetadata(config *types.Metadata) {
	if config != nil {
		p.meta = config
	}
}

func (p *listTestProvider) SetOptions(_ []byte) error {
	return nil
}

func (p *listTestProvider) IsClusterExist() (bool, []string, error) {
	if p.existErr != nil {
		return false, nil, p.existErr
	}
	return p.exist, nil, nil
}

// GetCluster simulates the remote probe: real providers return immediately
// when kubecfg is empty, only a live probe can block.
func (p *listTestProvider) GetCluster(kubecfg string) *types.ClusterInfo {
	if kubecfg != "" {
		time.Sleep(p.block)
	}
	name := p.name
	if p.meta != nil && p.meta.Name != "" {
		name = p.meta.Name
	}
	return &types.ClusterInfo{ID: "list-test." + p.name, Name: name, Provider: p.name}
}

func setupListTest(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "autok3s.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&common.ClusterState{}))
	origDB, origCfgPath := common.DefaultDB, common.CfgPath
	common.DefaultDB = &common.Store{DB: db}
	// Point the kubeconfig lookup at an empty temp dir so no cluster of the
	// host running the tests is ever contacted.
	common.CfgPath = t.TempDir()
	t.Cleanup(func() {
		common.DefaultDB = origDB
		common.CfgPath = origCfgPath
	})
}

func seedClusterState(t *testing.T, name, provider, status string) {
	t.Helper()
	state := &common.ClusterState{
		Metadata: types.Metadata{
			Name:        name,
			Provider:    provider,
			Master:      "1",
			Worker:      "2",
			ContextName: "list-test." + name,
		},
		Status: status,
	}
	require.NoError(t, common.DefaultDB.DB.Create(state).Error)
}

func clusterByName(t *testing.T, list []*types.ClusterInfo, name string) *types.ClusterInfo {
	t.Helper()
	for _, c := range list {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cluster %s not found in list", name)
	return nil
}

func TestListClustersReturnsEveryCluster(t *testing.T) {
	setupListTest(t)
	providers.RegisterProvider("list-test-fast", func() (providers.Provider, error) {
		return &listTestProvider{name: "list-test-fast", exist: true}, nil
	})
	seedClusterState(t, "fast", "list-test-fast", common.StatusRunning)
	seedClusterState(t, "unknown-provider", "no-such-provider", common.StatusRunning)

	list, err := ListClusters("")
	require.NoError(t, err)
	require.Len(t, list, 2)

	fast := clusterByName(t, list, "fast")
	assert.Equal(t, "list-test-fast", fast.Provider)

	// A cluster whose provider is not registered must still be listed, built
	// from its persisted state.
	unknown := clusterByName(t, list, "unknown-provider")
	assert.Equal(t, common.StatusUnknown, unknown.Status)
	assert.Equal(t, "no-such-provider", unknown.Provider)
	assert.Equal(t, "1", unknown.Master)
	assert.Equal(t, "2", unknown.Worker)
}

func TestListClustersBoundedProbe(t *testing.T) {
	setupListTest(t)
	providers.RegisterProvider("list-test-slow", func() (providers.Provider, error) {
		return &listTestProvider{name: "list-test-slow", exist: true, block: 500 * time.Millisecond}, nil
	})
	seedClusterState(t, "slow", "list-test-slow", common.StatusRunning)
	seedClusterState(t, "creating", "list-test-slow", common.StatusCreating)

	origTimeout := listProbeTimeout
	listProbeTimeout = 50 * time.Millisecond
	t.Cleanup(func() { listProbeTimeout = origTimeout })

	start := time.Now()
	list, err := ListClusters("")
	elapsed := time.Since(start)
	require.NoError(t, err)
	require.Len(t, list, 2)

	// The unreachable cluster must not block the list: it degrades to its
	// persisted state well before its probe would finish.
	assert.Less(t, elapsed, 2*time.Second)
	slow := clusterByName(t, list, "slow")
	assert.Equal(t, common.StatusUnknown, slow.Status)
	assert.Equal(t, "list-test.slow", slow.ID)

	// Clusters that need no live probe keep reporting their persisted status.
	creating := clusterByName(t, list, "creating")
	assert.Equal(t, common.StatusCreating, creating.Status)
}
