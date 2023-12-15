package client

const (
	ErrorType  = "error"
	ErrorField = "status"
)

type Error struct {
	int64 `json:"status,omitempty" yaml:"status,omitempty"`
}
