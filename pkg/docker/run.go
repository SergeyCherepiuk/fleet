package docker

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/docker/docker/api/types"
	apicontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"golang.org/x/exp/maps"
)

func Run(ctx context.Context, container container.Container) (string, error) {
	if err := pullImage(ctx, container.Image); err != nil {
		return "", err
	}

	id, err := createContainer(ctx, container)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	err = dockerClient.ContainerStart(ctx, id, types.ContainerStartOptions{})
	fmt.Println(err)
	return id, err
}

func pullImage(ctx context.Context, image image.Image) error {
	ref := image.RawRef()
	reader, err := dockerClient.ImagePull(ctx, ref, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	io.Copy(os.Stderr, reader) // TODO: Improve the formatting of the response
	reader.Close()
	return nil
}

func createContainer(ctx context.Context, container container.Container) (string, error) {
	ref := container.Image.RawRef()
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
		Resources: apicontainer.Resources{
			Memory:   container.RequiredResources.Memory,
			NanoCPUs: int64(container.RequiredResources.CPU * math.Pow(10, 9)),
		},
	}
	name := fmt.Sprintf("%s", container.Image.Tag)

	resp, err := dockerClient.ContainerCreate(ctx, &conf, &hconf, nil, nil, name)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
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
			{
				HostIP:   "localhost",
				HostPort: fmt.Sprintf("%d", destination),
			},
		}
		portMap[port] = bindings
	}
	return portMap
}
