package node

import (
	"errors"
	"net"
	"time"
)

type Addr struct {
	Addr net.IP
	Port uint16
}

type Node struct {
	Addr Addr
}

type Resources struct {
	CPU    CPUStat
	Memory MemoryStat
}

func (n Node) Resources() (Resources, error) {
	var err error

	memstat, e := n.Memory()
	err = errors.Join(err, e)

	cpustat, e := n.CPU(100 * time.Millisecond)
	err = errors.Join(err, e)

	stat := Resources{Memory: memstat, CPU: cpustat}
	return stat, err
}
