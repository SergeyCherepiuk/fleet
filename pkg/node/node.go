package node

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type Addr struct {
	Addr net.IP
	Port uint16
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.Addr, a.Port)
}

type Node struct {
	Addr Addr
}

type Resources struct {
	CPU    CPUStat
	Memory MemoryStat
	Disk   DiskStat
}

func (n Node) Resources() (Resources, error) {
	var err error

	memstat, e := Memory()
	err = errors.Join(err, e)

	cpustat, e := CPU(100 * time.Millisecond)
	err = errors.Join(err, e)

	diskstat, e := Disk()
	err = errors.Join(err, e)

	stat := Resources{Memory: memstat, CPU: cpustat, Disk: diskstat}
	return stat, err
}
