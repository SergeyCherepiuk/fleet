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
	"github.com/google/uuid"
)

func (r *Runtime) CreateAndRun(ctx context.Context, container container.Container) (string, error) {
	if err := r.pullImage(ctx, container.Image); err != nil {
		return "", err
	}

	id, err := r.createContainer(ctx, container)
	if err != nil {
		return "", err
	}

	return id, r.Client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (r *Runtime) pullImage(ctx context.Context, image image.Image) error {
	reader, err := r.Client.ImagePull(ctx, image.Ref, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	io.Copy(os.Stderr, reader)
	reader.Close()
	return nil
}

func (r *Runtime) createContainer(ctx context.Context, cont container.Container) (string, error) {
	config := apicontainer.Config{
		Image:        cont.Image.Ref,
		Env:          cont.Config.Env,
		Labels:       cont.Config.Labels,
		ExposedPorts: portSet(cont.Config.ExposedPorts),
	}
	hostConfig := apicontainer.HostConfig{
		PortBindings: portMap(cont.Config.ExposedPorts),
		Resources: apicontainer.Resources{
			Memory:   int64(cont.Config.RequiredResources.Memory),
			NanoCPUs: int64(cont.Config.RequiredResources.CPU * math.Pow(10, 9)),
		},
	}
	name := uuid.NewString()

	resp, err := r.Client.ContainerCreate(ctx, &config, &hostConfig, nil, nil, name)
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

func portMap(exposedPorts []uint16) nat.PortMap {
	portMap := nat.PortMap{}
	for _, port := range exposedPorts {
		port := nat.Port(fmt.Sprintf("%d/tcp", port))
		bindings := []nat.PortBinding{
			{
				HostIP:   "localhost",
				HostPort: fmt.Sprintf("%d", 0),
			},
		}
		portMap[port] = bindings
	}
	return portMap
}
