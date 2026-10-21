package credential

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/providers"
	"github.com/orangedeng/autok3s/pkg/types/apis"
	// registers the alibaba provider so its credential flags are resolvable.
	_ "github.com/orangedeng/autok3s/pkg/providers/alibaba"

	"github.com/glebarez/sqlite"
	apitypes "github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/v2/pkg/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestStore points common.DefaultDB at a throwaway sqlite database holding
// the production credentials table.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS credentials
		(
			id integer not null primary key autoincrement,
			provider TEXT not null,
			secrets BLOB
		);`).Error)

	prev := common.DefaultDB
	common.DefaultDB = &common.Store{DB: db}
	t.Cleanup(func() { common.DefaultDB = prev })

	return &Store{}
}

func alibabaSecrets(t *testing.T) map[string]interface{} {
	t.Helper()

	p, err := providers.GetProvider("alibaba")
	require.NoError(t, err)

	flags := p.GetCredentialFlags()
	require.NotEmpty(t, flags)

	secrets := map[string]interface{}{}
	for _, f := range flags {
		secrets[f.Name] = "secret-value"
	}
	return secrets
}

func TestGenerateCredential(t *testing.T) {
	newTestStore(t)

	secrets := alibabaSecrets(t)

	// unknown providers are rejected.
	_, err := generateCredential(secrets, "no-such-provider")
	require.Error(t, err)

	// every credential flag declared by the provider must be present.
	_, err = generateCredential(map[string]interface{}{}, "alibaba")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing credential")

	c, err := generateCredential(secrets, "alibaba")
	require.NoError(t, err)
	assert.Equal(t, "alibaba", c.Provider)

	stored := map[string]string{}
	require.NoError(t, json.Unmarshal(c.Secrets, &stored))
	assert.Equal(t, "secret-value", stored["access-key"])
}

func TestCredentialStoreCRUD(t *testing.T) {
	store := newTestStore(t)
	schema := &apitypes.APISchema{Schema: &schemas.Schema{ID: "credential"}}

	secrets := alibabaSecrets(t)

	created, err := store.Create(nil, schema, apitypes.APIObject{
		Object: map[string]interface{}{"provider": "alibaba", "secrets": secrets},
	})
	require.NoError(t, err)
	assert.Equal(t, "credential", created.Type)
	assert.NotEmpty(t, created.ID)

	cred, ok := created.Object.(*apis.Credential)
	require.True(t, ok)
	assert.Equal(t, "alibaba", cred.Provider)
	assert.Equal(t, "secret-value", cred.Secrets["access-key"])

	list, err := store.List(nil, schema)
	require.NoError(t, err)
	require.Len(t, list.Objects, 1)

	got, err := store.ByID(nil, schema, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// non numeric and unknown ids surface as API errors.
	_, err = store.ByID(nil, schema, "not-a-number")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")

	_, err = store.ByID(nil, schema, "999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not exist")

	// updating replaces the stored secrets.
	secrets["access-key"] = "rotated"
	updated, err := store.Update(nil, schema, apitypes.APIObject{
		Object: map[string]interface{}{"provider": "alibaba", "secrets": secrets},
	}, created.ID)
	require.NoError(t, err)
	cred, ok = updated.Object.(*apis.Credential)
	require.True(t, ok)
	assert.Equal(t, "rotated", cred.Secrets["access-key"])

	_, err = store.Delete(nil, schema, created.ID)
	require.NoError(t, err)
	list, err = store.List(nil, schema)
	require.NoError(t, err)
	assert.Empty(t, list.Objects)
}

func TestCredentialStoreCreateRejectsUnknownProvider(t *testing.T) {
	store := newTestStore(t)
	schema := &apitypes.APISchema{Schema: &schemas.Schema{ID: "credential"}}

	_, err := store.Create(nil, schema, apitypes.APIObject{
		Object: map[string]interface{}{"provider": "no-such-provider", "secrets": map[string]interface{}{}},
	})
	require.Error(t, err)
}
