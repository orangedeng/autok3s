package base

import (
	"errors"
	"fmt"
	"os"

	"github.com/orangedeng/autok3s/pkg/airgap"
	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/k3s"
	"github.com/orangedeng/autok3s/pkg/providers"
	"github.com/orangedeng/autok3s/pkg/types"
)

func (p *ProviderBase) UpgradeK3sCluster(clusterName, installScript, channel, version, packageName, packagePath string) error {
	if p.Provider == "k3d" {
		return errors.New("the upgrade cluster for K3d provider is not supported yet")
	}
	state, err := common.DefaultDB.GetCluster(clusterName, p.Provider)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("cluster %s is not exist", clusterName)
	}
	p.Name = clusterName
	p.ContextName = state.ContextName
	logFile, err := common.GetLogFile(state.ContextName)
	if err != nil {
		return err
	}
	p.Logger = common.NewLogger(logFile)
	p.Logger.Infof("[%s] begin to upgrade cluster %s...", p.Provider, clusterName)
	state.Status = common.StatusUpgrading
	// save cluster.
	err = common.DefaultDB.SaveClusterState(state)
	if err != nil {
		return err
	}

	defer func() {
		// update cluster status
		state.Status = common.StatusRunning
		_ = common.DefaultDB.SaveClusterState(state)
		// remove upgrade state file and save running state.
		_ = logFile.Close()
		if p.Callbacks != nil {
			if process, ok := p.Callbacks[state.ContextName]; ok && process.Event == "update" {
				logEvent := &common.LogEvent{
					Name:        process.Event,
					ContextType: "cluster",
					ContextName: state.ContextName,
				}
				process.Fn(logEvent)
			}
		}
	}()

	c := common.ConvertToCluster(state, true)

	if installScript != "" {
		c.InstallScript = installScript
		state.InstallScript = installScript
	}
	if channel != "" {
		c.K3sChannel = channel
		state.K3sChannel = channel
	}
	if version != "" {
		c.K3sVersion = version
		state.K3sVersion = version
	}

	// if online install specified, clean up offline options and ignore package name/path input
	if installScript != "" || channel != "" || version != "" {
		c.PackageName = ""
		state.PackageName = ""
		c.PackagePath = ""
		state.PackagePath = ""
	} else {
		if packageName != "" {
			c.PackageName = packageName
			state.PackageName = packageName
		}
		if packagePath != "" {
			c.PackagePath = packagePath
			state.PackagePath = packagePath
		}
	}

	return p.Upgrade(&c)
}

func (p *ProviderBase) Upgrade(cluster *types.Cluster) error {
	p.Logger.Infof("[%s] executing upgrade k3s cluster logic...", p.Provider)
	if len(cluster.MasterNodes) <= 0 || len(cluster.MasterNodes[0].InternalIPAddress) <= 0 {
		return errors.New("[cluster] master node internal ip address can not be empty")
	}

	pkg, err := airgap.PreparePackage(cluster)
	if err != nil {
		return err
	}
	// package's name is empty, it means that it is a temporary dir and it needs to be remove after.
	if pkg != nil && pkg.Name == "" {
		defer os.RemoveAll(pkg.FilePath)
	}

	// Upgrade is entered through the providers.Provider interface, whose
	// fixed signature cannot carry the extra args generators, so the
	// concrete provider is looked up by name here.
	provider, err := providers.GetProvider(p.Provider)
	if err != nil {
		return err
	}
	masterExtraArgs := cluster.MasterExtraArgs
	workerExtraArgs := cluster.WorkerExtraArgs

	publicIP := cluster.IP
	if cluster.IP == "" {
		cluster.IP = cluster.MasterNodes[0].InternalIPAddress[0]
		publicIP = cluster.MasterNodes[0].PublicIPAddress[0]
	}

	// upgrade server nodes
	for i, node := range cluster.MasterNodes {
		extraArgs := masterExtraArgs
		providerExtraArgs := provider.GenerateMasterExtraArgs(cluster, node)
		if providerExtraArgs != "" {
			extraArgs += providerExtraArgs
		}

		var cmd string

		if pkg != nil {
			if err := p.scpFiles(cluster.Name, pkg, &node, extraArgs); err != nil {
				return err
			}
			cmd = k3s.RestartCommand
		} else {
			cmd = k3s.InstallCommand(i == 0, publicIP, common.GetInstallUUID(), cluster, node, []string{extraArgs})
		}

		p.Logger.Infof("[cluster] upgrading k3s master %d command: %s", i+1, cmd)

		if _, err := p.Execute(&node, cmd); err != nil {
			return err
		}
	}

	// upgrade worker nodes
	for i, node := range cluster.WorkerNodes {
		extraArgs := workerExtraArgs
		providerExtraArgs := provider.GenerateWorkerExtraArgs(cluster, node)
		if providerExtraArgs != "" {
			extraArgs += providerExtraArgs
		}

		var cmd string
		if pkg != nil {
			if err := p.scpFiles(cluster.Name, pkg, &node, extraArgs); err != nil {
				return err
			}
			cmd = k3s.AgentRestartCommand
		} else {
			cmd = k3s.InstallCommand(false, publicIP, common.GetInstallUUID(), cluster, node, []string{extraArgs})
		}

		p.Logger.Infof("[cluster] upgrading k3s worker %d command: %s", i+1, cmd)

		if _, err := p.Execute(&node, cmd); err != nil {
			return err
		}
	}

	return nil
}
