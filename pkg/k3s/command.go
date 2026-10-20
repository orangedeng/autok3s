// Package k3s assembles the shell commands that install, join, restart and
// uninstall K3s on remote nodes. Everything here is pure string assembly on
// top of pkg/types: no IO, no SSH, no database, which keeps the security
// sensitive command surface easy to review and unit test.
package k3s

import (
	"fmt"
	"sort"
	"strings"

	"github.com/orangedeng/autok3s/pkg/types"
)

const (
	tlsSanArg = "--tls-san"

	// DataStoreCertificatesPath is the remote directory the datastore
	// certificates are written to before K3s starts.
	DataStoreCertificatesPath = "/etc/rancher/datastore"

	// Identity labels stamped on every node autok3s installs. They tie a
	// Node object back to the installation, the cluster and the provider
	// instance recorded in the persisted cluster state.
	InstallUUIDLabel = "autok3s.io/install-uuid"
	ClusterIDLabel   = "autok3s.io/cluster-id"
	InstanceIDLabel  = "autok3s.io/instance-id"
)

var (
	// TokenCommand reads the node token of a master node.
	TokenCommand = "cat /var/lib/rancher/k3s/server/node-token"
	// CatConfigCommand prints the kubeconfig of a master node.
	CatConfigCommand = "cat /etc/rancher/k3s/k3s.yaml"
	// DockerInstallCommand installs docker when it is not present yet.
	DockerInstallCommand = "if ! type docker; then curl -sSL %s | %s sh - %s; fi"
	// MasterUninstallCommand uninstalls K3s from a master node.
	MasterUninstallCommand = "[ -x /usr/local/bin/k3s-uninstall.sh ] && sh /usr/local/bin/k3s-uninstall.sh || true"
	// WorkerUninstallCommand uninstalls the K3s agent from a worker node.
	WorkerUninstallCommand = "[ -x /usr/local/bin/k3s-agent-uninstall.sh ] && sh /usr/local/bin/k3s-agent-uninstall.sh || true"
	// RestartCommand restarts the K3s server service.
	RestartCommand = `if [ -n "$(command -v systemctl)" ]; then systemctl restart k3s; elif [ -n "$(command -v service)" ]; then service k3s restart; fi`
	// AgentRestartCommand restarts the K3s agent service.
	AgentRestartCommand = `if [ -n "$(command -v systemctl)" ]; then systemctl restart k3s-agent; elif [ -n "$(command -v service)" ]; then service k3s-agent restart; fi`
)

// InstallCommand renders the shell command that installs K3s on one node.
// The first master initializes the cluster; every other node joins it.
// The installUUID stamps the identity labels of the installing autok3s
// instance; an empty value skips the labels entirely.
func InstallCommand(isFirstMaster bool, fixedIP string, installUUID string, cluster *types.Cluster, node types.Node, extraArgs []string) string {
	var commandPrefix, commandSuffix string
	envVar := map[string]string{}
	// prepare for install env
	for key, value := range cluster.InstallEnv {
		if key == "INSTALL_K3S_EXEC" {
			// add install exec to extraArgs
			extraArgs = append(extraArgs, value)
			continue
		}
		envVar[key] = value
	}
	// airgap install
	if cluster.PackageName != "" || cluster.PackagePath != "" {
		commandSuffix = "install.sh"
		envVar["INSTALL_K3S_SKIP_DOWNLOAD"] = "true"
	} else {
		commandPrefix = fmt.Sprintf("curl -sLS %s |", cluster.InstallScript)
		commandSuffix = "sh -"
		if cluster.K3sVersion != "" {
			envVar["INSTALL_K3S_VERSION"] = cluster.K3sVersion
		} else if cluster.K3sChannel != "" {
			envVar["INSTALL_K3S_CHANNEL"] = cluster.K3sChannel
		}
	}

	if cluster.Mirror != "" {
		kv := strings.SplitN(cluster.Mirror, "=", 2)
		if len(kv) < 2 {
			kv = append(kv, "")
		}
		envVar[kv[0]] = kv[1]
	}
	envVar["K3S_TOKEN"] = cluster.Token

	if !node.Master {
		envVar["K3S_URL"] = fmt.Sprintf("https://%s:6443", fixedIP)
	}

	runArgs := getRunArgs(isFirstMaster, fixedIP, cluster, node)
	runArgs = append(runArgs, NodeLabelArgs(installUUID, cluster, node)...)
	runArgs = append(runArgs, extraArgs...)
	envVar["INSTALL_K3S_EXEC"] = strings.Join(runArgs, " ")

	sortedEnvVars := []string{}
	for k, v := range envVar {
		sortedEnvVars = append(sortedEnvVars, fmt.Sprintf("%s='%s'", k, v))
	}
	sort.Strings(sortedEnvVars)
	return strings.TrimSpace(fmt.Sprintf("%s %s %s", commandPrefix, strings.Join(sortedEnvVars, " "), commandSuffix))
}

// NodeLabelArgs returns the identity label arguments for one node, in a
// stable order. Labels with an empty value are skipped, so a cluster
// without an install uuid still carries the cluster and instance ids.
func NodeLabelArgs(installUUID string, cluster *types.Cluster, node types.Node) []string {
	labels := map[string]string{
		InstallUUIDLabel: installUUID,
		ClusterIDLabel:   cluster.ContextName,
		InstanceIDLabel:  node.InstanceID,
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	args := make([]string, 0, len(keys))
	for _, key := range keys {
		if labels[key] == "" {
			continue
		}
		args = append(args, "--node-label="+key+"="+labels[key])
	}
	return args
}

// TLSSans returns the sorted, deduplicated list of TLS SANs of the cluster:
// the explicitly configured ones plus the addresses of all master nodes.
func TLSSans(cluster *types.Cluster) []string {
	dedump := map[string]bool{}
	for _, san := range cluster.TLSSans {
		dedump[san] = true
	}
	for _, master := range cluster.MasterNodes {
		if addr := FirstAddress(master.PublicIPAddress); addr != "" {
			dedump[addr] = true
		}
		if addr := FirstAddress(master.InternalIPAddress); addr != "" {
			dedump[addr] = true
		}
	}
	var rtn []string
	for k := range dedump {
		rtn = append(rtn, k)
	}
	sort.Strings(rtn)
	return rtn
}

// FirstAddress returns the first non empty address of the list.
func FirstAddress(addresses []string) string {
	for _, addr := range addresses {
		if addr != "" {
			return addr
		}
	}
	return ""
}

func getRunArgs(isFirstMaster bool, fixedIP string, cluster *types.Cluster, node types.Node) []string {
	runArgs := []string{}

	// should also handle join command as well
	if node.Master {
		isCluster := cluster.Cluster
		if cluster.DataStore != "" {
			isCluster = false
			runArgs = append(runArgs, "--datastore-endpoint="+cluster.DataStore)
		}

		if cluster.DataStoreCAFileContent != "" || cluster.DataStoreCAFile != "" {
			runArgs = append(runArgs, fmt.Sprintf("--datastore-cafile=%s/ds-ca.pem", DataStoreCertificatesPath))
		}

		if cluster.DataStoreCertFileContent != "" || cluster.DataStoreCertFile != "" {
			runArgs = append(runArgs, fmt.Sprintf("--datastore-certfile=%s/ds-cert.pem", DataStoreCertificatesPath))
		}

		if cluster.DataStoreKeyFileContent != "" || cluster.DataStoreKeyFile != "" {
			runArgs = append(runArgs, fmt.Sprintf("--datastore-keyfile=%s/ds-key.pem", DataStoreCertificatesPath))
		}

		if isCluster && isFirstMaster {
			runArgs = append(runArgs, "--cluster-init")
		}
		if isCluster && !isFirstMaster {
			runArgs = append(runArgs, "--server="+fmt.Sprintf("https://%s:6443", fixedIP))
		}

		if cluster.ClusterCidr != "" {
			runArgs = append(runArgs, "--cluster-cidr="+cluster.ClusterCidr)
		}

		for _, san := range TLSSans(cluster) {
			runArgs = append(runArgs, tlsSanArg+"="+san)
		}

		internalIPAddress := FirstAddress(node.InternalIPAddress)
		if internalIPAddress != "" {
			runArgs = append(runArgs, "--advertise-address="+internalIPAddress)
		}

		if cluster.SystemDefaultRegistry != "" {
			runArgs = append(runArgs, "--system-default-registry="+cluster.SystemDefaultRegistry)
		}
	}

	if externalAddr := FirstAddress(node.PublicIPAddress); externalAddr != "" {
		runArgs = append(runArgs, "--node-external-ip="+externalAddr)
	}

	if cluster.Network != "" {
		runArgs = append(runArgs, "--flannel-backend="+cluster.Network)
	}

	sort.Strings(runArgs)
	if node.Master {
		// ensure server arg is the first one
		runArgs = append([]string{"server"}, runArgs...)
	}
	return runArgs
}
