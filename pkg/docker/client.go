package docker

import (
	"log"

	"github.com/docker/docker/client"
)

var dockerClient *client.Client

func init() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
}
