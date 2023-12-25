package container

import (
	"github.com/SergeyCherepiuk/fleet/pkg/image"
)

type Labels map[string]string

func (l *Labels) With(other Labels) Labels {
	for k, v := range other {
		(*l)[k] = v
	}
	return *l
}

const (
	TypeLabelKey   = "com.fleet.key"
	TypeLabelValue = "container"
)

var DefaultLabels = Labels{
	TypeLabelKey: TypeLabelValue,
}

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

// TODO(SergeyCherepiuk): Add name field (manager's name + uuid)
type Config struct {
	ExposedPorts      []uint16
	Env               []string
	Labels            Labels
	RestartPolicy     RestartPolicy
	RequiredResources RequiredResources
}

type Container struct {
	ID     string
	Image  image.Image
	Config Config
}

func New(image image.Image, config Config) *Container {
	return &Container{
		Image: image,
		Config: Config{
			ExposedPorts:      config.ExposedPorts,
			Env:               config.Env,
			Labels:            config.Labels.With(DefaultLabels),
			RestartPolicy:     config.RestartPolicy,
			RequiredResources: config.RequiredResources,
		},
	}
}
