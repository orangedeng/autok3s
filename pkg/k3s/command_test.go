package k3s

import (
	"testing"

	"github.com/orangedeng/autok3s/pkg/types"

	"github.com/stretchr/testify/assert"
)

// newTestCluster mirrors the fixture the original command tests were written
// against: a 3 master HA cluster with 2 workers.
func newTestCluster() *types.Cluster {
	return &types.Cluster{
		Metadata: types.Metadata{
			Cluster:       true,
			K3sVersion:    "v1.24.3+k3s1",
			InstallScript: "https://get.k3s.io",
			Token:         "dd73df9b22f8ff22be0d17ec36e7267a",
			TLSSans: []string{
				"2.3.4.5",
			},
		},
		Status: types.Status{
			MasterNodes: []types.Node{
				{
					Master:            true,
					PublicIPAddress:   []string{"1.2.3.1"},
					InternalIPAddress: []string{"1.2.3.1"},
				},
				{
					Master:            true,
					PublicIPAddress:   []string{"1.2.3.2"},
					InternalIPAddress: []string{"1.2.3.2"},
				},
				{
					Master:            true,
					PublicIPAddress:   []string{"1.2.3.3"},
					InternalIPAddress: []string{"1.2.3.3"},
				},
			},
			WorkerNodes: []types.Node{
				{
					PublicIPAddress:   []string{"1.2.3.5"},
					InternalIPAddress: []string{"1.2.3.5"},
				},
				{
					PublicIPAddress:   []string{"1.2.3.6"},
					InternalIPAddress: []string{"1.2.3.6"},
				},
			},
		},
	}
}

func TestInstallCommand(t *testing.T) {
	testCluster := newTestCluster()

	masterCommands := []string{
		"curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --advertise-address=1.2.3.1 --cluster-init --node-external-ip=1.2.3.1 " +
			"--tls-san=1.2.3.1 --tls-san=1.2.3.2 --tls-san=1.2.3.3 --tls-san=2.3.4.5' " +
			"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' sh -",
		"curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --advertise-address=1.2.3.2 --node-external-ip=1.2.3.2 " +
			"--server=https://1.2.3.1:6443 --tls-san=1.2.3.1 --tls-san=1.2.3.2 --tls-san=1.2.3.3 --tls-san=2.3.4.5' " +
			"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' sh -",
		"curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --advertise-address=1.2.3.3 --node-external-ip=1.2.3.3 " +
			"--server=https://1.2.3.1:6443 --tls-san=1.2.3.1 --tls-san=1.2.3.2 --tls-san=1.2.3.3 --tls-san=2.3.4.5' " +
			"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' sh -",
	}
	workerCommands := []string{
		"curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='--node-external-ip=1.2.3.5' " +
			"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' K3S_URL='https://1.2.3.1:6443' sh -",
		"curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='--node-external-ip=1.2.3.6' " +
			"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' K3S_URL='https://1.2.3.1:6443' sh -",
	}
	fixedIP := FirstAddress(testCluster.Status.MasterNodes[0].PublicIPAddress)
	for i, node := range testCluster.MasterNodes {
		rtn := InstallCommand(i == 0, fixedIP, testCluster, node, []string{})
		assert.Equal(t, masterCommands[i], rtn)
	}
	for i, node := range testCluster.WorkerNodes {
		rtn := InstallCommand(false, fixedIP, testCluster, node, []string{})
		assert.Equal(t, workerCommands[i], rtn)
	}

	// testing datastore and network
	testCluster.DataStore = "mysql://root:root@tcp(1.2.3.4:3306)/k3s"
	testCluster.Network = "host-gw"
	expectFirstMasterCommand := "curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server " +
		"--advertise-address=1.2.3.1 " +
		"--datastore-endpoint=mysql://root:root@tcp(1.2.3.4:3306)/k3s " +
		"--flannel-backend=host-gw " +
		"--node-external-ip=1.2.3.1 " +
		"--tls-san=1.2.3.1 --tls-san=1.2.3.2 --tls-san=1.2.3.3 --tls-san=2.3.4.5' " +
		"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' sh -"

	assert.Equal(t, expectFirstMasterCommand, InstallCommand(true, fixedIP, testCluster, testCluster.MasterNodes[0], []string{}))
}

func TestInstallCommandAirgap(t *testing.T) {
	testCluster := newTestCluster()
	testCluster.PackageName = "autok3s-airgap.tar.gz"

	expect := "INSTALL_K3S_EXEC='--node-external-ip=1.2.3.5' INSTALL_K3S_SKIP_DOWNLOAD='true' " +
		"K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' K3S_URL='https://1.2.3.1:6443' install.sh"
	assert.Equal(t, expect, InstallCommand(false, "1.2.3.1", testCluster, testCluster.WorkerNodes[0], []string{}))
}

func TestInstallCommandMirrorAndChannel(t *testing.T) {
	testCluster := newTestCluster()
	testCluster.K3sVersion = ""
	testCluster.K3sChannel = "stable"
	testCluster.Mirror = "INSTALL_K3S_MIRROR=cn"

	expect := "curl -sLS https://get.k3s.io | INSTALL_K3S_CHANNEL='stable' INSTALL_K3S_EXEC='--node-external-ip=1.2.3.5' " +
		"INSTALL_K3S_MIRROR='cn' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' K3S_URL='https://1.2.3.1:6443' sh -"
	assert.Equal(t, expect, InstallCommand(false, "1.2.3.1", testCluster, testCluster.WorkerNodes[0], []string{}))
}

func TestInstallCommandInstallEnv(t *testing.T) {
	testCluster := newTestCluster()
	testCluster.InstallEnv = types.StringMap{
		"INSTALL_K3S_EXEC": "--node-label env=test",
		"HTTP_PROXY":       "http://127.0.0.1:3128",
	}

	expect := "curl -sLS https://get.k3s.io | HTTP_PROXY='http://127.0.0.1:3128' " +
		"INSTALL_K3S_EXEC='server --advertise-address=1.2.3.1 --cluster-init --node-external-ip=1.2.3.1 " +
		"--tls-san=1.2.3.1 --tls-san=1.2.3.2 --tls-san=1.2.3.3 --tls-san=2.3.4.5 --node-label env=test' " +
		"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' sh -"
	assert.Equal(t, expect, InstallCommand(true, "1.2.3.1", testCluster, testCluster.MasterNodes[0], []string{}))
}

func TestInstallCommandDataStoreCertificates(t *testing.T) {
	testCluster := newTestCluster()
	testCluster.DataStore = "mysql://root:root@tcp(1.2.3.4:3306)/k3s"
	testCluster.DataStoreCAFileContent = "ca"
	testCluster.SystemDefaultRegistry = "reg.example.com"

	expect := "curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server " +
		"--advertise-address=1.2.3.1 " +
		"--datastore-cafile=/etc/rancher/datastore/ds-ca.pem " +
		"--datastore-endpoint=mysql://root:root@tcp(1.2.3.4:3306)/k3s " +
		"--node-external-ip=1.2.3.1 " +
		"--system-default-registry=reg.example.com " +
		"--tls-san=1.2.3.1 --tls-san=1.2.3.2 --tls-san=1.2.3.3 --tls-san=2.3.4.5' " +
		"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' sh -"
	assert.Equal(t, expect, InstallCommand(true, "1.2.3.1", testCluster, testCluster.MasterNodes[0], []string{}))
}

func TestInstallCommandWorkerWithoutExternalIP(t *testing.T) {
	testCluster := newTestCluster()
	node := types.Node{InternalIPAddress: []string{"10.0.0.9"}}

	expect := "curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='' " +
		"INSTALL_K3S_VERSION='v1.24.3+k3s1' K3S_TOKEN='dd73df9b22f8ff22be0d17ec36e7267a' K3S_URL='https://1.2.3.1:6443' sh -"
	assert.Equal(t, expect, InstallCommand(false, "1.2.3.1", testCluster, node, []string{}))
}

func TestTLSSans(t *testing.T) {
	testCluster := newTestCluster()
	// a SAN equal to a master address must not be duplicated
	testCluster.TLSSans = []string{"1.2.3.1", "2.3.4.5"}

	assert.Equal(t, []string{"1.2.3.1", "1.2.3.2", "1.2.3.3", "2.3.4.5"}, TLSSans(testCluster))
}

func TestFirstAddress(t *testing.T) {
	assert.Equal(t, "b", FirstAddress([]string{"", "b", "c"}))
	assert.Equal(t, "", FirstAddress(nil))
}
