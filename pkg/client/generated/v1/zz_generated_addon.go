package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	AddonType             = "addon"
	AddonFieldDescription = "description"
	AddonFieldManifest    = "manifest"
	AddonFieldName        = "name"
	AddonFieldValues      = "values"
)

type Addon struct {
	types.Resource
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Manifest    string            `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Values      map[string]string `json:"values,omitempty" yaml:"values,omitempty"`
}

type AddonCollection struct {
	types.Collection
	Data   []Addon `json:"data,omitempty"`
	client *AddonClient
}

type AddonClient struct {
	apiClient *Client
}

type AddonOperations interface {
	List(opts *clientbase.ListOpts) (*AddonCollection, error)
	ListAll(opts *clientbase.ListOpts) (*AddonCollection, error)
	Create(opts *Addon) (*Addon, error)
	Update(existing *Addon, updates interface{}) (*Addon, error)
	Replace(existing *Addon) (*Addon, error)
	ByID(id string) (*Addon, error)
	Delete(container *Addon) error
}

func newAddonClient(apiClient *Client) *AddonClient {
	return &AddonClient{
		apiClient: apiClient,
	}
}

func (c *AddonClient) Create(container *Addon) (*Addon, error) {
	resp := &Addon{}
	err := c.apiClient.Ops.DoCreate(AddonType, container, resp)
	return resp, err
}

func (c *AddonClient) Update(existing *Addon, updates interface{}) (*Addon, error) {
	resp := &Addon{}
	err := c.apiClient.Ops.DoUpdate(AddonType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *AddonClient) Replace(obj *Addon) (*Addon, error) {
	resp := &Addon{}
	err := c.apiClient.Ops.DoReplace(AddonType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *AddonClient) List(opts *clientbase.ListOpts) (*AddonCollection, error) {
	resp := &AddonCollection{}
	err := c.apiClient.Ops.DoList(AddonType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *AddonClient) ListAll(opts *clientbase.ListOpts) (*AddonCollection, error) {
	resp := &AddonCollection{}
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

func (cc *AddonCollection) Next() (*AddonCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &AddonCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *AddonClient) ByID(id string) (*Addon, error) {
	resp := &Addon{}
	err := c.apiClient.Ops.DoByID(AddonType, id, resp)
	return resp, err
}

func (c *AddonClient) Delete(container *Addon) error {
	return c.apiClient.Ops.DoResourceDelete(AddonType, &container.Resource)
}
