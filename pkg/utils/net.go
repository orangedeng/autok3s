package utils

import (
	"fmt"
	"net"
)

// GetFreePort returns an available TCP port on localhost. Mirrors the
// behavior of the k3d helper it replaces.
func GetFreePort() (int, error) {
	tcpAddress, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("failed to resolve address 'localhost:0': %w", err)
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to create tcp listener: %w", err)
	}
	defer tcpListener.Close()

	return tcpListener.Addr().(*net.TCPAddr).Port, nil
}
