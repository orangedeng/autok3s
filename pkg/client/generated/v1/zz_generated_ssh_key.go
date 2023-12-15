package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	SshKeyType               = "sshKey"
	SshKeyFieldBits          = "bits"
	SshKeyFieldGenerateKey   = "generate-key"
	SshKeyFieldHasPassword   = "has-password"
	SshKeyFieldName          = "name"
	SshKeyFieldSSHCert       = "ssh-cert"
	SshKeyFieldSSHKey        = "ssh-key"
	SshKeyFieldSSHPassphrase = "ssh-passphrase"
	SshKeyFieldSSHPublicKey  = "ssh-key-public"
)

type SshKey struct {
	types.Resource
	Bits          *int64 `json:"bits,omitempty" yaml:"bits,omitempty"`
	GenerateKey   bool   `json:"generate-key,omitempty" yaml:"generate-key,omitempty"`
	HasPassword   bool   `json:"has-password,omitempty" yaml:"has-password,omitempty"`
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	SSHCert       string `json:"ssh-cert,omitempty" yaml:"ssh-cert,omitempty"`
	SSHKey        string `json:"ssh-key,omitempty" yaml:"ssh-key,omitempty"`
	SSHPassphrase string `json:"ssh-passphrase,omitempty" yaml:"ssh-passphrase,omitempty"`
	SSHPublicKey  string `json:"ssh-key-public,omitempty" yaml:"ssh-key-public,omitempty"`
}

type SshKeyCollection struct {
	types.Collection
	Data   []SshKey `json:"data,omitempty"`
	client *SshKeyClient
}

type SshKeyClient struct {
	apiClient *Client
}

type SshKeyOperations interface {
	List(opts *clientbase.ListOpts) (*SshKeyCollection, error)
	ListAll(opts *clientbase.ListOpts) (*SshKeyCollection, error)
	Create(opts *SshKey) (*SshKey, error)
	Update(existing *SshKey, updates interface{}) (*SshKey, error)
	Replace(existing *SshKey) (*SshKey, error)
	ByID(id string) (*SshKey, error)
	Delete(container *SshKey) error

	ActionExport(resource *SshKey) (*SshKey, error)
}

func newSshKeyClient(apiClient *Client) *SshKeyClient {
	return &SshKeyClient{
		apiClient: apiClient,
	}
}

func (c *SshKeyClient) Create(container *SshKey) (*SshKey, error) {
	resp := &SshKey{}
	err := c.apiClient.Ops.DoCreate(SshKeyType, container, resp)
	return resp, err
}

func (c *SshKeyClient) Update(existing *SshKey, updates interface{}) (*SshKey, error) {
	resp := &SshKey{}
	err := c.apiClient.Ops.DoUpdate(SshKeyType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SshKeyClient) Replace(obj *SshKey) (*SshKey, error) {
	resp := &SshKey{}
	err := c.apiClient.Ops.DoReplace(SshKeyType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *SshKeyClient) List(opts *clientbase.ListOpts) (*SshKeyCollection, error) {
	resp := &SshKeyCollection{}
	err := c.apiClient.Ops.DoList(SshKeyType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *SshKeyClient) ListAll(opts *clientbase.ListOpts) (*SshKeyCollection, error) {
	resp := &SshKeyCollection{}
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

func (cc *SshKeyCollection) Next() (*SshKeyCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SshKeyCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SshKeyClient) ByID(id string) (*SshKey, error) {
	resp := &SshKey{}
	err := c.apiClient.Ops.DoByID(SshKeyType, id, resp)
	return resp, err
}

func (c *SshKeyClient) Delete(container *SshKey) error {
	return c.apiClient.Ops.DoResourceDelete(SshKeyType, &container.Resource)
}

func (c *SshKeyClient) ActionExport(resource *SshKey) (*SshKey, error) {
	resp := &SshKey{}
	err := c.apiClient.Ops.DoAction(SshKeyType, "export", &resource.Resource, nil, resp)
	return resp, err
}
