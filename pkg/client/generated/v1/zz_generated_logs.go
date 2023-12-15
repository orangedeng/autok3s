package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	logType = "logs"
)

type log struct {
	types.Resource
}

type logCollection struct {
	types.Collection
	Data   []log `json:"data,omitempty"`
	client *logClient
}

type logClient struct {
	apiClient *Client
}

type logOperations interface {
	List(opts *clientbase.ListOpts) (*logCollection, error)
	ListAll(opts *clientbase.ListOpts) (*logCollection, error)
	Create(opts *log) (*log, error)
	Update(existing *log, updates interface{}) (*log, error)
	Replace(existing *log) (*log, error)
	ByID(id string) (*log, error)
	Delete(container *log) error
}

func newlogClient(apiClient *Client) *logClient {
	return &logClient{
		apiClient: apiClient,
	}
}

func (c *logClient) Create(container *log) (*log, error) {
	resp := &log{}
	err := c.apiClient.Ops.DoCreate(logType, container, resp)
	return resp, err
}

func (c *logClient) Update(existing *log, updates interface{}) (*log, error) {
	resp := &log{}
	err := c.apiClient.Ops.DoUpdate(logType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *logClient) Replace(obj *log) (*log, error) {
	resp := &log{}
	err := c.apiClient.Ops.DoReplace(logType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *logClient) List(opts *clientbase.ListOpts) (*logCollection, error) {
	resp := &logCollection{}
	err := c.apiClient.Ops.DoList(logType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *logClient) ListAll(opts *clientbase.ListOpts) (*logCollection, error) {
	resp := &logCollection{}
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

func (cc *logCollection) Next() (*logCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &logCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *logClient) ByID(id string) (*log, error) {
	resp := &log{}
	err := c.apiClient.Ops.DoByID(logType, id, resp)
	return resp, err
}

func (c *logClient) Delete(container *log) error {
	return c.apiClient.Ops.DoResourceDelete(logType, &container.Resource)
}
