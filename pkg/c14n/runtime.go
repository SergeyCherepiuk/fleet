package c14n

import (
	"context"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
)

type Runtime interface {
	CreateAndRun(context.Context, container.Container) (id string, err error)
	StopAndRemove(ctx context.Context, id string) error
	Containers(context.Context) ([]container.Container, error)
	ContainerState(ctx context.Context, id string) (container.State, error)
}
