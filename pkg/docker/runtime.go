package docker

import (
	"github.com/docker/docker/client"
)

type Runtime struct {
	Client *client.Client
}

func New() (*Runtime, error) {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Runtime{Client: client}, nil
}

func (r *Runtime) Name() string {
	return "docker"
}
