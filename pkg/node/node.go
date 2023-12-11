package node

import (
	"net"
	"time"
)

type Addr struct {
	Addr net.IP
	Port uint16
}

type Resources struct {
	CPU    float64
	Memory int64
}

type AllocatedResources struct {
	CPU       float64
	Memoty    int64
	Timestamp time.Time
}

type Node struct {
	Addr      Addr
	Resources Resources
}

func (n Node) Stat() (AllocatedResources, error) {
	return AllocatedResources{}, nil
}
