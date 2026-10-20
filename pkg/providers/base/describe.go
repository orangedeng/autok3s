// Package base implements the lifecycle orchestration shared by every
// provider: creating, joining, deleting and upgrading clusters, plus the
// flag and config plumbing providers embed. It depends on pkg/k3s for
// command assembly, pkg/cluster/kube for API server access and pkg/common
// for state, never on the provider registry itself.
package base

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/orangedeng/autok3s/pkg/cluster/kube"
	"github.com/orangedeng/autok3s/pkg/common"
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
			// both views are already in hand, so converging the persisted
			// state adds no provider or apiserver request.
			p.reconcileClusterState(client, instanceList)
		} else {
			c.Version = types.ClusterStatusUnknown
		}
	}

	return c
}
