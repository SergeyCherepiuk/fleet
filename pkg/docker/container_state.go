package docker

import (
	"context"
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
)

func (r *Runtime) ContainerState(ctx context.Context, id string) (container.State, error) {
	json, err := r.Client.ContainerInspect(ctx, id)
	if err != nil {
		return container.State{}, err
	}

	if json.State == nil {
		return container.State{}, errors.New("container's state is nil")
	}

	state := container.State{
		Status:   json.State.Status,
		ExitCode: json.State.ExitCode,
	}
	return state, nil
}
