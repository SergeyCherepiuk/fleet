package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	apicontainer "github.com/docker/docker/api/types/container"
)

func (r *Runtime) StopAndRemove(ctx context.Context, id string) error {
	r.Client.ContainerStop(ctx, id, apicontainer.StopOptions{})
	removeOpts := types.ContainerRemoveOptions{RemoveVolumes: true}
	return r.Client.ContainerRemove(ctx, id, removeOpts)
}
