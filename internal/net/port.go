package net

import (
	stdnet "net"
)

func RandomPort() (uint16, error) {
	listener, err := stdnet.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return uint16(listener.Addr().(*stdnet.TCPAddr).Port), nil
}
