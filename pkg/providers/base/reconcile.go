package base

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/orangedeng/autok3s/pkg/cluster/kube"
	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/k3s"
	"github.com/orangedeng/autok3s/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// reconcileStateMu serializes the read-modify-write cycle of the node
// arrays in the persisted state; autok3s is single process, so one mutex
// covers concurrent describes of the same installation.
var reconcileStateMu sync.Mutex

// ReconciledNodes is the outcome of converging the persisted node lists
// with the apiserver and provider views.
type ReconciledNodes struct {
	MasterNodes []types.Node
	WorkerNodes []types.Node
	// OrphanNodeNames are worker Node objects whose backing instance is
	// gone from the provider; deleting them keeps the apiserver free of
	// dead worker entries. Master objects are left to the control plane
	// because removing them affects the etcd member view.
	OrphanNodeNames []string
	Changed         bool
}

// reconcileNodes prunes persisted nodes whose instance no longer exists in
// the provider view. The provider is the arbiter: a node missing from it is
// gone even when its Node object still lingers in the apiserver, while a
// node present in it survives even before its agent has registered. An
// empty provider view proves nothing, so nothing is pruned against it.
func reconcileNodes(installUUID string, masters, workers []types.Node, kubeNodes []kube.NodeView, cloudInstanceIDs map[string]bool) ReconciledNodes {
	if len(cloudInstanceIDs) == 0 {
		return ReconciledNodes{MasterNodes: masters, WorkerNodes: workers}
	}

	// nodes installed by another autok3s instance are not ours to reason
	// about: they neither match persisted entries nor become orphans.
	usable := make([]kube.NodeView, 0, len(kubeNodes))
	for _, view := range kubeNodes {
		if owner, ok := view.Labels[k3s.InstallUUIDLabel]; ok && owner != installUUID {
			continue
		}
		usable = append(usable, view)
	}

	result := ReconciledNodes{
		MasterNodes: make([]types.Node, 0, len(masters)),
		WorkerNodes: make([]types.Node, 0, len(workers)),
	}
	for _, node := range masters {
		if cloudInstanceIDs[node.InstanceID] {
			result.MasterNodes = append(result.MasterNodes, node)
			continue
		}
		result.Changed = true
	}
	for _, node := range workers {
		if cloudInstanceIDs[node.InstanceID] {
			result.WorkerNodes = append(result.WorkerNodes, node)
			continue
		}
		result.Changed = true
		if view, ok := matchNodeView(node, usable); ok {
			result.OrphanNodeNames = append(result.OrphanNodeNames, view.Name)
		}
	}
	return result
}

// matchNodeView finds the apiserver view of a persisted node. The identity
// label is authoritative; the internal address and the node name equal to
// the instance id match nodes that predate the labels.
func matchNodeView(node types.Node, kubeNodes []kube.NodeView) (kube.NodeView, bool) {
	for _, view := range kubeNodes {
		if id, ok := view.Labels[k3s.InstanceIDLabel]; ok && id != "" {
			if id == node.InstanceID {
				return view, true
			}
			continue
		}
		if view.InternalIP != "" && len(node.InternalIPAddress) > 0 && view.InternalIP == node.InternalIPAddress[0] {
			return view, true
		}
		if view.Name == node.InstanceID {
			return view, true
		}
	}
	return kube.NodeView{}, false
}

// countNodes renders the node counts of a reconciled state.
func countNodes(masters, workers []types.Node) (string, string) {
	return strconv.Itoa(len(masters)), strconv.Itoa(len(workers))
}

// reconcileClusterState converges the persisted cluster state with the
// apiserver and provider views observed during a describe: nodes whose
// instance is gone from the provider are pruned from the state and their
// orphaned worker Node objects are deleted, so an externally deleted node
// stops haunting the cluster view. Clusters in a transitional status are
// left alone because their node arrays are still being written by the
// lifecycle that owns them.
func (p *ProviderBase) reconcileClusterState(client kubernetes.Interface, instanceList []types.Node) {
	if common.DefaultDB == nil || p.Name == "" {
		return
	}
	state, err := common.DefaultDB.GetCluster(p.Name, p.Provider)
	if err != nil || state == nil || state.Status != common.StatusRunning {
		return
	}

	masterNodes := make([]types.Node, 0)
	if len(state.MasterNodes) > 0 {
		if err := json.Unmarshal(state.MasterNodes, &masterNodes); err != nil {
			return
		}
	}
	workerNodes := make([]types.Node, 0)
	if len(state.WorkerNodes) > 0 {
		if err := json.Unmarshal(state.WorkerNodes, &workerNodes); err != nil {
			return
		}
	}

	cloudInstanceIDs := make(map[string]bool, len(instanceList))
	for _, instance := range instanceList {
		cloudInstanceIDs[instance.InstanceID] = true
	}

	kubeNodes, err := kube.ListNodeViews(client)
	if err != nil {
		p.Logger.Errorf("[%s] failed to list nodes of cluster %s: %v", p.Provider, p.Name, err)
		return
	}

	reconciled := reconcileNodes(common.GetInstallUUID(), masterNodes, workerNodes, kubeNodes, cloudInstanceIDs)
	if !reconciled.Changed {
		return
	}

	reconcileStateMu.Lock()
	defer reconcileStateMu.Unlock()

	masterJSON, err := json.Marshal(reconciled.MasterNodes)
	if err != nil {
		return
	}
	workerJSON, err := json.Marshal(reconciled.WorkerNodes)
	if err != nil {
		return
	}
	state.MasterNodes = masterJSON
	state.WorkerNodes = workerJSON
	state.Master, state.Worker = countNodes(reconciled.MasterNodes, reconciled.WorkerNodes)
	if err := common.DefaultDB.SaveClusterState(state); err != nil {
		p.Logger.Errorf("[%s] failed to save reconciled state of cluster %s: %v", p.Provider, p.Name, err)
		return
	}

	for _, name := range reconciled.OrphanNodeNames {
		if err := client.CoreV1().Nodes().Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
			p.Logger.Warnf("[%s] failed to delete orphan node object %s: %v", p.Provider, name, err)
			continue
		}
		p.Logger.Infof("[%s] pruned worker %s: its instance no longer exists", p.Provider, name)
	}
}
