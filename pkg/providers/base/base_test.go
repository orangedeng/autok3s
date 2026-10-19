package base

import (
	"testing"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestGetClusterStatusMissingKubeconfig(t *testing.T) {
	p := NewBaseProvider()

	// without a kubeconfig the simple status leaves the status unset and
	// reports the persisted node counts; the detailed describe reports the
	// cluster as missing instead.
	info := p.GetClusterStatus("", &types.ClusterInfo{}, nil)
	assert.Empty(t, info.Status)
	assert.Equal(t, "0", info.Master)

	detail := p.Describe("", &types.ClusterInfo{}, nil)
	assert.Equal(t, common.StatusMissing, detail.Status)
}
