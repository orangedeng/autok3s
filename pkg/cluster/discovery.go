package cluster

import (
	"time"

	"path/filepath"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/providers"
	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/sirupsen/logrus"
)

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
