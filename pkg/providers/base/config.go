package base

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/orangedeng/autok3s/pkg/utils"
)

// MergeConfig merge cluster config.
func (p *ProviderBase) MergeConfig() ([]byte, error) {
	state, err := common.DefaultDB.GetCluster(p.Name, p.Provider)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil
	}
	p.overwriteMetadata(state)
	p.Status = types.Status{
		Status:     state.Status,
		Standalone: state.Standalone,
	}
	masterNodes := make([]types.Node, 0)
	err = json.Unmarshal(state.MasterNodes, &masterNodes)
	if err != nil {
		return nil, err
	}
	workerNodes := make([]types.Node, 0)
	err = json.Unmarshal(state.WorkerNodes, &workerNodes)
	if err != nil {
		return nil, err
	}
	p.Status.MasterNodes = masterNodes
	p.Status.WorkerNodes = workerNodes

	source := reflect.ValueOf(&p.SSH).Elem()
	target := reflect.ValueOf(&state.SSH).Elem()
	utils.MergeConfig(source, target)

	return state.Options, nil
}

func (p *ProviderBase) overwriteMetadata(matched *common.ClusterState) {
	// doesn't need to be overwrite.
	p.Token = matched.Token
	p.IP = matched.IP
	p.UI = matched.UI
	p.ClusterCidr = matched.ClusterCidr
	p.DataStore = matched.DataStore
	p.Mirror = matched.Mirror
	p.DockerMirror = matched.DockerMirror
	p.InstallScript = matched.InstallScript
	p.Network = matched.Network
	p.Cluster = matched.Cluster
	p.Rollback = matched.Rollback
	// needed to be overwrite.
	if p.K3sChannel == "" {
		p.K3sChannel = matched.K3sChannel
	}
	if p.K3sVersion == "" {
		p.K3sVersion = matched.K3sVersion
	}
	if p.InstallScript == "" {
		p.InstallScript = matched.InstallScript
	}
	if p.DockerArg == "" {
		p.DockerArg = matched.DockerArg
	}
	if p.DockerScript == "" {
		p.DockerScript = matched.DockerScript
	}
	if p.Registry == "" {
		p.Registry = matched.Registry
	}
	if p.SystemDefaultRegistry == "" {
		p.SystemDefaultRegistry = matched.SystemDefaultRegistry
	}
	if p.MasterExtraArgs == "" {
		p.MasterExtraArgs = matched.MasterExtraArgs
	}
	if p.WorkerExtraArgs == "" {
		p.WorkerExtraArgs = matched.WorkerExtraArgs
	}
}

func (p *ProviderBase) CheckCreateArgs(checkClusterExist func() (bool, []string, error)) error {
	if p.Provider != "native" {
		masterNum, err := strconv.Atoi(p.Master)
		if masterNum < 1 || err != nil {
			return fmt.Errorf("[%s] calling preflight error: `--master` number must >= 1",
				p.Provider)
		}
		if p.Provider != "k3d" && masterNum > 1 && !p.Cluster && p.DataStore == "" {
			return fmt.Errorf("[%s] calling preflight error: need to set `--cluster` or `--datastore` when `--master` number > 1",
				p.Provider)
		}
		if p.Provider != "k3d" && strings.Contains(p.MasterExtraArgs, "--datastore-endpoint") && p.DataStore != "" {
			return fmt.Errorf("[%s] calling preflight error: `--masterExtraArgs='--datastore-endpoint'` is duplicated with `--datastore`",
				p.Provider)
		}
		if _, err = strconv.Atoi(p.Worker); err != nil {
			return fmt.Errorf("[%s] calling preflight error: `--worker` must be number",
				p.Provider)
		}
	}

	// check name exist.
	state, err := common.DefaultDB.GetCluster(p.Name, p.Provider)
	if err != nil {
		return err
	}

	if state != nil && state.Status != common.StatusFailed {
		return fmt.Errorf("[%s] cluster %s is already exist", p.Provider, p.Name)
	}

	exist, _, err := checkClusterExist()
	if err != nil {
		return err
	}

	if exist {
		return fmt.Errorf("[%s] calling preflight error: cluster `%s` is already exist",
			p.Provider, p.Name)
	}

	// check file exists.
	if p.SSHKeyPath != "" && !utils.IsFileExists(p.SSHKeyPath) {
		return fmt.Errorf("[%s] failed to check --ssh-key-path %s", p.Provider, p.SSHKeyPath)
	}
	if p.SSHCertPath != "" && !utils.IsFileExists(p.SSHCertPath) {
		return fmt.Errorf("[%s] failed to check --ssh-cert-path %s", p.Provider, p.SSHCertPath)
	}

	if p.Registry != "" && !utils.IsFileExists(p.Registry) {
		return fmt.Errorf("[%s] failed to check --registry %s", p.Provider, p.Registry)
	}

	if p.DataStoreCAFile != "" && !utils.IsFileExists(p.DataStoreCAFile) {
		return fmt.Errorf("[%s] failed to check --datastore-cafile %s", p.Provider, p.DataStoreCAFile)
	}

	if p.DataStoreCertFile != "" && !utils.IsFileExists(p.DataStoreCertFile) {
		return fmt.Errorf("[%s] failed to check --datastore-certfile %s", p.Provider, p.DataStoreCertFile)
	}

	if p.DataStoreKeyFile != "" && !utils.IsFileExists(p.DataStoreKeyFile) {
		return fmt.Errorf("[%s] failed to check --datastore-keyfile %s", p.Provider, p.DataStoreKeyFile)
	}

	if p.InstallEnv != nil {
		unsupportEnv := []string{}
		for key, value := range p.InstallEnv {
			if !strings.HasPrefix(key, "INSTALL_") {
				unsupportEnv = append(unsupportEnv, fmt.Sprintf("%v=%v", key, value))
			}
		}
		if len(unsupportEnv) > 0 {
			return fmt.Errorf("[%s] Invalid install environment %v. Only support INSTALL_* environments. For K3S_* variables please use config file args", p.Provider, unsupportEnv)
		}
	}

	return nil
}

func (p *ProviderBase) CheckJoinArgs(checkClusterExist func() (bool, []string, error)) error {
	// check cluster exist.
	exist, _, err := checkClusterExist()

	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("[%s] calling preflight error: cluster name `%s` do not exist",
			p.Provider, p.ContextName)
	}

	if strings.Contains(p.MasterExtraArgs, "--datastore-endpoint") && p.DataStore != "" {
		return fmt.Errorf("[%s] calling preflight error: `--masterExtraArgs='--datastore-endpoint'` is duplicated with `--datastore`",
			p.Provider)
	}

	if p.Provider != "native" {
		masterNum, err := strconv.Atoi(p.Master)
		if err != nil {
			return fmt.Errorf("[%s] calling preflight error: `--master` must be number",
				p.Provider)
		}

		if masterNum > 0 && p.DataStore == "" && !p.Cluster {
			return fmt.Errorf("[%s] calling preflight error: can't join master nodes to single node cluster", p.Provider)
		}

		workerNum, err := strconv.Atoi(p.Worker)
		if err != nil {
			return fmt.Errorf("[%s] calling preflight error: `--worker` must be number",
				p.Provider)
		}
		if masterNum < 1 && workerNum < 1 {
			return fmt.Errorf("[%s] calling preflight error: `--master` or `--worker` number must >= 1", p.Provider)
		}
	}

	if p.InstallEnv != nil {
		unsupportEnv := []string{}
		for key, value := range p.InstallEnv {
			if !strings.HasPrefix(key, "INSTALL_") {
				unsupportEnv = append(unsupportEnv, fmt.Sprintf("%v=%v", key, value))
			}
		}
		if len(unsupportEnv) > 0 {
			return fmt.Errorf("[%s] Invalid install environment %v. Only support INSTALL_* environments. For K3S_* variables please use config file args", p.Provider, unsupportEnv)
		}
	}

	return nil
}

// SetMetadata set metadata.
func (p *ProviderBase) SetMetadata(config *types.Metadata) {
	sourceMeta := reflect.ValueOf(&p.Metadata).Elem()
	targetMeta := reflect.ValueOf(config).Elem()
	utils.MergeConfig(sourceMeta, targetMeta)
}

// SetClusterConfig set cluster config.
func (p *ProviderBase) SetClusterConfig(config []byte) (*types.Cluster, error) {
	c := types.Cluster{}
	err := json.Unmarshal(config, &c)
	if err != nil {
		return nil, err
	}
	sourceMeta := reflect.ValueOf(&p.Metadata).Elem()
	targetMeta := reflect.ValueOf(&c.Metadata).Elem()
	utils.MergeConfig(sourceMeta, targetMeta)

	sourceSSH := reflect.ValueOf(&p.SSH).Elem()
	targetSSH := reflect.ValueOf(&c.SSH).Elem()
	utils.MergeConfig(sourceSSH, targetSSH)
	return &c, nil
}

// SaveCredential save credential to database.
func (p *ProviderBase) SaveCredential(secrets map[string]string) error {
	cs, err := common.DefaultDB.GetCredentialByProvider(p.Provider)
	if err != nil {
		return err
	}
	if len(cs) == 0 {
		s, err := json.Marshal(secrets)
		if err != nil {
			return err
		}
		err = common.DefaultDB.CreateCredential(&common.Credential{
			Provider: p.Provider,
			Secrets:  s,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *ProviderBase) ValidateRequireSSHPrivateKey() error {
	errStr := "ssh key is require but none of --ssh-key-path or --ssh-key-name is provided"
	if !common.IsCLI {
		errStr = "ssh key is require but none of --ssh-key-path, --ssh-key-name or --ssh-key is provided"
	}
	if p.SSHKey == "" && p.SSHKeyPath == "" && p.SSHKeyName == "" {
		return errors.New(errStr)
	}
	return nil
}
