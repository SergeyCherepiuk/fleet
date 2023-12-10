package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/docker/docker/api/types"
	apicontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"golang.org/x/exp/maps"
)

func Run(container container.Container) error {
	ctx := context.Background()

	// Pulling image
	ref := container.Image.RawRef()
	reader, err := dockerClient.ImagePull(ctx, ref, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	io.Copy(os.Stderr, reader) // TODO: Improve the formatting of the response
	reader.Close()

	// Creating container
	conf := apicontainer.Config{
		Image:        ref,
		Env:          container.Env,
		ExposedPorts: portSet(maps.Keys(container.ExposedPorts)),
	}
	hconf := apicontainer.HostConfig{
		PortBindings: portMap(container.ExposedPorts),
		RestartPolicy: apicontainer.RestartPolicy{
			Name: string(container.RestartPolicy),
		},
	}
	name := fmt.Sprintf("%s-%s", container.Image.Tag, container.ID)

	resp, err := dockerClient.ContainerCreate(ctx, &conf, &hconf, nil, nil, name)
	if err != nil {
		return err
	}

	// Running container
	return dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

func portSet(ports []uint16) nat.PortSet {
	portSet := nat.PortSet{}
	for _, p := range ports {
		port := nat.Port(fmt.Sprintf("%d/tcp", p))
		portSet[port] = struct{}{}
	}
	return portSet
}

func portMap(exposedPorts map[uint16]uint16) nat.PortMap {
	portMap := nat.PortMap{}
	for source, destination := range exposedPorts {
		port := nat.Port(fmt.Sprintf("%d/tcp", source))
		bindings := []nat.PortBinding{
			nat.PortBinding{
				HostIP:   "localhost",
				HostPort: fmt.Sprintf("%d", destination),
			},
		}
		portMap[port] = bindings
	}
	return portMap
}
