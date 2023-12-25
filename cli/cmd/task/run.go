package task

import (
	"errors"
	"net/http"

	httpinternal "github.com/SergeyCherepiuk/fleet/cli/cmd/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:  "run",
	RunE: runRun,
}

func runRun(_ *cobra.Command, _ []string) error {
	i := image.Image{
		Ref: "docker.io/library/nginx:alpine",
	}

	cc := container.Config{
		ExposedPorts:  []uint16{80},
		Env:           make([]string, 0),
		Labels:        make(container.Labels),
		RestartPolicy: container.OnFailure,
		RequiredResources: container.RequiredResources{
			CPU:    4.0,
			Memory: 150 * 1024 * 1024,
		},
	}

	c := container.New(i, cc)

	t := task.New(*c)

	resp, err := httpclient.Post(taskCmdOptions.managerAddr, "/task/run", t)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		message := httpinternal.ErrorMessage(resp.Body)
		return errors.New(message)
	}

	return nil
}
