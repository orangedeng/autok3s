package cluster

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/orangedeng/autok3s/pkg/cluster/kube"
	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/providers"
	putil "github.com/orangedeng/autok3s/pkg/providers/utils"
	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/sirupsen/logrus"
)

// GetClusterStatus get cluster status.
func (p *ProviderBase) GetClusterStatus(kubeCfg string, c *types.ClusterInfo, describeFunc func() ([]types.Node, error)) *types.ClusterInfo {
	p.Logger = logrus.StandardLogger()

	c.Master = p.Master
	c.Worker = p.Worker
	if p.Cluster {
		c.IsHAMode = true
		c.DataStoreType = "Embedded DB(etcd)"
	} else if p.DataStore != "" {
		c.IsHAMode = true
		dataStoreArray := strings.Split(p.DataStore, "://")
		if dataStoreArray[0] == "http" {
			c.DataStoreType = "External DB(etcd)"
		} else {
			c.DataStoreType = fmt.Sprintf("External DB(%s)", dataStoreArray[0])
		}
	}

	if kubeCfg == "" {
		return c
	}

	client, err := kube.Client(p.ContextName, kubeCfg)
	if err != nil {
		p.Logger.Errorf("[%s] failed to generate kube client for cluster %s: %v", p.Provider, p.ContextName, err)
		c.Status = types.ClusterStatusUnknown
		c.Version = types.ClusterStatusUnknown
		return c
	}

	c.Status = kube.Status(client)
	if c.Status == types.ClusterStatusRunning {
		c.Version = kube.Version(client)
	} else {
		c.Version = types.ClusterStatusUnknown
	}

	if describeFunc != nil {
		instanceList, err := describeFunc()
		if err != nil {
			p.Logger.Errorf("%v", err)
			c.Master = "0"
			c.Worker = "0"
			return c
		}
		masterCount := 0
		workerCount := 0
		for _, ins := range instanceList {
			if ins.Master {
				masterCount++
				continue
			}
			workerCount++
		}
		c.Master = strconv.Itoa(masterCount)
		c.Worker = strconv.Itoa(workerCount)
	}

	return c
}

// listProbeTimeout bounds the remote probing of a single cluster when listing,
// so an unreachable cluster degrades to its persisted state instead of
// stalling the whole list behind the caller's request timeout.
var listProbeTimeout = 20 * time.Second

// ListClusters list clusters.
func ListClusters(providerName string) ([]*types.ClusterInfo, error) {
	stateList, err := common.DefaultDB.ListCluster(providerName)
	if err != nil {
		return nil, err
	}
	kubeCfg := filepath.Join(common.CfgPath, common.KubeCfgFile)
	clusterList := make([]*types.ClusterInfo, 0)
	for _, state := range stateList {
		clusterList = append(clusterList, describeClusterState(state, kubeCfg))
	}
	return clusterList, nil
}

// describeClusterState builds the list entry of a single cluster. Every remote
// probe is bounded by listProbeTimeout; a cluster that cannot be reached in
// time falls back to its persisted state instead of blocking the remaining
// clusters.
func describeClusterState(state *common.ClusterState, kubeCfg string) *types.ClusterInfo {
	provider, err := providers.GetProvider(state.Provider)
	if err != nil {
		logrus.Errorf("failed to get provider %v: %v", state.Provider, err)
		return stateClusterInfo(state, common.StatusUnknown)
	}
	provider.SetMetadata(&state.Metadata)
	_ = provider.SetOptions(state.Options)
	contextName := provider.GenerateClusterName()
	if state.Status != common.StatusRunning {
		info := provider.GetCluster("")
		info.Status = state.Status
		info.Master = state.Master
		info.Worker = state.Worker
		return info
	}
	type existCheck struct {
		exist bool
		err   error
	}
	check, ok := withTimeout(func() existCheck {
		isExist, _, existErr := provider.IsClusterExist()
		return existCheck{exist: isExist, err: existErr}
	}, listProbeTimeout)
	if !ok {
		logrus.Warnf("timeout checking cluster %s (provider %s) exist, return persisted state", state.Name, state.Provider)
		return stateClusterInfo(state, common.StatusUnknown)
	}
	if check.err != nil {
		logrus.Errorf("failed to check provider %s cluster %s exist, got error: %v ", state.Provider, state.Name, check.err)
		return stateClusterInfo(state, common.StatusUnknown)
	}
	if !check.exist {
		logrus.Warnf("cluster %s (provider %s) is not exist, will remove from config", state.Name, state.Provider)
		// remove kube config if cluster not exist
		if err := common.FileManager.ClearCfgByContext(contextName); err != nil {
			logrus.Errorf("failed to remove unexist cluster %s from kube config", state.Name)
		}
		// update status to missing
		state.Status = common.StatusMissing
		if err := common.DefaultDB.SaveClusterState(state); err != nil {
			logrus.Errorf("failed to update cluster %s state to missing", state.Name)
		}
		info := provider.GetCluster("")
		info.Status = state.Status
		info.Master = state.Master
		info.Worker = state.Worker
		return info
	}
	info, ok := withTimeout(func() *types.ClusterInfo {
		return provider.GetCluster(kubeCfg)
	}, listProbeTimeout)
	if !ok {
		logrus.Warnf("timeout probing cluster %s (provider %s), return persisted state", state.Name, state.Provider)
		return stateClusterInfo(state, common.StatusUnknown)
	}
	return info
}

// stateClusterInfo builds a cluster list entry from the persisted state only.
func stateClusterInfo(state *common.ClusterState, status string) *types.ClusterInfo {
	return &types.ClusterInfo{
		ID:       state.ContextName,
		Name:     state.Name,
		Provider: state.Provider,
		Status:   status,
		Master:   state.Master,
		Worker:   state.Worker,
	}
}

// withTimeout runs fn and reports whether it finished within timeout. A timed
// out fn keeps running in the background; callers must fall back to their own
// data instead of waiting for its result.
func withTimeout[T any](fn func() T, timeout time.Duration) (T, bool) {
	result := make(chan T, 1)
	go func() {
		result <- fn()
	}()
	select {
	case value := <-result:
		return value, true
	case <-time.After(timeout):
		var zero T
		return zero, false
	}
}

func (p *ProviderBase) syncExistNodes() {
	p.M.Range(func(key, value interface{}) bool {
		v := value.(types.Node)
		nodes := p.Status.WorkerNodes
		if v.Master {
			nodes = p.Status.MasterNodes
		}
		index, b := putil.IsExistedNodes(nodes, v.InstanceID)
		if !b {
			nodes = append(nodes, v)
		} else {
			node := nodes[index]
			node.InstanceStatus = v.InstanceStatus
			nodes[index] = node
		}
		if v.Master {
			p.Status.MasterNodes = nodes
		} else {
			p.Status.WorkerNodes = nodes
		}
		return true
	})
}

// Describe describe cluster info.
func (p *ProviderBase) Describe(kubeCfg string, c *types.ClusterInfo, describeInstance func() ([]types.Node, error)) *types.ClusterInfo {
	c.Master = p.Master
	c.Worker = p.Worker

	if p.Cluster {
		c.IsHAMode = true
		c.DataStoreType = "Embedded DB(etcd)"
	} else if p.DataStore != "" {
		c.IsHAMode = true
		dataStoreArray := strings.Split(p.DataStore, "://")
		if dataStoreArray[0] == "http" {
			c.DataStoreType = "External DB(etcd)"
		} else {
			c.DataStoreType = fmt.Sprintf("External DB(%s)", dataStoreArray[0])
		}
	}

	if kubeCfg == "" {
		c.Status = common.StatusMissing
		return c
	}
	p.Logger = logrus.StandardLogger()
	client, err := kube.Client(p.ContextName, kubeCfg)
	if err != nil {
		p.Logger.Errorf("[%s] failed to generate kube client for cluster %s: %v", p.Provider, p.Name, err)
		c.Status = types.ClusterStatusUnknown
		c.Version = types.ClusterStatusUnknown
		return c
	}
	c.Status = kube.Status(client)
	if describeInstance != nil {
		instanceList, err := describeInstance()
		if err != nil {
			p.Logger.Errorf("[%s] failed to get instance for cluster %s: %v", p.Provider, p.Name, err)
			c.Master = "0"
			c.Worker = "0"
			return c
		}
		instanceNodes := make([]types.ClusterNode, 0)
		masterCount := 0
		workerCount := 0
		for _, instance := range instanceList {
			instanceNodes = append(instanceNodes, types.ClusterNode{
				InstanceID:              instance.InstanceID,
				InstanceStatus:          instance.InstanceStatus,
				InternalIP:              instance.InternalIPAddress,
				ExternalIP:              instance.PublicIPAddress,
				Standalone:              instance.Standalone,
				Status:                  types.ClusterStatusUnknown,
				ContainerRuntimeVersion: types.ClusterStatusUnknown,
				Version:                 types.ClusterStatusUnknown,
			})
			if instance.Master {
				masterCount++
				continue
			}
			workerCount++
		}
		c.Master = strconv.Itoa(masterCount)
		c.Worker = strconv.Itoa(workerCount)
		c.Nodes = instanceNodes
		if c.Status == types.ClusterStatusRunning {
			c.Version = kube.Version(client)
			nodes, err := kube.DescribeNodes(client, instanceNodes)
			if err != nil {
				p.Logger.Errorf("[%s] failed to list nodes of cluster %s: %v", p.Provider, p.Name, err)
				return c
			}
			c.Nodes = nodes
		} else {
			c.Version = types.ClusterStatusUnknown
		}
	}

	return c
}
