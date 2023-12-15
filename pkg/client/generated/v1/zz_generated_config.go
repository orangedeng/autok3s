package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	ConfigType         = "config"
	ConfigFieldContext = "context"
)

type Config struct {
	types.Resource
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
}

type ConfigCollection struct {
	types.Collection
	Data   []Config `json:"data,omitempty"`
	client *ConfigClient
}

type ConfigClient struct {
	apiClient *Client
}

type ConfigOperations interface {
	List(opts *clientbase.ListOpts) (*ConfigCollection, error)
	ListAll(opts *clientbase.ListOpts) (*ConfigCollection, error)
	Create(opts *Config) (*Config, error)
	Update(existing *Config, updates interface{}) (*Config, error)
	Replace(existing *Config) (*Config, error)
	ByID(id string) (*Config, error)
	Delete(container *Config) error
}

func newConfigClient(apiClient *Client) *ConfigClient {
	return &ConfigClient{
		apiClient: apiClient,
	}
}

func (c *ConfigClient) Create(container *Config) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoCreate(ConfigType, container, resp)
	return resp, err
}

func (c *ConfigClient) Update(existing *Config, updates interface{}) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoUpdate(ConfigType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ConfigClient) Replace(obj *Config) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoReplace(ConfigType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ConfigClient) List(opts *clientbase.ListOpts) (*ConfigCollection, error) {
	resp := &ConfigCollection{}
	err := c.apiClient.Ops.DoList(ConfigType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *ConfigClient) ListAll(opts *clientbase.ListOpts) (*ConfigCollection, error) {
	resp := &ConfigCollection{}
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

func (cc *ConfigCollection) Next() (*ConfigCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ConfigCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ConfigClient) ByID(id string) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoByID(ConfigType, id, resp)
	return resp, err
}

func (c *ConfigClient) Delete(container *Config) error {
	return c.apiClient.Ops.DoResourceDelete(ConfigType, &container.Resource)
}
