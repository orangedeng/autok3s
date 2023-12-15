package client

const (
	EnableExplorerOutputType      = "enableExplorerOutput"
	EnableExplorerOutputFieldData = "data"
)

type EnableExplorerOutput struct {
	Data string `json:"data,omitempty" yaml:"data,omitempty"`
}
