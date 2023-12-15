package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	MutualType = "mutual"
)

type Mutual struct {
	types.Resource
}

type MutualCollection struct {
	types.Collection
	Data   []Mutual `json:"data,omitempty"`
	client *MutualClient
}

type MutualClient struct {
	apiClient *Client
}

type MutualOperations interface {
	List(opts *clientbase.ListOpts) (*MutualCollection, error)
	ListAll(opts *clientbase.ListOpts) (*MutualCollection, error)
	Create(opts *Mutual) (*Mutual, error)
	Update(existing *Mutual, updates interface{}) (*Mutual, error)
	Replace(existing *Mutual) (*Mutual, error)
	ByID(id string) (*Mutual, error)
	Delete(container *Mutual) error
}

func newMutualClient(apiClient *Client) *MutualClient {
	return &MutualClient{
		apiClient: apiClient,
	}
}

func (c *MutualClient) Create(container *Mutual) (*Mutual, error) {
	resp := &Mutual{}
	err := c.apiClient.Ops.DoCreate(MutualType, container, resp)
	return resp, err
}

func (c *MutualClient) Update(existing *Mutual, updates interface{}) (*Mutual, error) {
	resp := &Mutual{}
	err := c.apiClient.Ops.DoUpdate(MutualType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *MutualClient) Replace(obj *Mutual) (*Mutual, error) {
	resp := &Mutual{}
	err := c.apiClient.Ops.DoReplace(MutualType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *MutualClient) List(opts *clientbase.ListOpts) (*MutualCollection, error) {
	resp := &MutualCollection{}
	err := c.apiClient.Ops.DoList(MutualType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *MutualClient) ListAll(opts *clientbase.ListOpts) (*MutualCollection, error) {
	resp := &MutualCollection{}
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

func (cc *MutualCollection) Next() (*MutualCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &MutualCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *MutualClient) ByID(id string) (*Mutual, error) {
	resp := &Mutual{}
	err := c.apiClient.Ops.DoByID(MutualType, id, resp)
	return resp, err
}

func (c *MutualClient) Delete(container *Mutual) error {
	return c.apiClient.Ops.DoResourceDelete(MutualType, &container.Resource)
}
