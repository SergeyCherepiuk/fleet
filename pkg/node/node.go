package node

import (
	"net"
)

type Addr struct {
	Addr net.IP
	Port uint16
}

type Resources struct {
	CPU    float64
	Memory int64
}

type Node struct {
	Addr      Addr
	Resources Resources
}

func (n Node) ResourceUsage() (Resources, error) {
	return Resources{}, nil
}
