package ssh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/hosts"
	"github.com/orangedeng/autok3s/pkg/hosts/dialer"
	autok3stypes "github.com/orangedeng/autok3s/pkg/types"

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
}

// Handler ssh handler.
func Handler(apiOp *types.APIRequest) (types.APIObjectList, error) {
	err := handler(apiOp)
	if err != nil {
		logrus.Errorf("error during ssh %v", err)
	}
	return types.APIObjectList{}, validation.ErrComplete
}

func handler(apiOp *types.APIRequest) error {
	queryParams := apiOp.Request.URL.Query()
	provider := queryParams.Get("provider")
	id := queryParams.Get("cluster")
	node := queryParams.Get("node")
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
	if provider == "" || id == "" || node == "" {
		return apierror.NewAPIError(validation.InvalidOption, "provider, cluster, node can't be empty")
	}
	upgrader.CheckOrigin = func(_ *http.Request) bool {
		return true
	}
	c, err := upgrader.Upgrade(apiOp.Response, apiOp.Request, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()

	dialer, err := NewDialer(id, node, c)
	if err != nil {
		return err
	}
	defer dialer.Close()

	err = dialer.Terminal(rows, columns)
	if err != nil {
		return err
	}

	return dialer.ReadMessage(apiOp.Context())
}

// NewDialer returns a websocket dialer connected to the given node of the cluster.
func NewDialer(id, node string, conn *websocket.Conn) (*hosts.WebSocketDialer, error) {
	// get exist cluster's state from database.
	state, err := common.DefaultDB.GetClusterByID(id)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("cluster %s is not exist", id)
	}

	// aggregate exist cluster's nodes.
	allNodes := make([]autok3stypes.Node, 0)
	err = json.Unmarshal(state.MasterNodes, &allNodes)
	if err != nil {
		return nil, err
	}
	nodes := make([]autok3stypes.Node, 0)
	err = json.Unmarshal(state.WorkerNodes, &nodes)
	if err != nil {
		return nil, err
	}
	allNodes = append(allNodes, nodes...)

	// find the matching node and open dialer.
	for _, n := range allNodes {
		if n.InstanceID == node {
			return newNodeDialer(state, n, conn)
		}
	}
	return nil, apierror.NewAPIError(validation.NotFound, fmt.Sprintf("node %s is not found for cluster [%s]", node, id))
}

// NewMasterDialer returns a websocket dialer connected to the first master
// node of the cluster.
func NewMasterDialer(id string, conn *websocket.Conn) (*hosts.WebSocketDialer, error) {
	// get exist cluster's state from database.
	state, err := common.DefaultDB.GetClusterByID(id)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("cluster %s is not exist", id)
	}

	masters := make([]autok3stypes.Node, 0)
	if err := json.Unmarshal(state.MasterNodes, &masters); err != nil {
		return nil, err
	}
	if len(masters) == 0 {
		return nil, fmt.Errorf("cluster %s has no master node", id)
	}

	return newNodeDialer(state, masters[0], conn)
}

// newNodeDialer opens the proper dialer for the node based on the provider.
func newNodeDialer(state *common.ClusterState, n autok3stypes.Node, conn *websocket.Conn) (*hosts.WebSocketDialer, error) {
	if state.Provider == "k3d" {
		dockerShell, err := dialer.NewDockerShell(&n)
		if err != nil {
			return nil, err
		}
		return hosts.NewWebSocketDialer(conn, dockerShell), nil
	}

	sshDialer, err := dialer.NewSSHDialer(&n, true, common.NewLogger(nil))
	if err != nil {
		return nil, err
	}
	shell, err := sshDialer.OpenShell()
	if err != nil {
		return nil, err
	}
	return hosts.NewWebSocketDialer(conn, shell), nil
}
