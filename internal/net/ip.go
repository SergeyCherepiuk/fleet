package net

import (
	"errors"
	"net"
)

type ErrIPv4NotFound error

func LocalIPv4() (net.IP, error) {
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

	return net.IP{}, ErrIPv4NotFound(errors.New("no ipv4 interface"))
}
