package cluster

import (
	"fmt"
	"os"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/orangedeng/autok3s/pkg/utils"
	"github.com/sirupsen/logrus"
)

// DeleteCluster delete cluster.
func (p *ProviderBase) DeleteCluster(force bool, delete func(f bool) (string, error)) error {
	isConfirmed := true

	if !force {
		isConfirmed = utils.AskForConfirmation(fmt.Sprintf("[%s] are you sure to delete cluster %s", p.Provider, p.Name), false)
	}
	if isConfirmed {
		logFile, err := common.GetLogFile(p.ContextName)
		if err != nil {
			return err
		}
		defer func() {
			_ = logFile.Close()
			// remove log file.
			_ = os.RemoveAll(common.GetClusterContextPath(p.ContextName))
		}()
		state, err := common.DefaultDB.GetCluster(p.Name, p.Provider)
		if err != nil && !force {
			return fmt.Errorf("[%s] failed to get cluster %s, got error %v", p.Provider, p.Name, err)
		}
		p.Logger = common.NewLogger(logFile)
		p.Logger.Infof("[%s] begin to delete cluster %v...", p.Provider, p.Name)
		if state != nil {
			state.Status = common.StatusRemoving
			err = common.DefaultDB.SaveClusterState(state)
			if err != nil && !force {
				return fmt.Errorf("[%s] failed to update cluster %s status to removing, got error %v", p.Provider, p.Name, err)
			}
		}

		contextName, err := delete(force)
		if err != nil {
			return err
		}
		err = common.FileManager.ClearCfgByContext(contextName)
		if err != nil && !force {
			return fmt.Errorf("[%s] merge kubeconfig error, msg: %v", p.Provider, err)
		}
		err = common.DefaultDB.DeleteCluster(p.Name, p.Provider)
		if err != nil && !force {
			return fmt.Errorf("[%s] failed to delete cluster state, msg: %v", p.Provider, err)
		}

		p.Logger.Infof("[%s] successfully deleted cluster %s", p.Provider, p.Name)
	}
	return nil
}

// RollbackCluster rollback when error occur.
func (p *ProviderBase) RollbackCluster(rollbackInstance func(ids []string) error) error {
	if !p.Rollback {
		p.Logger.Warnf("[%s] skip executing rollback logic. This is only used for troubleshooting, the instances and cluster is out of control by AutoK3s. Please check K3s error log and uninstalled manually if needed.", p.Provider)
		return nil
	}
	p.Logger.Infof("[%s] executing rollback logic...", p.Provider)
	if rollbackInstance != nil {
		ids := make([]string, 0)
		// support for partial rollback
		if len(p.ErrM) > 0 {
			p.Logger.Warnf("[%s] The following instances need to rollback in some of reasons...", p.Provider)
			for key, value := range p.ErrM {
				p.Logger.Warnf("[%s] The instance %s is failed to join to the K3s cluster with error: %v", p.Provider, key, value)
				ids = append(ids, key)
			}
		} else {
			p.M.Range(func(key, value interface{}) bool {
				v := value.(types.Node)
				if v.RollBack {
					ids = append(ids, key.(string))
				}
				return true
			})
		}

		p.Logger.Infof("[%s] instances %s will be rollback", p.Provider, ids)

		// remove instance.
		if err := rollbackInstance(ids); err != nil {
			return err
		}

		state, err := common.DefaultDB.GetCluster(p.Name, p.Provider)
		if err != nil {
			return err
		}
		if state == nil || state.Status != common.StatusRunning {
			// remove context.
			if err := common.FileManager.ClearCfgByContext(p.ContextName); err != nil {
				logrus.Errorf("failed to remove cluster context %s from kube config", p.ContextName)
			}
		}

		p.Logger.Infof("[%s] successfully executed rollback logic", p.Provider)
	}

	return nil
}
