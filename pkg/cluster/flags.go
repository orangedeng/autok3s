package cluster

import (
	"fmt"

	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/orangedeng/autok3s/pkg/utils"
	"github.com/rancher/wrangler/v2/pkg/schemas"
)

// GetCreateOptions get create command flag options.
func (p *ProviderBase) GetCreateOptions() []types.Flag {
	return []types.Flag{
		{
			Name:  "cluster",
			P:     &p.Cluster,
			V:     p.Cluster,
			Usage: "Form k3s cluster using embedded etcd (requires K8s >= 1.19), see: https://docs.k3s.io/installation/ha-embedded",
		},
		{
			Name:  "manifests",
			P:     &p.Manifests,
			V:     p.Manifests,
			Usage: "A folder path for multiple manifest files(only support one directory) or a manifest file path. Auto-deploying manifests to K3s which is a manner similar to `kubectl apply`",
		},
		{
			Name:  "enable",
			P:     &p.Enable,
			V:     p.Enable,
			Usage: "Deploy add-ons (internal add-on: \"rancher\"), e.g.(--enable rancher). Other add-ons can be found by `autok3s add-ons ls`",
		},
		{
			Name:  "package-name",
			P:     &p.PackageName,
			V:     p.PackageName,
			Usage: "The airgap package name from managed package list",
		},
		{
			Name:  "package-path",
			P:     &p.PackagePath,
			V:     p.PackagePath,
			Usage: "The airgap package path. The \"package-name\" flag will be ignored if this flag is also provided",
		},
		{
			Name:  "set",
			P:     &p.Values,
			V:     p.Values,
			Usage: "set values for add-on when enabled by --enable. e.g. --enable rancher --set rancher.Version=v2.7.5 --set rancher.Hostname=aa.bb.cc",
		},
	}
}

// GetClusterOptions get cluster flag options.
func (p *ProviderBase) GetClusterOptions() []types.Flag {
	fs := []types.Flag{
		{
			Name:      "name",
			P:         &p.Name,
			V:         p.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:  "k3s-version",
			P:     &p.K3sVersion,
			V:     p.K3sVersion,
			Usage: "Used to specify the version of k3s cluster, overrides k3s-channel",
		},
		{
			Name:  "k3s-channel",
			P:     &p.K3sChannel,
			V:     p.K3sChannel,
			Usage: "Channel to use for fetching K3s download URL. Defaults to “stable”. Options include: stable, latest, testing",
		},
		{
			Name:  "k3s-install-script",
			P:     &p.InstallScript,
			V:     p.InstallScript,
			Usage: "Change the default upstream k3s install script address, see: https://docs.k3s.io/installation/configuration#options-for-installation-with-script",
		},
		{
			Name:  "k3s-install-mirror",
			P:     &p.Mirror,
			V:     p.Mirror,
			Usage: "For Chinese users, set INSTALL_K3S_MIRROR=cn to use the mirror address to accelerate k3s binary file download",
		},
		{
			Name:  "docker-arg",
			P:     &p.DockerArg,
			V:     p.DockerArg,
			Usage: "Parameter for docker script, wrapped in quotes.  e.g.(--docker-arg 'P1=\"xxxx\" P2=\"xxxx\"') for multiple parameters, P1 P2 are the parameters in docker install script",
		},
		{
			Name:  "docker-script",
			P:     &p.DockerScript,
			V:     p.DockerScript,
			Usage: fmt.Sprintf("Change the default docker install script address, default is : %s", dockerInstallScript),
		},
		{
			Name:  "master-extra-args",
			P:     &p.MasterExtraArgs,
			V:     p.MasterExtraArgs,
			Usage: "Master extra arguments for k3s installer, wrapped in quotes. e.g.(--master-extra-args '--disable metrics-server'), for more information, please see: https://docs.k3s.io/reference/server-config",
		},
		{
			Name:  "worker-extra-args",
			P:     &p.WorkerExtraArgs,
			V:     p.WorkerExtraArgs,
			Usage: "Worker extra arguments for k3s installer, wrapped in quotes. e.g.(--worker-extra-args '--node-taint key=value:NoExecute'), for more information, please see: https://docs.k3s.io/reference/agent-config",
		},
		{
			Name:  "registry",
			P:     &p.Registry,
			V:     p.Registry,
			Usage: "K3s registry file, see: https://docs.k3s.io/installation/private-registry",
		},
		{
			Name:  "system-default-registry",
			P:     &p.SystemDefaultRegistry,
			V:     p.SystemDefaultRegistry,
			Usage: "K3s private registry to be used for all system images, see: https://docs.k3s.io/reference/server-config",
		},
		{
			Name:  "datastore",
			P:     &p.DataStore,
			V:     p.DataStore,
			Usage: "K3s datastore endpoint, Specify etcd, Mysql, Postgres, or Sqlite (default) data source name, see: https://docs.k3s.io/reference/server-config#database",
		},
		{
			Name:  "datastore-cafile",
			P:     &p.DataStoreCAFile,
			V:     p.DataStoreCAFile,
			Usage: "TLS Certificate Authority (CA) file used to help secure communication with the datastore, see: https://docs.k3s.io/installation/datastore#external-datastore-configuration-parameters",
		},
		{
			Name:  "datastore-certfile",
			P:     &p.DataStoreCertFile,
			V:     p.DataStoreCertFile,
			Usage: "TLS certificate file used for client certificate based authentication to your datastore, see: https://docs.k3s.io/installation/datastore#external-datastore-configuration-parameters",
		},
		{
			Name:  "datastore-keyfile",
			P:     &p.DataStoreKeyFile,
			V:     p.DataStoreKeyFile,
			Usage: "TLS key file used for client certificate based authentication to your datastore, see: https://docs.k3s.io/installation/datastore#external-datastore-configuration-parameters",
		},
		{
			Name:  "token",
			P:     &p.Token,
			V:     p.Token,
			Usage: "K3s token, if empty will automatically generated, see: https://docs.k3s.io/reference/server-config#cluster-options",
		},
		{
			Name:  "tls-sans",
			P:     &p.TLSSans,
			V:     p.TLSSans,
			Usage: "Add additional hostnames or IPv4/IPv6 addresses as Subject Alternative Names on the server TLS cert, e.g.(--tls-sans 192.168.1.10 --tls-sans 192.168.2.10)",
		},
		{
			Name:  "master",
			P:     &p.Master,
			V:     p.Master,
			Usage: "Number of master node",
		},
		{
			Name:  "worker",
			P:     &p.Worker,
			V:     p.Worker,
			Usage: "Number of worker node",
		},
		{
			Name:  "rollback",
			P:     &p.Rollback,
			V:     p.Rollback,
			Usage: "Whether to rollback when the K3s cluster installation or join nodes failed.",
		},
		{
			Name:  "install-env",
			P:     &p.InstallEnv,
			V:     p.InstallEnv,
			Usage: "The install environment variables config for K3s with install script(only support env starts with INSTALL_), e.g. --install-env INSTALL_K3S_SKIP_SELINUX_RPM=true, see: https://docs.k3s.io/installation/configuration#configuration-with-install-script",
		},
		{
			Name:  "server-config-file",
			P:     &p.ServerConfigFile,
			V:     p.ServerConfigFile,
			Usage: "Config K3s server with configuration file which can do more complex configuration than environment variables and CLI arguments, see: https://docs.k3s.io/installation/configuration#configuration-file",
		},
		{
			Name:  "agent-config-file",
			P:     &p.AgentConfigFile,
			V:     p.AgentConfigFile,
			Usage: "Config K3s agent with configuration file which can do more complex configuration than environment variables and CLI arguments, see: https://docs.k3s.io/installation/configuration#configuration-file",
		},
	}

	fs = append(fs, p.GetSSHOptions()...)

	return fs
}

// GetSSHOptions get ssh flag options.
func (p *ProviderBase) GetSSHOptions() []types.Flag {
	return []types.Flag{
		{
			Name:  "ssh-user",
			P:     &p.SSHUser,
			V:     p.SSHUser,
			Usage: "SSH user for host",
		},
		{
			Name:  "ssh-port",
			P:     &p.SSHPort,
			V:     p.SSHPort,
			Usage: "SSH port for host",
		},
		{
			Name:  "ssh-key-path",
			P:     &p.SSHKeyPath,
			V:     p.SSHKeyPath,
			Usage: "SSH private key path",
		},
		{
			Name:  "ssh-key-passphrase",
			P:     &p.SSHKeyPassphrase,
			V:     p.SSHKeyPassphrase,
			Usage: "SSH passphrase of private key",
		},
		{
			Name:  "ssh-cert-path",
			P:     &p.SSHCertPath,
			V:     p.SSHCertPath,
			Usage: "SSH private key certificate path",
		},
		{
			Name:  "ssh-password",
			P:     &p.SSHPassword,
			V:     p.SSHPassword,
			Usage: "SSH login password",
		},
		{
			Name:  "ssh-agent-auth",
			P:     &p.SSHAgentAuth,
			V:     p.SSHAgentAuth,
			Usage: "Enable ssh agent",
		},
		{
			Name:  "ssh-key-name",
			P:     &p.SSHKeyName,
			V:     p.SSHKeyName,
			Usage: "Use the stored ssh key with name",
		},
	}
}

// GetCommonConfig get common config.
func (p *ProviderBase) GetCommonConfig(sshFunc func() *types.SSH) (map[string]schemas.Field, error) {
	ssh := sshFunc()
	sshConfig, err := utils.ConvertToFields(*ssh)
	if err != nil {
		return nil, err
	}
	metaConfig, err := utils.ConvertToFields(p.Metadata)
	if err != nil {
		return nil, err
	}
	for k, v := range sshConfig {
		metaConfig[k] = v
	}
	return metaConfig, nil
}
