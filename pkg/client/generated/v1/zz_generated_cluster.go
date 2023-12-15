package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	ClusterType                          = "cluster"
	ClusterFieldCluster                  = "cluster"
	ClusterFieldClusterCidr              = "cluster-cidr"
	ClusterFieldContextName              = "context-name"
	ClusterFieldDataStore                = "datastore"
	ClusterFieldDataStoreCAFile          = "datastore-cafile"
	ClusterFieldDataStoreCAFileContent   = "datastore-cafile-content"
	ClusterFieldDataStoreCertFile        = "datastore-certfile"
	ClusterFieldDataStoreCertFileContent = "datastore-certfile-content"
	ClusterFieldDataStoreKeyFile         = "datastore-keyfile"
	ClusterFieldDataStoreKeyFileContent  = "datastore-keyfile-content"
	ClusterFieldDataStoreType            = "datastore-type"
	ClusterFieldDockerArg                = "docker-arg"
	ClusterFieldDockerMirror             = "dockerMirror"
	ClusterFieldDockerScript             = "docker-script"
	ClusterFieldEnable                   = "enable"
	ClusterFieldIP                       = "ip"
	ClusterFieldInstallScript            = "k3s-install-script"
	ClusterFieldIsHAMode                 = "is-ha-mode"
	ClusterFieldK3sChannel               = "k3s-channel"
	ClusterFieldK3sVersion               = "k3s-version"
	ClusterFieldManifests                = "manifests"
	ClusterFieldMaster                   = "master"
	ClusterFieldMasterExtraArgs          = "master-extra-args"
	ClusterFieldMasterNodes              = "master-nodes"
	ClusterFieldMirror                   = "k3s-install-mirror"
	ClusterFieldName                     = "name"
	ClusterFieldNetwork                  = "network"
	ClusterFieldOptions                  = "options"
	ClusterFieldPackageName              = "package-name"
	ClusterFieldPackagePath              = "package-path"
	ClusterFieldProvider                 = "provider"
	ClusterFieldRegistry                 = "registry"
	ClusterFieldRegistryContent          = "registry-content"
	ClusterFieldRollback                 = "rollback"
	ClusterFieldSSHAgentAuth             = "ssh-agent-auth"
	ClusterFieldSSHCert                  = "ssh-cert"
	ClusterFieldSSHCertPath              = "ssh-cert-path"
	ClusterFieldSSHKey                   = "ssh-key"
	ClusterFieldSSHKeyName               = "ssh-key-name"
	ClusterFieldSSHKeyPassphrase         = "ssh-key-passphrase"
	ClusterFieldSSHKeyPath               = "ssh-key-path"
	ClusterFieldSSHPassword              = "ssh-password"
	ClusterFieldSSHPort                  = "ssh-port"
	ClusterFieldSSHUser                  = "ssh-user"
	ClusterFieldStandalone               = "standalone"
	ClusterFieldStatus                   = "status"
	ClusterFieldSystemDefaultRegistry    = "system-default-registry"
	ClusterFieldTLSSans                  = "tls-sans"
	ClusterFieldToken                    = "token"
	ClusterFieldUI                       = "ui"
	ClusterFieldValues                   = "values"
	ClusterFieldWorker                   = "worker"
	ClusterFieldWorkerExtraArgs          = "worker-extra-args"
	ClusterFieldWorkerNodes              = "worker-nodes"
)

type Cluster struct {
	types.Resource
	Cluster                  bool              `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	ClusterCidr              string            `json:"cluster-cidr,omitempty" yaml:"cluster-cidr,omitempty"`
	ContextName              string            `json:"context-name,omitempty" yaml:"context-name,omitempty"`
	DataStore                string            `json:"datastore,omitempty" yaml:"datastore,omitempty"`
	DataStoreCAFile          string            `json:"datastore-cafile,omitempty" yaml:"datastore-cafile,omitempty"`
	DataStoreCAFileContent   string            `json:"datastore-cafile-content,omitempty" yaml:"datastore-cafile-content,omitempty"`
	DataStoreCertFile        string            `json:"datastore-certfile,omitempty" yaml:"datastore-certfile,omitempty"`
	DataStoreCertFileContent string            `json:"datastore-certfile-content,omitempty" yaml:"datastore-certfile-content,omitempty"`
	DataStoreKeyFile         string            `json:"datastore-keyfile,omitempty" yaml:"datastore-keyfile,omitempty"`
	DataStoreKeyFileContent  string            `json:"datastore-keyfile-content,omitempty" yaml:"datastore-keyfile-content,omitempty"`
	DataStoreType            string            `json:"datastore-type,omitempty" yaml:"datastore-type,omitempty"`
	DockerArg                string            `json:"docker-arg,omitempty" yaml:"docker-arg,omitempty"`
	DockerMirror             string            `json:"dockerMirror,omitempty" yaml:"dockerMirror,omitempty"`
	DockerScript             string            `json:"docker-script,omitempty" yaml:"docker-script,omitempty"`
	Enable                   []string          `json:"enable,omitempty" yaml:"enable,omitempty"`
	IP                       string            `json:"ip,omitempty" yaml:"ip,omitempty"`
	InstallScript            string            `json:"k3s-install-script,omitempty" yaml:"k3s-install-script,omitempty"`
	IsHAMode                 bool              `json:"is-ha-mode,omitempty" yaml:"is-ha-mode,omitempty"`
	K3sChannel               string            `json:"k3s-channel,omitempty" yaml:"k3s-channel,omitempty"`
	K3sVersion               string            `json:"k3s-version,omitempty" yaml:"k3s-version,omitempty"`
	Manifests                string            `json:"manifests,omitempty" yaml:"manifests,omitempty"`
	Master                   string            `json:"master,omitempty" yaml:"master,omitempty"`
	MasterExtraArgs          string            `json:"master-extra-args,omitempty" yaml:"master-extra-args,omitempty"`
	MasterNodes              []Node            `json:"master-nodes,omitempty" yaml:"master-nodes,omitempty"`
	Mirror                   string            `json:"k3s-install-mirror,omitempty" yaml:"k3s-install-mirror,omitempty"`
	Name                     string            `json:"name,omitempty" yaml:"name,omitempty"`
	Network                  string            `json:"network,omitempty" yaml:"network,omitempty"`
	Options                  interface{}       `json:"options,omitempty" yaml:"options,omitempty"`
	PackageName              string            `json:"package-name,omitempty" yaml:"package-name,omitempty"`
	PackagePath              string            `json:"package-path,omitempty" yaml:"package-path,omitempty"`
	Provider                 string            `json:"provider,omitempty" yaml:"provider,omitempty"`
	Registry                 string            `json:"registry,omitempty" yaml:"registry,omitempty"`
	RegistryContent          string            `json:"registry-content,omitempty" yaml:"registry-content,omitempty"`
	Rollback                 bool              `json:"rollback,omitempty" yaml:"rollback,omitempty"`
	SSHAgentAuth             bool              `json:"ssh-agent-auth,omitempty" yaml:"ssh-agent-auth,omitempty"`
	SSHCert                  string            `json:"ssh-cert,omitempty" yaml:"ssh-cert,omitempty"`
	SSHCertPath              string            `json:"ssh-cert-path,omitempty" yaml:"ssh-cert-path,omitempty"`
	SSHKey                   string            `json:"ssh-key,omitempty" yaml:"ssh-key,omitempty"`
	SSHKeyName               string            `json:"ssh-key-name,omitempty" yaml:"ssh-key-name,omitempty"`
	SSHKeyPassphrase         string            `json:"ssh-key-passphrase,omitempty" yaml:"ssh-key-passphrase,omitempty"`
	SSHKeyPath               string            `json:"ssh-key-path,omitempty" yaml:"ssh-key-path,omitempty"`
	SSHPassword              string            `json:"ssh-password,omitempty" yaml:"ssh-password,omitempty"`
	SSHPort                  string            `json:"ssh-port,omitempty" yaml:"ssh-port,omitempty"`
	SSHUser                  string            `json:"ssh-user,omitempty" yaml:"ssh-user,omitempty"`
	Standalone               bool              `json:"standalone,omitempty" yaml:"standalone,omitempty"`
	Status                   string            `json:"status,omitempty" yaml:"status,omitempty"`
	SystemDefaultRegistry    string            `json:"system-default-registry,omitempty" yaml:"system-default-registry,omitempty"`
	TLSSans                  []string          `json:"tls-sans,omitempty" yaml:"tls-sans,omitempty"`
	Token                    string            `json:"token,omitempty" yaml:"token,omitempty"`
	UI                       bool              `json:"ui,omitempty" yaml:"ui,omitempty"`
	Values                   map[string]string `json:"values,omitempty" yaml:"values,omitempty"`
	Worker                   string            `json:"worker,omitempty" yaml:"worker,omitempty"`
	WorkerExtraArgs          string            `json:"worker-extra-args,omitempty" yaml:"worker-extra-args,omitempty"`
	WorkerNodes              []Node            `json:"worker-nodes,omitempty" yaml:"worker-nodes,omitempty"`
}

type ClusterCollection struct {
	types.Collection
	Data   []Cluster `json:"data,omitempty"`
	client *ClusterClient
}

type ClusterClient struct {
	apiClient *Client
}

type ClusterOperations interface {
	List(opts *clientbase.ListOpts) (*ClusterCollection, error)
	ListAll(opts *clientbase.ListOpts) (*ClusterCollection, error)
	Create(opts *Cluster) (*Cluster, error)
	Update(existing *Cluster, updates interface{}) (*Cluster, error)
	Replace(existing *Cluster) (*Cluster, error)
	ByID(id string) (*Cluster, error)
	Delete(container *Cluster) error

	ActionDisableExplorer(resource *Cluster) error

	ActionDownloadKubeconfig(resource *Cluster) (*KubeconfigOutput, error)

	ActionEnableExplorer(resource *Cluster) (*EnableExplorerOutput, error)

	ActionJoin(resource *Cluster, input *Cluster) error

	ActionUpgrade(resource *Cluster, input *UpgradeInput) error
}

func newClusterClient(apiClient *Client) *ClusterClient {
	return &ClusterClient{
		apiClient: apiClient,
	}
}

func (c *ClusterClient) Create(container *Cluster) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoCreate(ClusterType, container, resp)
	return resp, err
}

func (c *ClusterClient) Update(existing *Cluster, updates interface{}) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoUpdate(ClusterType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterClient) Replace(obj *Cluster) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoReplace(ClusterType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ClusterClient) List(opts *clientbase.ListOpts) (*ClusterCollection, error) {
	resp := &ClusterCollection{}
	err := c.apiClient.Ops.DoList(ClusterType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *ClusterClient) ListAll(opts *clientbase.ListOpts) (*ClusterCollection, error) {
	resp := &ClusterCollection{}
	resp, err := c.List(opts)
	if err != nil {
		return resp, err
	}
	data := resp.Data
	for next, err := resp.Next(); next != nil && err == nil; next, err = next.Next() {
		data = append(data, next.Data...)
		resp = next
		resp.Data = data
	}
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (cc *ClusterCollection) Next() (*ClusterCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ClusterCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ClusterClient) ByID(id string) (*Cluster, error) {
	resp := &Cluster{}
	err := c.apiClient.Ops.DoByID(ClusterType, id, resp)
	return resp, err
}

func (c *ClusterClient) Delete(container *Cluster) error {
	return c.apiClient.Ops.DoResourceDelete(ClusterType, &container.Resource)
}

func (c *ClusterClient) ActionDisableExplorer(resource *Cluster) error {
	err := c.apiClient.Ops.DoAction(ClusterType, "disableExplorer", &resource.Resource, nil, nil)
	return err
}

func (c *ClusterClient) ActionDownloadKubeconfig(resource *Cluster) (*KubeconfigOutput, error) {
	resp := &KubeconfigOutput{}
	err := c.apiClient.Ops.DoAction(ClusterType, "downloadKubeconfig", &resource.Resource, nil, resp)
	return resp, err
}

func (c *ClusterClient) ActionEnableExplorer(resource *Cluster) (*EnableExplorerOutput, error) {
	resp := &EnableExplorerOutput{}
	err := c.apiClient.Ops.DoAction(ClusterType, "enableExplorer", &resource.Resource, nil, resp)
	return resp, err
}

func (c *ClusterClient) ActionJoin(resource *Cluster, input *Cluster) error {
	err := c.apiClient.Ops.DoAction(ClusterType, "join", &resource.Resource, input, nil)
	return err
}

func (c *ClusterClient) ActionUpgrade(resource *Cluster, input *UpgradeInput) error {
	err := c.apiClient.Ops.DoAction(ClusterType, "upgrade", &resource.Resource, input, nil)
	return err
}
