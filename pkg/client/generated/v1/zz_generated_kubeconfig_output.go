package client

const (
	KubeconfigOutputType        = "kubeconfigOutput"
	KubeconfigOutputFieldConfig = "config"
)

type KubeconfigOutput struct {
	Config string `json:"config,omitempty" yaml:"config,omitempty"`
}
