package kube

import (
	"testing"

	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func makeNode(name string, labels map[string]string, addresses []v1.NodeAddress, ready v1.ConditionStatus, kubeletVersion string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Status: v1.NodeStatus{
			Addresses: addresses,
			NodeInfo: v1.NodeSystemInfo{
				KubeletVersion:          kubeletVersion,
				ContainerRuntimeVersion: "containerd://1.7.0",
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: ready,
				},
			},
		},
	}
}

func TestDescribeNodes(t *testing.T) {
	master := makeNode("master-1",
		map[string]string{"node-role.kubernetes.io/master": "", "node-role.kubernetes.io/etcd": ""},
		[]v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
			{Type: v1.NodeHostName, Address: "master-1"},
		}, v1.ConditionTrue, "v1.30.11+k3s1")
	// matched by external IP, no role labels, not ready
	worker := makeNode("ip-10-0-0-2", nil,
		[]v1.NodeAddress{
			{Type: v1.NodeExternalIP, Address: "1.1.1.2"},
			{Type: v1.NodeHostName, Address: "worker-2"},
		}, v1.ConditionFalse, "v1.30.4+k3s1")
	// matched only because its name equals the instance ID
	byName := makeNode("ip-10-0-0-3", nil,
		[]v1.NodeAddress{{Type: v1.NodeInternalIP, Address: "10.0.0.3"}},
		v1.ConditionTrue, "v1.30.4+k3s1")

	client := fake.NewSimpleClientset(master, worker, byName)

	instanceNodes := []types.ClusterNode{
		{InstanceID: "i-1", InternalIP: []string{"10.0.0.1"}, Status: types.ClusterStatusUnknown},
		{InstanceID: "i-2", ExternalIP: []string{"1.1.1.2"}, Status: types.ClusterStatusUnknown},
		{InstanceID: "ip-10-0-0-3", Status: types.ClusterStatusUnknown},
		{InstanceID: "i-4", Status: types.ClusterStatusUnknown},
	}

	result, err := DescribeNodes(client, instanceNodes)
	require.NoError(t, err)
	require.Len(t, result, 4)

	// matched by internal IP, roles joined in label order
	assert.Equal(t, "master-1", result[0].HostName)
	assert.Equal(t, "v1.30.11+k3s1", result[0].Version)
	assert.Equal(t, "containerd://1.7.0", result[0].ContainerRuntimeVersion)
	assert.Equal(t, "etcd,master", result[0].Roles)
	assert.Equal(t, "Ready", result[0].Status)

	// matched by external IP, no roles reported as <none>
	assert.Equal(t, "worker-2", result[1].HostName)
	assert.Equal(t, "<none>", result[1].Roles)
	assert.Equal(t, "NotReady", result[1].Status)

	// matched by instance ID equal to node name
	assert.Equal(t, "Ready", result[2].Status)
	assert.Equal(t, "<none>", result[2].Roles)

	// instance without a Kubernetes counterpart is left untouched
	assert.Equal(t, types.ClusterStatusUnknown, result[3].Status)
	assert.Empty(t, result[3].Version)
}
