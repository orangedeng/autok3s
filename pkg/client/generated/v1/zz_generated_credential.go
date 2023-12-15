package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	CredentialType          = "credential"
	CredentialFieldProvider = "provider"
	CredentialFieldSecrets  = "secrets"
)

type Credential struct {
	types.Resource
	Provider string            `json:"provider,omitempty" yaml:"provider,omitempty"`
	Secrets  map[string]string `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type CredentialCollection struct {
	types.Collection
	Data   []Credential `json:"data,omitempty"`
	client *CredentialClient
}

type CredentialClient struct {
	apiClient *Client
}

type CredentialOperations interface {
	List(opts *clientbase.ListOpts) (*CredentialCollection, error)
	ListAll(opts *clientbase.ListOpts) (*CredentialCollection, error)
	Create(opts *Credential) (*Credential, error)
	Update(existing *Credential, updates interface{}) (*Credential, error)
	Replace(existing *Credential) (*Credential, error)
	ByID(id string) (*Credential, error)
	Delete(container *Credential) error
}

func newCredentialClient(apiClient *Client) *CredentialClient {
	return &CredentialClient{
		apiClient: apiClient,
	}
}

func (c *CredentialClient) Create(container *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.apiClient.Ops.DoCreate(CredentialType, container, resp)
	return resp, err
}

func (c *CredentialClient) Update(existing *Credential, updates interface{}) (*Credential, error) {
	resp := &Credential{}
	err := c.apiClient.Ops.DoUpdate(CredentialType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *CredentialClient) Replace(obj *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.apiClient.Ops.DoReplace(CredentialType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *CredentialClient) List(opts *clientbase.ListOpts) (*CredentialCollection, error) {
	resp := &CredentialCollection{}
	err := c.apiClient.Ops.DoList(CredentialType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *CredentialClient) ListAll(opts *clientbase.ListOpts) (*CredentialCollection, error) {
	resp := &CredentialCollection{}
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

func (cc *CredentialCollection) Next() (*CredentialCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &CredentialCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *CredentialClient) ByID(id string) (*Credential, error) {
	resp := &Credential{}
	err := c.apiClient.Ops.DoByID(CredentialType, id, resp)
	return resp, err
}

func (c *CredentialClient) Delete(container *Credential) error {
	return c.apiClient.Ops.DoResourceDelete(CredentialType, &container.Resource)
}
