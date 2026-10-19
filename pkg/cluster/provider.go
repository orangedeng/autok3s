package cluster

import (
	"sync"

	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/syncmap"
)

const (
	k3sVersion          = ""
	k3sChannel          = "stable"
	k3sInstallScript    = "https://get.k3s.io"
	master              = "0"
	worker              = "0"
	ui                  = false
	embedEtcd           = false
	uploadManifestCmd   = "echo \"%s\" | base64 -d | tee \"%s/%s\""
	dockerInstallScript = "https://get.docker.com"

	deployPluginCmd = "echo \"%s\" | base64 -d | tee \"%s/%s.yaml\""
)

// ProviderBase provider base struct.
type ProviderBase struct {
	types.Metadata `json:",inline"`
	types.Status   `json:"status"`
	types.SSH      `json:",inline"`
	M              *sync.Map
	ErrM           map[string]string
	Logger         *logrus.Logger
	Callbacks      map[string]*providerProcess
}

type providerProcess struct {
	ContextName string
	Event       string
	Fn          func(interface{})
}

// NewBaseProvider new base provider.
func NewBaseProvider() *ProviderBase {
	return &ProviderBase{
		Metadata: types.Metadata{
			UI:            ui,
			K3sVersion:    k3sVersion,
			K3sChannel:    k3sChannel,
			InstallScript: k3sInstallScript,
			Cluster:       embedEtcd,
			Master:        master,
			Worker:        worker,
			DockerScript:  dockerInstallScript,
			Rollback:      true,
		},
		Status: types.Status{
			MasterNodes: make([]types.Node, 0),
			WorkerNodes: make([]types.Node, 0),
		},
		SSH: types.SSH{
			SSHPort: "22",
		},
		M:    new(syncmap.Map),
		ErrM: map[string]string{},
	}
}

// RegisterCallbacks register callbacks.
func (p *ProviderBase) RegisterCallbacks(name, event string, fn func(interface{})) {
	if p.Callbacks == nil {
		p.Callbacks = map[string]*providerProcess{}
	}
	p.Callbacks[name] = &providerProcess{
		ContextName: name,
		Event:       event,
		Fn:          fn,
	}
}
