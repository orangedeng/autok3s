package base

import (
	"os"
	"strconv"

	"path/filepath"

	"github.com/orangedeng/autok3s/pkg/cluster/kube"
	"github.com/orangedeng/autok3s/pkg/common"
	pkgsshkey "github.com/orangedeng/autok3s/pkg/sshkey"
	"github.com/orangedeng/autok3s/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
)

// InitCluster init K3S cluster.
func (p *ProviderBase) InitCluster(options interface{}, deployCCM func() []string,
	cloudInstanceFunc func(ssh *types.SSH) (*types.Cluster, error), customInstallK3s func() (string, string, error), rollbackInstance func(ids []string) error,
	generateMasterExtraArgs, generateWorkerExtraArgs func(*types.Cluster, types.Node) string) (er error) {
	logFile, err := common.GetLogFile(p.ContextName)
	if err != nil {
		return err
	}

	c := &types.Cluster{
		Metadata: p.Metadata,
		Options:  options,
		SSH:      p.SSH,
		Status:   p.Status,
	}
	defer func() {
		if er != nil || len(p.ErrM) > 0 {
			// save failed status.
			state, err := common.DefaultDB.GetCluster(p.Name, p.Provider)
			if err == nil {
				var c types.Cluster
				if state == nil {
					c = types.Cluster{
						Metadata: p.Metadata,
						Options:  options,
						SSH:      p.SSH,
						Status:   types.Status{},
					}
				} else {
					c = common.ConvertToCluster(state, true)
				}
				if er != nil {
					p.Logger.Errorf("%v", er)
					c.Status.Status = common.StatusFailed
				}
				if len(p.ErrM) > 0 && len(c.WorkerNodes) > 0 {
					// remove failed node from cluster state
					workers := []types.Node{}
					for _, n := range c.WorkerNodes {
						if _, ok := p.ErrM[n.InstanceID]; !ok {
							workers = append(workers, n)
						}
					}
					c.WorkerNodes = workers
					c.Worker = strconv.Itoa(len(workers))
				}
				_ = common.DefaultDB.SaveCluster(&c)
			}
			_ = p.RollbackCluster(rollbackInstance)
		}
		if er == nil && len(p.Status.MasterNodes) > 0 {
			p.Logger.Info(common.UsageInfoTitle)
			p.Logger.Infof(common.UsageContext, p.ContextName)
			p.Logger.Info(common.UsagePods)
		}
		_ = logFile.Close()
		if p.Callbacks != nil {
			if process, ok := p.Callbacks[p.ContextName]; ok && process.Event == "create" {
				logEvent := &common.LogEvent{
					Name:        process.Event,
					ContextType: "cluster",
					ContextName: p.ContextName,
				}
				process.Fn(logEvent)
			}
		}
	}()
	p.Logger = common.NewLogger(logFile)
	p.Logger.Infof("[%s] begin to create cluster %s...", p.Provider, p.Name)
	c.Status.Status = common.StatusCreating
	// save cluster.
	if err = common.DefaultDB.SaveCluster(c); err != nil {
		return err
	}
	// store ssh key
	if newSSH, err := pkgsshkey.StoreClusterSSHKeys(p.ContextName, &c.SSH); err != nil {
		return err
	} else if newSSH != nil {
		p.Logger.Infof("[%s] cluster's ssh keys saved", p.Name)
		c.SSH = *newSSH
		// update cluster with stored key
		if err = common.DefaultDB.SaveCluster(c); err != nil {
			return err
		}
		p.SSH = *newSSH
	}

	c, err = cloudInstanceFunc(&c.SSH)
	if err != nil {
		return err
	}
	p.syncExistNodes()
	c.Status = p.Status

	if customInstallK3s == nil {
		// use install scripts to initialize K3s cluster.
		if err = p.InitK3sCluster(c, deployCCM, generateMasterExtraArgs, generateWorkerExtraArgs); err != nil {
			return err
		}
	} else {
		// some providers do not need to initialize the K3s cluster with scripts,
		// so we need to fill in the missing key information.
		cfg, ip, err := customInstallK3s()
		if err != nil {
			return err
		}
		// save current cluster's kubeConfig.
		if err := kube.SaveConfig(cfg, ip, c.ContextName); err != nil {
			return err
		}
		_ = os.Setenv(clientcmd.RecommendedConfigPathEnvVar, filepath.Join(common.CfgPath, common.KubeCfgFile))
		// change & save current cluster's status to database.
		c.Status.Status = common.StatusRunning
		if err = common.DefaultDB.SaveCluster(c); err != nil {
			return err
		}
	}

	cmds := []string{}

	if p.Manifests != "" {
		deployCmd, err := p.GetCustomManifests()
		if err != nil {
			p.Logger.Errorf("[%s] failed to get custom manifests by manifest %s: %v", p.Provider, p.Manifests, err)
		}
		cmds = append(cmds, deployCmd...)
	}

	if p.Enable != nil {
		for _, plugin := range p.Enable {
			cmd, err := p.addonInstallation(plugin)
			if err != nil {
				continue
			}
			cmds = append(cmds, cmd)
		}
	}

	// deploy custom manifests.
	if len(cmds) > 0 {
		if err = p.DeployExtraManifest(c, cmds); err != nil {
			return err
		}
		p.Logger.Infof("[%s] successfully deployed custom manifests", p.Provider)
	}

	return
}
