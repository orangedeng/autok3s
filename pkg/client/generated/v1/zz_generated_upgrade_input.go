package client

const (
	UpgradeInputType               = "upgradeInput"
	UpgradeInputFieldInstallScript = "k3s-install-script"
	UpgradeInputFieldK3sChannel    = "k3s-channel"
	UpgradeInputFieldK3sVersion    = "k3s-version"
	UpgradeInputFieldPackageName   = "package-name"
	UpgradeInputFieldPackagePath   = "package-path"
)

type UpgradeInput struct {
	InstallScript string `json:"k3s-install-script,omitempty" yaml:"k3s-install-script,omitempty"`
	K3sChannel    string `json:"k3s-channel,omitempty" yaml:"k3s-channel,omitempty"`
	K3sVersion    string `json:"k3s-version,omitempty" yaml:"k3s-version,omitempty"`
	PackageName   string `json:"package-name,omitempty" yaml:"package-name,omitempty"`
	PackagePath   string `json:"package-path,omitempty" yaml:"package-path,omitempty"`
}
