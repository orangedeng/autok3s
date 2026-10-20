package base

import (
	"testing"

	"github.com/orangedeng/autok3s/pkg/cluster/kube"
	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestReconcileNodes(t *testing.T) {
	master := types.Node{Master: true, InstanceID: "i-m1", InternalIPAddress: []string{"10.0.0.1"}}
	worker := types.Node{InstanceID: "i-w1", InternalIPAddress: []string{"10.0.0.2"}}
	legacyWorker := types.Node{InstanceID: "i-w2", InternalIPAddress: []string{"10.0.0.3"}}

	labelledView := kube.NodeView{
		Name:       "worker-1",
		Labels:     map[string]string{"autok3s.io/install-uuid": "uuid-1", "autok3s.io/instance-id": "i-w1"},
		InternalIP: "10.0.0.2",
	}
	legacyView := kube.NodeView{Name: "ip-10-0-0-3", Labels: map[string]string{}, InternalIP: "10.0.0.3"}
	foreignView := kube.NodeView{
		Name:       "foreign-1",
		Labels:     map[string]string{"autok3s.io/install-uuid": "uuid-2", "autok3s.io/instance-id": "i-w1"},
		InternalIP: "10.0.0.9",
	}

	masters := []types.Node{master}
	workers := []types.Node{worker, legacyWorker}

	t.Run("provider view has every instance", func(t *testing.T) {
		result := reconcileNodes("uuid-1", masters, workers, []kube.NodeView{labelledView, legacyView},
			map[string]bool{"i-m1": true, "i-w1": true, "i-w2": true})
		assert.False(t, result.Changed)
		assert.Len(t, result.MasterNodes, 1)
		assert.Len(t, result.WorkerNodes, 2)
		assert.Empty(t, result.OrphanNodeNames)
	})

	t.Run("worker gone from provider is pruned with its node object", func(t *testing.T) {
		result := reconcileNodes("uuid-1", masters, workers, []kube.NodeView{labelledView, legacyView},
			map[string]bool{"i-m1": true, "i-w2": true})
		assert.True(t, result.Changed)
		assert.Len(t, result.WorkerNodes, 1)
		assert.Equal(t, "i-w2", result.WorkerNodes[0].InstanceID)
		// the label matched the object of the removed instance
		assert.Equal(t, []string{"worker-1"}, result.OrphanNodeNames)
	})

	t.Run("master gone from provider is pruned without touching its object", func(t *testing.T) {
		result := reconcileNodes("uuid-1", masters, workers, nil,
			map[string]bool{"i-w1": true, "i-w2": true})
		assert.True(t, result.Changed)
		assert.Empty(t, result.MasterNodes)
		assert.Empty(t, result.OrphanNodeNames)
	})

	t.Run("legacy node without labels matches by address", func(t *testing.T) {
		result := reconcileNodes("uuid-1", masters, workers, []kube.NodeView{legacyView},
			map[string]bool{"i-m1": true, "i-w1": true})
		assert.True(t, result.Changed)
		assert.Equal(t, []string{"ip-10-0-0-3"}, result.OrphanNodeNames)
	})

	t.Run("empty provider view proves nothing", func(t *testing.T) {
		result := reconcileNodes("uuid-1", masters, workers, []kube.NodeView{labelledView}, nil)
		assert.False(t, result.Changed)
		assert.Len(t, result.MasterNodes, 1)
		assert.Len(t, result.WorkerNodes, 2)
	})

	t.Run("nodes of another installation never match", func(t *testing.T) {
		result := reconcileNodes("uuid-1", masters, workers, []kube.NodeView{foreignView},
			map[string]bool{"i-m1": true, "i-w2": true})
		assert.True(t, result.Changed)
		// i-w1 is gone from the provider and its object belongs to another
		// installation, so the state entry goes but the object stays.
		assert.Empty(t, result.OrphanNodeNames)
	})

	t.Run("instance present in provider survives without apiserver match", func(t *testing.T) {
		// an instance whose agent has not registered yet is not pruned.
		result := reconcileNodes("uuid-1", masters, workers, nil,
			map[string]bool{"i-m1": true, "i-w1": true, "i-w2": true})
		assert.False(t, result.Changed)
	})
}

func TestMatchNodeView(t *testing.T) {
	node := types.Node{InstanceID: "i-w1", InternalIPAddress: []string{"10.0.0.2"}}

	// the identity label wins over the address
	view, ok := matchNodeView(node, []kube.NodeView{
		{Name: "other", Labels: map[string]string{"autok3s.io/instance-id": "i-w1"}, InternalIP: "10.9.9.9"},
	})
	assert.True(t, ok)
	assert.Equal(t, "other", view.Name)

	// a view carrying a different instance id does not fall back to
	// address matching, so a mislabelled object never owns the entry
	_, ok = matchNodeView(node, []kube.NodeView{
		{Labels: map[string]string{"autok3s.io/instance-id": "i-other"}, InternalIP: "10.0.0.2"},
	})
	assert.False(t, ok)

	// label-free views match by internal address
	view, ok = matchNodeView(node, []kube.NodeView{{Name: "ip-10-0-0-2", InternalIP: "10.0.0.2"}})
	assert.True(t, ok)
	assert.Equal(t, "ip-10-0-0-2", view.Name)

	// k3d and native nodes carry the instance id as the node name
	_, ok = matchNodeView(types.Node{InstanceID: "k3d-demo-agent-0"}, []kube.NodeView{{Name: "k3d-demo-agent-0"}})
	assert.True(t, ok)
}
