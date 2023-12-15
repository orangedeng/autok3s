package client

import (
	"github.com/cnrancher/autok3s/pkg/client/clientbase"
)

type Client struct {
	clientbase.APIBaseClient

	Cluster         ClusterOperations
	ClusterTemplate ClusterTemplateOperations
	Explorer        ExplorerOperations
	Config          ConfigOperations
	log             logOperations
	Setting         SettingOperations
	Package         PackageOperations
	Addon           AddonOperations
	Credential      CredentialOperations
	Mutual          MutualOperations
	SshKey          SshKeyOperations
}

func NewClient(opts *clientbase.ClientOpts) (*Client, error) {
	baseClient, err := clientbase.NewAPIClient(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
		APIBaseClient: baseClient,
	}

	client.Cluster = newClusterClient(client)
	client.ClusterTemplate = newClusterTemplateClient(client)
	client.Explorer = newExplorerClient(client)
	client.Config = newConfigClient(client)
	client.log = newlogClient(client)
	client.Setting = newSettingClient(client)
	client.Package = newPackageClient(client)
	client.Addon = newAddonClient(client)
	client.Credential = newCredentialClient(client)
	client.Mutual = newMutualClient(client)
	client.SshKey = newSshKeyClient(client)

	return client, nil
}
