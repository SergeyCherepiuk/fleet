package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:  "run",
	RunE: runRun,
}

func runRun(_ *cobra.Command, _ []string) error {
	t := task.Task{
		ID:    uuid.New(),
		State: task.Pending,
		Container: container.Container{
			Image: image.Image{
				Registy: "docker.io/library",
				Tag:     "nginx",
				Version: "alpine",
			},
			ExposedPorts:  []uint16{80},
			Env:           []string{"foo=bar"},
			RestartPolicy: container.OnFailure,
			RequiredResources: container.RequiredResources{
				CPU:    4.0,
				Memory: 150 * 1024 * 1024,
			},
		},
	}

	marshaledTask, err := json.Marshal(t)
	if err != nil {
		return err
	}

	url, err := url.JoinPath("http://", taskCmdOptions.managerAddr, manager.TaskRunEndpoint)
	if err != nil {
		return err
	}
	body := bytes.NewReader(marshaledTask)

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		message, err := errorMessage(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(message)
	}

	return nil
}

func errorMessage(body io.ReadCloser) (string, error) {
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	var b struct{ Message string }
	if err := json.Unmarshal(data, &b); err != nil {
		return "", err
	}

	return b.Message, nil
}
