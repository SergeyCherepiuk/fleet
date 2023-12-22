package node

import (
	"errors"
	"net"
)

var ErrIPv4NotFound = errors.New("no ipv4 interface")

func (*Node) LocalIPv4() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return net.IP{}, err
	}

	for _, addr := range addrs {
		network, ok := addr.(*net.IPNet)
		if ok && !network.IP.IsLoopback() && network.IP.To4() != nil {
			return network.IP, nil
		}
	}

	return net.IP{}, ErrIPv4NotFound
}

func (*Node) RandomPort() (uint16, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return uint16(listener.Addr().(*net.TCPAddr).Port), nil
}
