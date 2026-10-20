package base

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"encoding/base64"
	"path/filepath"

	"github.com/orangedeng/autok3s/pkg/airgap"
	"github.com/orangedeng/autok3s/pkg/cluster/kube"
	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/hosts/dialer"
	"github.com/orangedeng/autok3s/pkg/k3s"
	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/orangedeng/autok3s/pkg/utils"
	"github.com/rancher/wharfie/pkg/registries"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	registryPath = "/etc/rancher/k3s"
)

// InitK3sCluster initial K3S cluster.
func (p *ProviderBase) InitK3sCluster(cluster *types.Cluster, deployCCM func() []string,
	generateMasterExtraArgs, generateWorkerExtraArgs func(*types.Cluster, types.Node) string) error {
	p.Logger.Infof("[%s] executing init k3s cluster logic...", p.Provider)

	pkg, err := airgap.PreparePackage(cluster)
	if err != nil {
		return err
	}
	// package's name is empty, it means that it is a temporary dir and it needs to be remove after.
	if pkg != nil && pkg.Name == "" {
		defer os.RemoveAll(pkg.FilePath)
	}

	if cluster.Token == "" {
		token, err := utils.RandomToken(16)
		if err != nil {
			return err
		}
		cluster.Token = token
	}

	if len(cluster.MasterNodes) <= 0 || len(cluster.MasterNodes[0].InternalIPAddress) <= 0 {
		return errors.New("[cluster] master node internal ip address can not be empty")
	}

	publicIP := cluster.IP
	if cluster.IP == "" {
		cluster.IP = cluster.MasterNodes[0].InternalIPAddress[0]
		publicIP = cluster.MasterNodes[0].PublicIPAddress[0]
	}

	// initialize the first master node and worker node to validate the K3s configuration.
	var firstControl, firstWorker types.Node
	controlNodes := []types.Node{}
	for i, control := range cluster.MasterNodes {
		if i == 0 {
			firstControl = control
			continue
		}
		controlNodes = append(controlNodes, control)
	}
	workerNodes := []types.Node{}
	if len(p.WorkerNodes) > 0 {
		for i, w := range cluster.WorkerNodes {
			if i == 0 {
				firstWorker = w
				continue
			}
			workerNodes = append(workerNodes, w)
		}
	}

	err = p.validateClusterConfig(cluster, publicIP, pkg, firstControl, firstWorker, deployCCM,
		generateMasterExtraArgs, generateWorkerExtraArgs)
	if err != nil {
		return err
	}

	for i, master := range controlNodes {
		p.Logger.Infof("[%s] join k3s control-%d...", p.Provider, i+1)
		if err := p.initControlNode(cluster, publicIP, pkg, master, false, generateMasterExtraArgs); err != nil {
			return err
		}
		p.Logger.Infof("[%s] successfully created k3s master-%d", p.Provider, i+1)
	}

	// batch join worker nodes
	var wg sync.WaitGroup
	var l sync.RWMutex
	for i, worker := range workerNodes {
		wg.Add(1)
		go func(i int, worker types.Node) {
			defer wg.Done()
			p.Logger.Infof("[%s] creating k3s worker-%d...", p.Provider, i+1)
			if err := p.initWorkerNode(cluster, publicIP, pkg, worker, generateWorkerExtraArgs); err != nil {
				l.Lock()
				p.ErrM[worker.InstanceID] = err.Error()
				l.Unlock()
			}
			p.Logger.Infof("[%s] successfully created k3s worker-%d", p.Provider, i+1)
		}(i, worker)
	}
	wg.Wait()

	// get k3s cluster config.
	cfg, err := p.executeWithRetry(3, &cluster.MasterNodes[0], k3s.CatConfigCommand)
	if err != nil {
		return err
	}

	// merge current cluster to kube config.
	if err := kube.SaveConfig(cfg, publicIP, cluster.ContextName); err != nil {
		return err
	}
	_ = os.Setenv(clientcmd.RecommendedConfigPathEnvVar, filepath.Join(common.CfgPath, common.KubeCfgFile))
	cluster.Status.Status = common.StatusRunning

	// write current cluster to state file.
	if err := common.DefaultDB.SaveCluster(cluster); err != nil {
		return err
	}

	p.Logger.Infof("[%s] deploying additional manifests", p.Provider)

	// deploy additional UI manifests.
	enabledPlugins := map[string]bool{}

	// deploy plugin
	if cluster.Enable != nil {
		for _, comp := range cluster.Enable {
			enabledPlugins[comp] = true
		}
	}

	p.Logger.Infof("[%s] successfully deployed additional manifests", p.Provider)
	p.Logger.Infof("[%s] successfully executed init k3s cluster logic", p.Provider)
	return nil
}

// DeployExtraManifest deploy extra K3S manifest.
func (p *ProviderBase) DeployExtraManifest(cluster *types.Cluster, cmds []string) error {
	if _, err := p.Execute(&cluster.MasterNodes[0], []string{fmt.Sprintf("mkdir -p %s", common.K3sManifestsDir)}...); err != nil {
		return err
	}
	if _, err := p.Execute(&cluster.MasterNodes[0], cmds...); err != nil {
		return err
	}
	return nil
}

func (p *ProviderBase) initNode(isFirstMaster bool, fixedIP string, cluster *types.Cluster, node types.Node, extraArgs string, pkg *common.Package) error {
	if strings.Contains(extraArgs, "--docker") {
		dockerCmd := fmt.Sprintf(k3s.DockerInstallCommand, cluster.DockerScript, cluster.DockerArg, cluster.DockerMirror)
		p.Logger.Infof("[cluster] install docker command %s", dockerCmd)
		if _, err := p.Execute(&node, dockerCmd); err != nil {
			return err
		}
	}

	if cluster.Registry != "" || cluster.RegistryContent != "" {
		if err := p.handleRegistry(&node, cluster); err != nil {
			return err
		}
	}

	if cluster.DataStoreCAFile != "" || cluster.DataStoreCAFileContent != "" ||
		cluster.DataStoreCertFileContent != "" || cluster.DataStoreCertFile != "" ||
		cluster.DataStoreKeyFileContent != "" || cluster.DataStoreKeyFile != "" {
		if err := p.handleDataStoreCertificate(&node, cluster); err != nil {
			return err
		}
	}

	if pkg != nil {
		if err := p.scpFiles(cluster.Name, pkg, &node, extraArgs); err != nil {
			return err
		}
	}

	// handle configuration file
	if err := p.handleConfiguration(&node, cluster); err != nil {
		return err
	}

	cmd := k3s.InstallCommand(isFirstMaster, fixedIP, common.GetInstallUUID(), cluster, node, []string{extraArgs})
	nodeRole := "master"
	if !node.Master {
		nodeRole = "worker"
	}

	p.Logger.Infof("[cluster] k3s %s command: %s", nodeRole, cmd)

	if _, err := p.Execute(&node, cmd); err != nil {
		return err
	}

	return nil
}

func (p *ProviderBase) handleRegistry(n *types.Node, c *types.Cluster) (err error) {
	if c.Registry == "" && c.RegistryContent == "" {
		return nil
	}
	var cmd []string

	registry, err := utils.VerifyRegistryFileContent(p.Registry, p.RegistryContent)
	if err != nil {
		return err
	}

	tls, err := registryTLSMap(registry)
	if err != nil {
		return err
	}

	if len(tls) > 0 {
		cmd, err = saveRegistryTLS(registry, tls)
		if err != nil {
			return err
		}
	} else {
		cmd = []string{fmt.Sprintf("mkdir -p %s", registryPath)}
	}

	registryContent, err := utils.RegistryToString(registry)
	if err != nil {
		return err
	}

	cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"/etc/rancher/k3s/registries.yaml\"",
		base64.StdEncoding.EncodeToString([]byte(registryContent))))
	_, err = p.Execute(n, cmd...)
	return err
}

func (p *ProviderBase) handleConfiguration(n *types.Node, c *types.Cluster) (err error) {
	if c.ServerConfigFileContent == "" && c.ServerConfigFile == "" && c.AgentConfigFileContent == "" && c.AgentConfigFile == "" {
		return nil
	}
	var configFileContent []byte
	if n.Master {
		if c.ServerConfigFileContent != "" {
			configFileContent, err = base64.StdEncoding.DecodeString(c.ServerConfigFileContent)
			if err != nil {
				return err
			}
		}
		if c.ServerConfigFile != "" && c.ServerConfigFileContent == "" {
			configFileContent, err = os.ReadFile(c.ServerConfigFile)
			if err != nil {
				return err
			}
		}
	} else {
		if c.AgentConfigFileContent != "" {
			configFileContent, err = base64.StdEncoding.DecodeString(c.AgentConfigFileContent)
			if err != nil {
				return err
			}
		}
		if c.AgentConfigFile != "" && c.AgentConfigFileContent == "" {
			configFileContent, err = os.ReadFile(c.AgentConfigFile)
			if err != nil {
				return err
			}
		}
	}
	if len(configFileContent) == 0 {
		return errors.New("configuration file is empty")
	}
	cmd := []string{"mkdir -p /etc/rancher/k3s"}
	cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"/etc/rancher/k3s/config.yaml\"",
		base64.StdEncoding.EncodeToString(configFileContent)))
	_, err = p.Execute(n, cmd...)
	return err
}

func registryTLSMap(registry *registries.Registry) (m map[string]map[string][]byte, err error) {
	m = make(map[string]map[string][]byte)
	if registry == nil {
		err = fmt.Errorf("registry is nil")
		return
	}

	for r, c := range registry.Configs {
		if _, ok := m[r]; !ok {
			m[r] = map[string][]byte{}
		}
		if c.TLS == nil {
			continue
		}
		if c.TLS.CertFile != "" {
			b, err := os.ReadFile(c.TLS.CertFile)
			if err != nil {
				return m, err
			}
			m[r]["cert"] = b
		}
		if c.TLS.KeyFile != "" {
			b, err := os.ReadFile(c.TLS.KeyFile)
			if err != nil {
				return m, err
			}
			m[r]["key"] = b
		}
		if c.TLS.CAFile != "" {
			b, err := os.ReadFile(c.TLS.CAFile)
			if err != nil {
				return m, err
			}
			m[r]["ca"] = b
		}
	}

	return
}

func saveRegistryTLS(registry *registries.Registry, m map[string]map[string][]byte) ([]string, error) {
	cmd := make([]string, 0)
	for r, c := range m {
		if r != "" {
			if _, ok := registry.Configs[r]; !ok {
				return cmd, fmt.Errorf("registry map is not match the struct: %s", r)
			}

			// i.e /etc/rancher/k3s/mycustomreg:5000/.
			path := fmt.Sprintf("/etc/rancher/k3s/%s", r)
			cmd = append(cmd, fmt.Sprintf("mkdir -p %s", path))

			for f, b := range c {
				// i.e /etc/rancher/k3s/mycustomreg:5000/{ca,key,cert}.
				file := fmt.Sprintf("%s/%s", path, f)
				cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s\"", base64.StdEncoding.EncodeToString(b), file))
				cmd = append(cmd, fmt.Sprintf("chmod 755 %s", file))

				switch f {
				case "cert":
					registry.Configs[r].TLS.CertFile = file
				case "key":
					registry.Configs[r].TLS.KeyFile = file
				case "ca":
					registry.Configs[r].TLS.CAFile = file
				}
			}
		}
	}

	return cmd, nil
}

func (p *ProviderBase) scpFiles(clusterName string, pkg *common.Package, node *types.Node, extraArgs string) error {
	dialer, err := dialer.NewSSHDialer(node, true, p.Logger)
	if err != nil {
		return err
	}
	defer dialer.Close()
	return airgap.ScpFiles(p.Logger, clusterName, pkg, dialer, extraArgs)
}

func (p *ProviderBase) handleDataStoreCertificate(n *types.Node, c *types.Cluster) error {
	cmd := make([]string, 0)
	cmd = append(cmd, fmt.Sprintf("mkdir -p %s", k3s.DataStoreCertificatesPath))
	if c.DataStoreCAFile != "" {
		caFile, err := os.ReadFile(c.DataStoreCAFile)
		if err != nil {
			return err
		}
		cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s/ds-ca.pem\"",
			base64.StdEncoding.EncodeToString(caFile), k3s.DataStoreCertificatesPath))
	}
	if c.DataStoreCertFile != "" {
		certFile, err := os.ReadFile(c.DataStoreCertFile)
		if err != nil {
			return err
		}
		cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s/ds-cert.pem\"",
			base64.StdEncoding.EncodeToString(certFile), k3s.DataStoreCertificatesPath))
	}
	if c.DataStoreKeyFile != "" {
		keyFile, err := os.ReadFile(c.DataStoreKeyFile)
		if err != nil {
			return err
		}
		cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s/ds-key.pem\"",
			base64.StdEncoding.EncodeToString(keyFile), k3s.DataStoreCertificatesPath))
	}
	if c.DataStoreCAFileContent != "" {
		cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s/ds-ca.pem\"",
			base64.StdEncoding.EncodeToString([]byte(p.DataStoreCAFileContent)), k3s.DataStoreCertificatesPath))
	}
	if c.DataStoreKeyFileContent != "" {
		cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s/ds-key.pem\"",
			base64.StdEncoding.EncodeToString([]byte(p.DataStoreKeyFileContent)), k3s.DataStoreCertificatesPath))
	}
	if c.DataStoreCertFileContent != "" {
		cmd = append(cmd, fmt.Sprintf("echo \"%s\" | base64 -d | tee \"%s/ds-cert.pem\"",
			base64.StdEncoding.EncodeToString([]byte(p.DataStoreCertFileContent)), k3s.DataStoreCertificatesPath))
	}
	_, err := p.Execute(n, cmd...)
	return err
}

func (p *ProviderBase) validateClusterConfig(cluster *types.Cluster, publicIP string, pkg *common.Package, firstControl, firstWorker types.Node, deployCCM func() []string,
	generateMasterExtraArgs, generateWorkerExtraArgs func(*types.Cluster, types.Node) string) error {
	p.Logger.Infof("[%s] initialize control node...", p.Provider)
	if err := p.initControlNode(cluster, publicIP, pkg, firstControl, true, generateMasterExtraArgs); err != nil {
		return err
	}
	if deployCCM != nil {
		extraManifests := deployCCM()
		err := p.DeployExtraManifest(cluster, extraManifests)
		if err != nil {
			return err
		}
	}
	p.Logger.Infof("[%s] successfully initialize the first control node", p.Provider)

	if len(firstWorker.PublicIPAddress) <= 0 && firstWorker.InstanceID == "" {
		// skip with empty worker node
		return nil
	}
	p.Logger.Infof("[%s] initialize worker node...", p.Provider)
	if err := p.initWorkerNode(cluster, publicIP, pkg, firstWorker, generateWorkerExtraArgs); err != nil {
		return err
	}
	p.Logger.Infof("[%s] successfully initialize the first worker node", p.Provider)

	return nil
}

func (p *ProviderBase) initControlNode(cluster *types.Cluster, publicIP string, pkg *common.Package, controlNode types.Node, isFirst bool,
	generateMasterExtraArgs func(*types.Cluster, types.Node) string) error {
	masterExtraArgs := cluster.MasterExtraArgs
	providerExtraArgs := generateMasterExtraArgs(cluster, controlNode)
	if providerExtraArgs != "" {
		masterExtraArgs += providerExtraArgs
	}

	return p.initNode(isFirst, publicIP, cluster, controlNode, masterExtraArgs, pkg)
}

func (p *ProviderBase) initWorkerNode(cluster *types.Cluster, publicIP string, pkg *common.Package, workerNode types.Node,
	generateWorkerExtraArgs func(*types.Cluster, types.Node) string) error {
	workerExtraArgs := cluster.WorkerExtraArgs
	providerExtraArgs := generateWorkerExtraArgs(cluster, workerNode)
	if providerExtraArgs != "" {
		workerExtraArgs += providerExtraArgs
	}
	return p.initNode(false, publicIP, cluster, workerNode, workerExtraArgs, pkg)
}
