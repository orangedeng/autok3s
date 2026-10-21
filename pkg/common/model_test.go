package common

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestStore builds a store backed by a throwaway sqlite file, using the
// same schema statements and migrations as InitStorage.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	require.NoError(t, err)

	setup(db)
	require.NoError(t, db.AutoMigrate(
		&ClusterState{},
		&Template{},
		&Package{},
		&Setting{},
		&SSHKey{},
		&Addon{},
	))

	return &Store{DB: db, broadcaster: NewBroadcaster()}
}

func marshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	out, err := json.Marshal(v)
	require.NoError(t, err)
	return out
}

func TestCredentialIsSinglePerProvider(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.CreateCredential(&Credential{Provider: "alibaba", Secrets: marshal(t, map[string]string{"access-key": "first"})}))

	list, err := store.GetCredentialByProvider("alibaba")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, marshal(t, map[string]string{"access-key": "first"}), []byte(list[0].Secrets))

	// a second credential for the same provider replaces the stored secrets
	// instead of adding a second row.
	require.NoError(t, store.CreateCredential(&Credential{Provider: "alibaba", Secrets: marshal(t, map[string]string{"access-key": "second"})}))
	list, err = store.GetCredentialByProvider("alibaba")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, marshal(t, map[string]string{"access-key": "second"}), []byte(list[0].Secrets))

	// credentials of other providers are unaffected.
	require.NoError(t, store.CreateCredential(&Credential{Provider: "native", Secrets: marshal(t, map[string]string{"secret": "native"})}))
	all, err := store.ListCredential()
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestCredentialCRUD(t *testing.T) {
	store := newTestStore(t)

	created := &Credential{Provider: "alibaba", Secrets: marshal(t, map[string]string{"access-key": "a"})}
	require.NoError(t, store.CreateCredential(created))
	require.NotZero(t, created.ID)

	cred, err := store.GetCredential(created.ID)
	require.NoError(t, err)
	require.NotNil(t, cred)
	assert.Equal(t, "alibaba", cred.Provider)

	// missing ids resolve to nil without an error.
	missing, err := store.GetCredential(created.ID + 100)
	require.NoError(t, err)
	assert.Nil(t, missing)

	updated := &Credential{ID: created.ID, Provider: "alibaba", Secrets: marshal(t, map[string]string{"access-key": "b"})}
	require.NoError(t, store.UpdateCredential(updated))
	cred, err = store.GetCredential(created.ID)
	require.NoError(t, err)
	assert.Equal(t, marshal(t, map[string]string{"access-key": "b"}), []byte(cred.Secrets))

	require.NoError(t, store.DeleteCredential(created.ID))
	all, err := store.ListCredential()
	require.NoError(t, err)
	assert.Empty(t, all)

	// deleting an unknown id is not an error.
	assert.NoError(t, store.DeleteCredential(created.ID))
}

func TestClusterCRUD(t *testing.T) {
	store := newTestStore(t)

	cluster := &types.Cluster{
		Metadata: types.Metadata{Name: "c1", Provider: "native", ContextName: "ctx-c1", Master: "1", Worker: "1"},
		Status: types.Status{
			Status:      StatusRunning,
			MasterNodes: []types.Node{{InstanceID: "m-1", PublicIPAddress: []string{"192.0.2.10"}}},
			WorkerNodes: []types.Node{{InstanceID: "w-1", PublicIPAddress: []string{"192.0.2.11"}}},
		},
	}
	require.NoError(t, store.SaveCluster(cluster))

	state, err := store.GetCluster("c1", "native")
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, StatusRunning, state.Status)
	assert.Equal(t, "ctx-c1", state.ContextName)

	masterNodes := make([]types.Node, 0)
	require.NoError(t, json.Unmarshal(state.MasterNodes, &masterNodes))
	require.Len(t, masterNodes, 1)
	assert.Equal(t, "m-1", masterNodes[0].InstanceID)

	workerNodes := make([]types.Node, 0)
	require.NoError(t, json.Unmarshal(state.WorkerNodes, &workerNodes))
	require.Len(t, workerNodes, 1)
	assert.Equal(t, "w-1", workerNodes[0].InstanceID)

	// saving an existing cluster updates it in place instead of duplicating it.
	cluster.Status.Status = StatusFailed
	require.NoError(t, store.SaveCluster(cluster))
	list, err := store.ListCluster("native")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, StatusFailed, list[0].Status)

	// clusters can be looked up by context name and across providers.
	byID, err := store.GetClusterByID("ctx-c1")
	require.NoError(t, err)
	require.NotNil(t, byID)
	found, err := store.FindCluster("c1", "")
	require.NoError(t, err)
	assert.Len(t, found, 1)

	require.NoError(t, store.DeleteCluster("c1", "native"))
	state, err = store.GetCluster("c1", "native")
	require.NoError(t, err)
	assert.Nil(t, state)

	// deleting an unknown cluster is a no-op.
	assert.NoError(t, store.DeleteCluster("missing", "native"))
}

func TestGetClusterMissing(t *testing.T) {
	store := newTestStore(t)

	state, err := store.GetCluster("missing", "native")
	require.NoError(t, err)
	assert.Nil(t, state)

	byID, err := store.GetClusterByID("missing")
	require.NoError(t, err)
	assert.Nil(t, byID)
}
