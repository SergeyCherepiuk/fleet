package docker

import (
	"context"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/docker/docker/api/types"
	apicontainer "github.com/docker/docker/api/types/container"
)

func Stop(ctx context.Context, container container.Container) error {
	err := dockerClient.ContainerStop(ctx, container.ID, apicontainer.StopOptions{})
	if err != nil {
		return err
	}

	return dockerClient.ContainerRemove(
		ctx, container.ID, types.ContainerRemoveOptions{RemoveVolumes: true},
	)
}
