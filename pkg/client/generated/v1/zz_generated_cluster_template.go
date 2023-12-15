package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	ClusterTemplateType                          = "clusterTemplate"
	ClusterTemplateFieldCluster                  = "cluster"
	ClusterTemplateFieldClusterCidr              = "cluster-cidr"
	ClusterTemplateFieldContextName              = "context-name"
	ClusterTemplateFieldDataStore                = "datastore"
	ClusterTemplateFieldDataStoreCAFile          = "datastore-cafile"
	ClusterTemplateFieldDataStoreCAFileContent   = "datastore-cafile-content"
	ClusterTemplateFieldDataStoreCertFile        = "datastore-certfile"
	ClusterTemplateFieldDataStoreCertFileContent = "datastore-certfile-content"
	ClusterTemplateFieldDataStoreKeyFile         = "datastore-keyfile"
	ClusterTemplateFieldDataStoreKeyFileContent  = "datastore-keyfile-content"
	ClusterTemplateFieldDataStoreType            = "datastore-type"
	ClusterTemplateFieldDockerArg                = "docker-arg"
	ClusterTemplateFieldDockerMirror             = "dockerMirror"
	ClusterTemplateFieldDockerScript             = "docker-script"
	ClusterTemplateFieldEnable                   = "enable"
	ClusterTemplateFieldIP                       = "ip"
	ClusterTemplateFieldInstallScript            = "k3s-install-script"
	ClusterTemplateFieldIsDefault                = "is-default"
	ClusterTemplateFieldIsHAMode                 = "is-ha-mode"
	ClusterTemplateFieldK3sChannel               = "k3s-channel"
	ClusterTemplateFieldK3sVersion               = "k3s-version"
	ClusterTemplateFieldManifests                = "manifests"
	ClusterTemplateFieldMaster                   = "master"
	ClusterTemplateFieldMasterExtraArgs          = "master-extra-args"
	ClusterTemplateFieldMirror                   = "k3s-install-mirror"
	ClusterTemplateFieldName                     = "name"
	ClusterTemplateFieldNetwork                  = "network"
	ClusterTemplateFieldOptions                  = "options"
	ClusterTemplateFieldPackageName              = "package-name"
	ClusterTemplateFieldPackagePath              = "package-path"
	ClusterTemplateFieldProvider                 = "provider"
	ClusterTemplateFieldRegistry                 = "registry"
	ClusterTemplateFieldRegistryContent          = "registry-content"
	ClusterTemplateFieldRollback                 = "rollback"
	ClusterTemplateFieldSSHAgentAuth             = "ssh-agent-auth"
	ClusterTemplateFieldSSHCert                  = "ssh-cert"
	ClusterTemplateFieldSSHCertPath              = "ssh-cert-path"
	ClusterTemplateFieldSSHKey                   = "ssh-key"
	ClusterTemplateFieldSSHKeyName               = "ssh-key-name"
	ClusterTemplateFieldSSHKeyPassphrase         = "ssh-key-passphrase"
	ClusterTemplateFieldSSHKeyPath               = "ssh-key-path"
	ClusterTemplateFieldSSHPassword              = "ssh-password"
	ClusterTemplateFieldSSHPort                  = "ssh-port"
	ClusterTemplateFieldSSHUser                  = "ssh-user"
	ClusterTemplateFieldStatus                   = "status"
	ClusterTemplateFieldSystemDefaultRegistry    = "system-default-registry"
	ClusterTemplateFieldTLSSans                  = "tls-sans"
	ClusterTemplateFieldToken                    = "token"
	ClusterTemplateFieldUI                       = "ui"
	ClusterTemplateFieldValues                   = "values"
	ClusterTemplateFieldWorker                   = "worker"
	ClusterTemplateFieldWorkerExtraArgs          = "worker-extra-args"
)

type ClusterTemplate struct {
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
	IsDefault                bool              `json:"is-default,omitempty" yaml:"is-default,omitempty"`
	IsHAMode                 bool              `json:"is-ha-mode,omitempty" yaml:"is-ha-mode,omitempty"`
	K3sChannel               string            `json:"k3s-channel,omitempty" yaml:"k3s-channel,omitempty"`
	K3sVersion               string            `json:"k3s-version,omitempty" yaml:"k3s-version,omitempty"`
	Manifests                string            `json:"manifests,omitempty" yaml:"manifests,omitempty"`
	Master                   string            `json:"master,omitempty" yaml:"master,omitempty"`
	MasterExtraArgs          string            `json:"master-extra-args,omitempty" yaml:"master-extra-args,omitempty"`
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
	Status                   string            `json:"status,omitempty" yaml:"status,omitempty"`
	SystemDefaultRegistry    string            `json:"system-default-registry,omitempty" yaml:"system-default-registry,omitempty"`
	TLSSans                  []string          `json:"tls-sans,omitempty" yaml:"tls-sans,omitempty"`
	Token                    string            `json:"token,omitempty" yaml:"token,omitempty"`
	UI                       bool              `json:"ui,omitempty" yaml:"ui,omitempty"`
	Values                   map[string]string `json:"values,omitempty" yaml:"values,omitempty"`
	Worker                   string            `json:"worker,omitempty" yaml:"worker,omitempty"`
	WorkerExtraArgs          string            `json:"worker-extra-args,omitempty" yaml:"worker-extra-args,omitempty"`
}

type ClusterTemplateCollection struct {
	types.Collection
	Data   []ClusterTemplate `json:"data,omitempty"`
	client *ClusterTemplateClient
}

type ClusterTemplateClient struct {
	apiClient *Client
}

type ClusterTemplateOperations interface {
	List(opts *clientbase.ListOpts) (*ClusterTemplateCollection, error)
	ListAll(opts *clientbase.ListOpts) (*ClusterTemplateCollection, error)
	Create(opts *ClusterTemplate) (*ClusterTemplate, error)
	Update(existing *ClusterTemplate, updates interface{}) (*ClusterTemplate, error)
	Replace(existing *ClusterTemplate) (*ClusterTemplate, error)
	ByID(id string) (*ClusterTemplate, error)
	Delete(container *ClusterTemplate) error
}

func newClusterTemplateClient(apiClient *Client) *ClusterTemplateClient {
	return &ClusterTemplateClient{
		apiClient: apiClient,
	}
}

func (c *ClusterTemplateClient) Create(container *ClusterTemplate) (*ClusterTemplate, error) {
	resp := &ClusterTemplate{}
	err := c.apiClient.Ops.DoCreate(ClusterTemplateType, container, resp)
	return resp, err
}

func (c *ClusterTemplateClient) Update(existing *ClusterTemplate, updates interface{}) (*ClusterTemplate, error) {
	resp := &ClusterTemplate{}
	err := c.apiClient.Ops.DoUpdate(ClusterTemplateType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterTemplateClient) Replace(obj *ClusterTemplate) (*ClusterTemplate, error) {
	resp := &ClusterTemplate{}
	err := c.apiClient.Ops.DoReplace(ClusterTemplateType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ClusterTemplateClient) List(opts *clientbase.ListOpts) (*ClusterTemplateCollection, error) {
	resp := &ClusterTemplateCollection{}
	err := c.apiClient.Ops.DoList(ClusterTemplateType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *ClusterTemplateClient) ListAll(opts *clientbase.ListOpts) (*ClusterTemplateCollection, error) {
	resp := &ClusterTemplateCollection{}
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

func (cc *ClusterTemplateCollection) Next() (*ClusterTemplateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ClusterTemplateCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ClusterTemplateClient) ByID(id string) (*ClusterTemplate, error) {
	resp := &ClusterTemplate{}
	err := c.apiClient.Ops.DoByID(ClusterTemplateType, id, resp)
	return resp, err
}

func (c *ClusterTemplateClient) Delete(container *ClusterTemplate) error {
	return c.apiClient.Ops.DoResourceDelete(ClusterTemplateType, &container.Resource)
}
