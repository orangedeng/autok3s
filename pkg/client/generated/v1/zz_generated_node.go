package client

const (
	NodeType                   = "node"
	NodeFieldEipAllocationIds  = "eip-allocation-ids"
	NodeFieldInstanceID        = "instance-id"
	NodeFieldInstanceStatus    = "instance-status"
	NodeFieldInternalIPAddress = "internal-ip-address"
	NodeFieldLocalHostname     = "local-hostname"
	NodeFieldMaster            = "master"
	NodeFieldPublicIPAddress   = "public-ip-address"
	NodeFieldSSHAgentAuth      = "ssh-agent-auth"
	NodeFieldSSHCert           = "ssh-cert"
	NodeFieldSSHCertPath       = "ssh-cert-path"
	NodeFieldSSHKey            = "ssh-key"
	NodeFieldSSHKeyName        = "ssh-key-name"
	NodeFieldSSHKeyPassphrase  = "ssh-key-passphrase"
	NodeFieldSSHKeyPath        = "ssh-key-path"
	NodeFieldSSHPassword       = "ssh-password"
	NodeFieldSSHPort           = "ssh-port"
	NodeFieldSSHUser           = "ssh-user"
	NodeFieldStandalone        = "standalone"
)

type Node struct {
	EipAllocationIds  []string `json:"eip-allocation-ids,omitempty" yaml:"eip-allocation-ids,omitempty"`
	InstanceID        string   `json:"instance-id,omitempty" yaml:"instance-id,omitempty"`
	InstanceStatus    string   `json:"instance-status,omitempty" yaml:"instance-status,omitempty"`
	InternalIPAddress []string `json:"internal-ip-address,omitempty" yaml:"internal-ip-address,omitempty"`
	LocalHostname     string   `json:"local-hostname,omitempty" yaml:"local-hostname,omitempty"`
	Master            bool     `json:"master,omitempty" yaml:"master,omitempty"`
	PublicIPAddress   []string `json:"public-ip-address,omitempty" yaml:"public-ip-address,omitempty"`
	SSHAgentAuth      bool     `json:"ssh-agent-auth,omitempty" yaml:"ssh-agent-auth,omitempty"`
	SSHCert           string   `json:"ssh-cert,omitempty" yaml:"ssh-cert,omitempty"`
	SSHCertPath       string   `json:"ssh-cert-path,omitempty" yaml:"ssh-cert-path,omitempty"`
	SSHKey            string   `json:"ssh-key,omitempty" yaml:"ssh-key,omitempty"`
	SSHKeyName        string   `json:"ssh-key-name,omitempty" yaml:"ssh-key-name,omitempty"`
	SSHKeyPassphrase  string   `json:"ssh-key-passphrase,omitempty" yaml:"ssh-key-passphrase,omitempty"`
	SSHKeyPath        string   `json:"ssh-key-path,omitempty" yaml:"ssh-key-path,omitempty"`
	SSHPassword       string   `json:"ssh-password,omitempty" yaml:"ssh-password,omitempty"`
	SSHPort           string   `json:"ssh-port,omitempty" yaml:"ssh-port,omitempty"`
	SSHUser           string   `json:"ssh-user,omitempty" yaml:"ssh-user,omitempty"`
	Standalone        bool     `json:"standalone,omitempty" yaml:"standalone,omitempty"`
}
