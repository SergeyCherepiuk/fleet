package container

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	ID    string
	Image image.Image

	ExposedPorts  nat.PortSet
	RestartPolicy RestartPolicy

	CPU    float64
	Memory int
	Disk   int

	StartedAt  time.Time
	FinishedAt time.Time
}
