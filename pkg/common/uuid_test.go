package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupInstallUUIDStable(t *testing.T) {
	// the uuid is assigned once and stays the same value afterwards.
	require.NoError(t, SetupInstallUUID())
	first := GetInstallUUID()
	require.NotEmpty(t, first)

	require.NoError(t, SetupInstallUUID())
	assert.Equal(t, first, GetInstallUUID())

	// the value looks like a random token, not a sequential counter.
	assert.Len(t, first, 32)
}
