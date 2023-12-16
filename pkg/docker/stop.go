package docker

import (
	"context"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/docker/docker/api/types"
	apicontainer "github.com/docker/docker/api/types/container"
)

func (r *Runtime) Stop(ctx context.Context, container container.Container) error {
	err := r.Client.ContainerStop(ctx, container.ID, apicontainer.StopOptions{})
	if err != nil {
		return err
	}

	removeOpts := types.ContainerRemoveOptions{RemoveVolumes: true}
	return r.Client.ContainerRemove(ctx, container.ID, removeOpts)
}
