package base

import (
	"fmt"
	"os"
	"strings"
	"time"

	errors "github.com/pkg/errors"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/hosts/dialer"
	"github.com/orangedeng/autok3s/pkg/k3s"
	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/orangedeng/autok3s/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Connect ssh & connect to the K3S node.
func (p *ProviderBase) Connect(ip string, ssh *types.SSH, c *types.Cluster, getStatus func() ([]types.Node, error),
	isRunning func(status string) bool, customConnect func(id string, cluster *types.Cluster) error) error {
	p.Logger = logrus.StandardLogger()
	p.Logger.Infof("[%s] executing ssh logic...", p.Provider)

	if getStatus == nil {
		return fmt.Errorf("failed to get status for provider %s", p.Provider)
	}

	status, err := getStatus()
	if err != nil {
		return err
	}

	ids := make(map[string]string, len(status))

	if ip == "" {
		// generate the node name and determine the current state of the node.
		for _, s := range status {
			var info string
			if len(s.PublicIPAddress) > 0 {
				info = s.PublicIPAddress[0]
			} else {
				info = s.InstanceID
			}
			if s.Master {
				info = fmt.Sprintf("%s (master)", info)
			} else {
				info = fmt.Sprintf("%s (worker)", info)
			}
			if !isRunning(s.InstanceStatus) {
				info = fmt.Sprintf("%s - Unhealthy(%s)", info, s.InstanceStatus)
			}
			ids[s.InstanceID] = info
		}
	}

	if ip == "" {
		ip = strings.Split(utils.AskForSelectItem(fmt.Sprintf("[%s] choose ssh node to connect", p.Provider), ids), " (")[0]
	}

	if ip == "" {
		return fmt.Errorf("[%s] choose incorrect ssh node", p.Provider)
	}

	if customConnect == nil {
		// ssh to the typically node.
		if err := SSHK3sNode(ip, c, ssh); err != nil {
			return err
		}
	} else {
		// some providers do not typically use IP connections,
		// so we need to use a custom connect function.
		if err := customConnect(ip, c); err != nil {
			return err
		}
	}

	p.Logger.Infof("[%s] successfully executed ssh logic", p.Provider)
	return nil
}

// SSHK3sNode ssh to K3S node.
func SSHK3sNode(ip string, cluster *types.Cluster, ssh *types.SSH) error {
	var node types.Node

	for _, n := range cluster.Status.MasterNodes {
		if n.PublicIPAddress[0] == ip || n.InstanceID == ip {
			node = n
			break
		}
	}

	for _, n := range cluster.Status.WorkerNodes {
		if n.PublicIPAddress[0] == ip || n.InstanceID == ip {
			node = n
			break
		}
	}

	if ssh.SSHUser != "" {
		node.SSH.SSHUser = ssh.SSHUser
	}
	if ssh.SSHPort != "" {
		node.SSH.SSHPort = ssh.SSHPort
	}
	if ssh.SSHPassword != "" {
		node.SSH.SSHPassword = ssh.SSHPassword
	}
	if ssh.SSHKeyPath != "" {
		node.SSH.SSHKeyPath = ssh.SSHKeyPath
	}
	if ssh.SSHCert != "" {
		node.SSH.SSHCert = ssh.SSHCert
	}
	if ssh.SSHCertPath != "" {
		node.SSH.SSHCertPath = ssh.SSHCertPath
	}
	if ssh.SSHKeyPassphrase != "" {
		node.SSH.SSHKeyPassphrase = ssh.SSHKeyPassphrase
	}
	if ssh.SSHAgentAuth {
		node.SSH.SSHAgentAuth = ssh.SSHAgentAuth
	}
	if node.PublicIPAddress == nil {
		node.PublicIPAddress = []string{ip}
	}

	if node.SSH.SSHPort == "" {
		node.SSH.SSHPort = "22"
	}

	// preCheck ssh config
	if node.SSH.SSHUser == "" || (node.SSH.SSHPassword == "" && node.SSH.SSHKeyPath == "") {
		return fmt.Errorf("couldn't ssh to chosen node with current ssh config: --ssh-user %s --ssh-port %s --ssh-password %s --ssh-key-path %s", node.SSH.SSHUser, node.SSH.SSHPort, node.SSH.SSHPassword, node.SSH.SSHKeyPath)
	}

	return terminal(&node)
}

// UninstallK3sNodes uninstall K3S on the given nodes.
func (p *ProviderBase) UninstallK3sNodes(nodes []types.Node) (warnMsg []string) {
	for _, node := range nodes {
		if node.Master {
			_, e := p.Execute(&node, k3s.MasterUninstallCommand)
			if e != nil {
				warnMsg = append(warnMsg, fmt.Sprintf("failed to uninstall k3s on master node %s: %s", node.InstanceID, e.Error()))
			}
		} else {
			_, e := p.Execute(&node, k3s.WorkerUninstallCommand)
			if e != nil {
				warnMsg = append(warnMsg, fmt.Sprintf("failed to uninstall k3s on worker node %s: %s", node.InstanceID, e.Error()))
			}
		}
	}

	return
}

func (p *ProviderBase) Execute(n *types.Node, cmds ...string) (string, error) {
	if len(cmds) <= 0 {
		return "", nil
	}

	dialer, err := dialer.NewSSHDialer(n, true, p.Logger)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = dialer.Close()
	}()
	output, err := dialer.ExecuteCommands(cmds...)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, output)
	}

	return output, nil
}

func (p *ProviderBase) executeWithRetry(count int, n *types.Node, cmds ...string) (string, error) {
	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   1,
		Steps:    count,
	}
	var rtn string
	var lastError error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		rtn, lastError = p.Execute(n, cmds...)
		if lastError != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", errors.Wrapf(lastError, "failed to execute cmd with max retry count %d", count)
	}
	return rtn, nil
}

func terminal(n *types.Node) error {
	dialer, err := dialer.NewSSHDialer(n, true, common.NewLogger(nil))
	if err != nil {
		return err
	}

	defer func() {
		_ = dialer.Close()
	}()
	shell, err := dialer.OpenShell()
	if err != nil {
		return err
	}

	shell.SetIO(os.Stdout, os.Stderr, os.Stdin)

	return shell.Terminal()
}
