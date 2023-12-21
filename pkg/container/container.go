package container

import (
	"github.com/SergeyCherepiuk/fleet/pkg/image"
)

type RestartPolicy string

const (
	Always        RestartPolicy = "always"
	OnFailure     RestartPolicy = "on-failure"
	UnlessStopped RestartPolicy = "unless-stopped"
	Never         RestartPolicy = "never"
)

type RequiredResources struct {
	CPU    float64
	Memory int64
}

type Container struct {
	ID    string
	Image image.Image

	ExposedPorts      []uint16
	Env               []string
	RestartPolicy     RestartPolicy
	RequiredResources RequiredResources
}
