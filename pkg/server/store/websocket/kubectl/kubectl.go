//go:build darwin || linux
// +build darwin linux

package kubectl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/orangedeng/autok3s/pkg/common"
	akws "github.com/orangedeng/autok3s/pkg/server/store/websocket"
	wssh "github.com/orangedeng/autok3s/pkg/server/store/websocket/ssh"

	"github.com/gorilla/websocket"
	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/v2/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    10240,
	WriteBufferSize:   10240,
	HandshakeTimeout:  60 * time.Second,
	EnableCompression: true,
	CheckOrigin:       akws.CheckSameOrigin,
}

// Shell struct for shell.
type Shell struct {
	// conn *websocket.Conn
	// ptmx *os.File
}

// KubeHandler kubectl handler for websocket.
func KubeHandler(apiOp *types.APIRequest) (types.APIObject, error) {
	err := ptyHandler(apiOp)
	if err != nil {
		logrus.Errorf("error during kubectl handler %v", err)
	}
	return types.APIObject{}, validation.ErrComplete
}

func ptyHandler(apiOp *types.APIRequest) error {
	queryParams := apiOp.Request.URL.Query()
	height := queryParams.Get("height")
	width := queryParams.Get("width")
	rows := 150
	columns := 300
	var err error
	if height != "" {
		rows, err = strconv.Atoi(height)
		if err != nil {
			return apierror.NewAPIError(validation.InvalidOption, fmt.Sprintf("invalid height %s", height))
		}
	}
	if width != "" {
		columns, err = strconv.Atoi(width)
		if err != nil {
			return apierror.NewAPIError(validation.InvalidOption, fmt.Sprintf("invalid width %s", width))
		}
	}

	c, err := upgrader.Upgrade(apiOp.Response, apiOp.Request, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()

	// apiOp.Name is the kubectl context name, which is also the cluster ID.
	// Connect to the first master node of that cluster and rely on the
	// kubectl that ships with every K3s installation, so no local kubectl
	// binary is required.
	dialer, err := wssh.NewMasterDialer(apiOp.Name, c)
	if err != nil {
		return err
	}
	defer dialer.Close()

	if err := dialer.Terminal(rows, columns); err != nil {
		return err
	}

	// k3d node containers ship a standalone kubectl binary; cloud nodes
	// expose it through the k3s subcommand.
	aliasCmd := "alias kubectl='k3s kubectl'\nalias k='k3s kubectl'\n"
	if state, err := common.DefaultDB.GetClusterByID(apiOp.Name); err == nil && state != nil && state.Provider == "k3d" {
		aliasCmd = "alias k='kubectl'\n"
	}

	if err := dialer.Write([]byte(aliasCmd)); err != nil {
		return err
	}

	return dialer.ReadMessage(apiOp.Context())
}
