package native

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/types"
	nativetypes "github.com/orangedeng/autok3s/pkg/types/native"
	"github.com/orangedeng/autok3s/pkg/utils"

	"github.com/glebarez/sqlite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFlagDefinitions(t *testing.T) {
	p := newProvider()

	flagSets := map[string][]types.Flag{
		"create":     p.GetCreateFlags(),
		"join":       p.GetJoinFlags(),
		"option":     p.GetOptionFlags(),
		"delete":     p.GetDeleteFlags(),
		"ssh":        p.GetSSHFlags(),
		"credential": p.GetCredentialFlags(),
	}

	for set, flags := range flagSets {
		t.Run(set, func(t *testing.T) {
			seen := map[string]bool{}
			for _, f := range flags {
				if f.Name == "" {
					t.Fatalf("flag with empty name in %s flags", set)
				}
				if f.P == nil {
					t.Fatalf("flag %s has no value pointer", f.Name)
				}
				if seen[f.Name] {
					t.Fatalf("duplicate flag %s in %s flags", f.Name, set)
				}
				seen[f.Name] = true
			}

			// every declared flag must be bindable by the CLI, i.e. its value
			// type is one of the types ConvertFlags knows how to register.
			cmd := &cobra.Command{Use: "test"}
			utils.ConvertFlags(cmd, flags)
			for name := range seen {
				if cmd.Flags().Lookup(name) == nil {
					t.Fatalf("flag %s is not registered by ConvertFlags", name)
				}
			}
		})
	}
}

func TestUsageExamples(t *testing.T) {
	p := newProvider()
	for _, action := range []string{"create", "join", "delete", "ssh"} {
		if p.GetUsageExample(action) == "" {
			t.Fatalf("missing usage example for action %s", action)
		}
	}
}

// setupMergeDB persists a cluster state that MergeClusterOptions merges the
// current flags against.
func setupMergeDB(t *testing.T, state *common.ClusterState) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&common.ClusterState{}))
	require.NoError(t, db.Create(state).Error)

	prev := common.DefaultDB
	common.DefaultDB = &common.Store{DB: db}
	t.Cleanup(func() { common.DefaultDB = prev })
}

func mergeState(t *testing.T, options nativetypes.Options, masters, workers []types.Node) *common.ClusterState {
	t.Helper()

	state := &common.ClusterState{
		Metadata: types.Metadata{Name: "c1", Provider: "native", ContextName: "c1"},
		Status:   common.StatusRunning,
	}
	var err error
	state.Options, err = json.Marshal(options)
	require.NoError(t, err)
	state.MasterNodes, err = json.Marshal(masters)
	require.NoError(t, err)
	state.WorkerNodes, err = json.Marshal(workers)
	require.NoError(t, err)
	return state
}

func TestMergeClusterOptionsAppendsWithDedupe(t *testing.T) {
	setupMergeDB(t, mergeState(t,
		nativetypes.Options{MasterIps: "192.0.2.1", WorkerIps: "192.0.2.2"},
		[]types.Node{{InstanceID: "m-1", PublicIPAddress: []string{"192.0.2.1"}}},
		[]types.Node{{InstanceID: "w-1", PublicIPAddress: []string{"192.0.2.2"}}},
	))

	p := newProvider()
	p.Name = "c1"
	p.MasterIps = "192.0.2.9,192.0.2.1"
	p.WorkerIps = "192.0.2.3"

	require.NoError(t, p.MergeClusterOptions())

	// existing ips come first, new ips are appended, duplicates are dropped.
	assert.Equal(t, "192.0.2.1,192.0.2.9", p.MasterIps)
	assert.Equal(t, "192.0.2.2,192.0.2.3", p.WorkerIps)
}

func TestMergeClusterOptionsDropsStaleWorkerIPs(t *testing.T) {
	// 192.0.2.2 is stored as a worker ip but no longer backed by a known
	// worker node, so the merge drops it instead of rejoining it.
	setupMergeDB(t, mergeState(t,
		nativetypes.Options{WorkerIps: "192.0.2.2,192.0.2.4"},
		nil,
		[]types.Node{{InstanceID: "w-1", PublicIPAddress: []string{"192.0.2.4"}}},
	))

	p := newProvider()
	p.Name = "c1"
	p.WorkerIps = "192.0.2.5"

	require.NoError(t, p.MergeClusterOptions())
	assert.Equal(t, "192.0.2.4,192.0.2.5", p.WorkerIps)
}

func TestMergeClusterOptionsMissingCluster(t *testing.T) {
	setupMergeDB(t, &common.ClusterState{
		Metadata: types.Metadata{Name: "other", Provider: "native", ContextName: "other"},
	})

	p := newProvider()
	p.Name = "c1"
	p.MasterIps = "192.0.2.1"

	// no stored state means there is nothing to merge and no error.
	require.NoError(t, p.MergeClusterOptions())
	assert.Equal(t, "192.0.2.1", p.MasterIps)
}
