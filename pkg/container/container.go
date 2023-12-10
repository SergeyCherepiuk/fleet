package container

import (
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/image"
)

type Container struct {
	ID    string
	Image image.Image

	ExposedPorts  map[uint16]uint16
	Env           []string
	RestartPolicy RestartPolicy

	CPU    float64
	Memory int
	Disk   int

	StartedAt  time.Time
	FinishedAt time.Time
}
