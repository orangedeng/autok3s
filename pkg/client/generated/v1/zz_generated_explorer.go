package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	ExplorerType             = "explorer"
	ExplorerFieldContextName = "context-name"
	ExplorerFieldEnabled     = "enabled"
	ExplorerFieldPort        = "port"
)

type Explorer struct {
	types.Resource
	ContextName string `json:"context-name,omitempty" yaml:"context-name,omitempty"`
	Enabled     bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Port        int64  `json:"port,omitempty" yaml:"port,omitempty"`
}

type ExplorerCollection struct {
	types.Collection
	Data   []Explorer `json:"data,omitempty"`
	client *ExplorerClient
}

type ExplorerClient struct {
	apiClient *Client
}

type ExplorerOperations interface {
	List(opts *clientbase.ListOpts) (*ExplorerCollection, error)
	ListAll(opts *clientbase.ListOpts) (*ExplorerCollection, error)
	Create(opts *Explorer) (*Explorer, error)
	Update(existing *Explorer, updates interface{}) (*Explorer, error)
	Replace(existing *Explorer) (*Explorer, error)
	ByID(id string) (*Explorer, error)
	Delete(container *Explorer) error
}

func newExplorerClient(apiClient *Client) *ExplorerClient {
	return &ExplorerClient{
		apiClient: apiClient,
	}
}

func (c *ExplorerClient) Create(container *Explorer) (*Explorer, error) {
	resp := &Explorer{}
	err := c.apiClient.Ops.DoCreate(ExplorerType, container, resp)
	return resp, err
}

func (c *ExplorerClient) Update(existing *Explorer, updates interface{}) (*Explorer, error) {
	resp := &Explorer{}
	err := c.apiClient.Ops.DoUpdate(ExplorerType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExplorerClient) Replace(obj *Explorer) (*Explorer, error) {
	resp := &Explorer{}
	err := c.apiClient.Ops.DoReplace(ExplorerType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ExplorerClient) List(opts *clientbase.ListOpts) (*ExplorerCollection, error) {
	resp := &ExplorerCollection{}
	err := c.apiClient.Ops.DoList(ExplorerType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *ExplorerClient) ListAll(opts *clientbase.ListOpts) (*ExplorerCollection, error) {
	resp := &ExplorerCollection{}
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

func (cc *ExplorerCollection) Next() (*ExplorerCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ExplorerCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ExplorerClient) ByID(id string) (*Explorer, error) {
	resp := &Explorer{}
	err := c.apiClient.Ops.DoByID(ExplorerType, id, resp)
	return resp, err
}

func (c *ExplorerClient) Delete(container *Explorer) error {
	return c.apiClient.Ops.DoResourceDelete(ExplorerType, &container.Resource)
}
