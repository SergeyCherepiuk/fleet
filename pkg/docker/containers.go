package docker

import (
	"context"
	"fmt"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func (r *Runtime) Containers(ctx context.Context) ([]container.Container, error) {
	dockerContainers, err := r.Client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: labelFilter(),
	})
	if err != nil {
		return nil, err
	}

	fleetConts := make([]container.Container, len(dockerContainers))
	for i, dockerContainer := range dockerContainers {
		fleetConts[i] = toFleetContainer(dockerContainer)
	}
	return fleetConts, nil
}

func labelFilter() filters.Args {
	label := fmt.Sprintf("%s=%s", container.TypeLabelKey, container.TypeLabelValue)
	return filters.NewArgs(filters.Arg("label", label))
}

func toFleetContainer(dockerContainer types.Container) container.Container {
	return container.Container{
		ID: dockerContainer.ID,
		Image: image.Image{
			ID:  dockerContainer.ImageID,
			Ref: dockerContainer.Image,
		},
		Config: container.Config{
			ExposedPorts: privatePorts(dockerContainer.Ports),
			Labels:       dockerContainer.Labels,
			// NOTE(SergeyCherepiuk): Didn't found the way to get envs
			// NOTE(SergeyCherepiuk): Restart policy and required resources are ignored
		},
	}
}

func privatePorts(ports []types.Port) []uint16 {
	private := make([]uint16, len(ports))
	for i, port := range ports {
		private[i] = port.PrivatePort
	}
	return private
}
