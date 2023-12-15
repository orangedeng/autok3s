package client

const (
	SSHType                  = "ssh"
	SSHFieldSSHAgentAuth     = "ssh-agent-auth"
	SSHFieldSSHCert          = "ssh-cert"
	SSHFieldSSHCertPath      = "ssh-cert-path"
	SSHFieldSSHKey           = "ssh-key"
	SSHFieldSSHKeyName       = "ssh-key-name"
	SSHFieldSSHKeyPassphrase = "ssh-key-passphrase"
	SSHFieldSSHKeyPath       = "ssh-key-path"
	SSHFieldSSHPassword      = "ssh-password"
	SSHFieldSSHPort          = "ssh-port"
	SSHFieldSSHUser          = "ssh-user"
)

type SSH struct {
	SSHAgentAuth     bool   `json:"ssh-agent-auth,omitempty" yaml:"ssh-agent-auth,omitempty"`
	SSHCert          string `json:"ssh-cert,omitempty" yaml:"ssh-cert,omitempty"`
	SSHCertPath      string `json:"ssh-cert-path,omitempty" yaml:"ssh-cert-path,omitempty"`
	SSHKey           string `json:"ssh-key,omitempty" yaml:"ssh-key,omitempty"`
	SSHKeyName       string `json:"ssh-key-name,omitempty" yaml:"ssh-key-name,omitempty"`
	SSHKeyPassphrase string `json:"ssh-key-passphrase,omitempty" yaml:"ssh-key-passphrase,omitempty"`
	SSHKeyPath       string `json:"ssh-key-path,omitempty" yaml:"ssh-key-path,omitempty"`
	SSHPassword      string `json:"ssh-password,omitempty" yaml:"ssh-password,omitempty"`
	SSHPort          string `json:"ssh-port,omitempty" yaml:"ssh-port,omitempty"`
	SSHUser          string `json:"ssh-user,omitempty" yaml:"ssh-user,omitempty"`
}
