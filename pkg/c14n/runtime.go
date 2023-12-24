package c14n

import (
	"context"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
)

type Runtime interface {
	Run(ctx context.Context, cont container.Container) (id string, err error)
	Stop(ctx context.Context, container container.Container) error
	Containers(ctx context.Context) ([]container.Container, error)
	ContainerState(ctx context.Context, containerId string) (state string, err error)
}
