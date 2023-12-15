package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	PackageType            = "package"
	PackageFieldArchs      = "archs"
	PackageFieldFilePath   = "filePath"
	PackageFieldK3sVersion = "k3sVersion"
	PackageFieldName       = "name"
	PackageFieldState      = "state"
)

type Package struct {
	types.Resource
	Archs      []string `json:"archs,omitempty" yaml:"archs,omitempty"`
	FilePath   string   `json:"filePath,omitempty" yaml:"filePath,omitempty"`
	K3sVersion string   `json:"k3sVersion,omitempty" yaml:"k3sVersion,omitempty"`
	Name       string   `json:"name,omitempty" yaml:"name,omitempty"`
	State      string   `json:"state,omitempty" yaml:"state,omitempty"`
}

type PackageCollection struct {
	types.Collection
	Data   []Package `json:"data,omitempty"`
	client *PackageClient
}

type PackageClient struct {
	apiClient *Client
}

type PackageOperations interface {
	List(opts *clientbase.ListOpts) (*PackageCollection, error)
	ListAll(opts *clientbase.ListOpts) (*PackageCollection, error)
	Create(opts *Package) (*Package, error)
	Update(existing *Package, updates interface{}) (*Package, error)
	Replace(existing *Package) (*Package, error)
	ByID(id string) (*Package, error)
	Delete(container *Package) error

	ActionCancel(resource *Package) error

	ActionDownload(resource *Package) error

	CollectionActionImport(resource *PackageCollection) (*Package, error)

	CollectionActionUpdateInstallScript(resource *PackageCollection) error
}

func newPackageClient(apiClient *Client) *PackageClient {
	return &PackageClient{
		apiClient: apiClient,
	}
}

func (c *PackageClient) Create(container *Package) (*Package, error) {
	resp := &Package{}
	err := c.apiClient.Ops.DoCreate(PackageType, container, resp)
	return resp, err
}

func (c *PackageClient) Update(existing *Package, updates interface{}) (*Package, error) {
	resp := &Package{}
	err := c.apiClient.Ops.DoUpdate(PackageType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PackageClient) Replace(obj *Package) (*Package, error) {
	resp := &Package{}
	err := c.apiClient.Ops.DoReplace(PackageType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *PackageClient) List(opts *clientbase.ListOpts) (*PackageCollection, error) {
	resp := &PackageCollection{}
	err := c.apiClient.Ops.DoList(PackageType, opts, resp)
	resp.client = c
	return resp, err
}

func (c *PackageClient) ListAll(opts *clientbase.ListOpts) (*PackageCollection, error) {
	resp := &PackageCollection{}
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

func (cc *PackageCollection) Next() (*PackageCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &PackageCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *PackageClient) ByID(id string) (*Package, error) {
	resp := &Package{}
	err := c.apiClient.Ops.DoByID(PackageType, id, resp)
	return resp, err
}

func (c *PackageClient) Delete(container *Package) error {
	return c.apiClient.Ops.DoResourceDelete(PackageType, &container.Resource)
}

func (c *PackageClient) ActionCancel(resource *Package) error {
	err := c.apiClient.Ops.DoAction(PackageType, "cancel", &resource.Resource, nil, nil)
	return err
}

func (c *PackageClient) ActionDownload(resource *Package) error {
	err := c.apiClient.Ops.DoAction(PackageType, "download", &resource.Resource, nil, nil)
	return err
}

func (c *PackageClient) CollectionActionImport(resource *PackageCollection) (*Package, error) {
	resp := &Package{}
	err := c.apiClient.Ops.DoCollectionAction(PackageType, "import", &resource.Collection, nil, resp)
	return resp, err
}

func (c *PackageClient) CollectionActionUpdateInstallScript(resource *PackageCollection) error {
	err := c.apiClient.Ops.DoCollectionAction(PackageType, "updateInstallScript", &resource.Collection, nil, nil)
	return err
}
