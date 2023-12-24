package docker

import (
	"context"
	"errors"
)

func (r *Runtime) ContainerState(ctx context.Context, containerId string) (string, error) {
	json, err := r.Client.ContainerInspect(ctx, containerId)
	if err != nil {
		return "", err
	}

	if json.State == nil {
		return "", errors.New("container's state is nil")
	}

	return json.State.Status, nil
}
