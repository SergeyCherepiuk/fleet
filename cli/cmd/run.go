package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type RunCmdOptions struct {
	managerAddr string
}

var (
	RunCmd = &cobra.Command{
		Use:     "run",
		PreRunE: runPreRunE,
		RunE:    runRunE,
	}
	runCmdOptions RunCmdOptions
)

func init() {
	RunCmd.Flags().StringVar(&runCmdOptions.managerAddr, "manager", "", "Address and port of the manager node")
}

func runPreRunE(cmd *cobra.Command, args []string) error {
	if runCmdOptions.managerAddr == "" {
		return errors.New("manager address is not provided")
	}

	return nil
}

func runRunE(cmd *cobra.Command, args []string) error {
	t := task.Task{
		ID:    uuid.New(),
		State: task.Pending,
		Container: container.Container{
			Image: image.Image{
				Registy: "docker.io/library",
				Tag:     "nginx",
				Version: "alpine",
			},
			ExposedPorts:  map[uint16]uint16{80: 80},
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

	url, err := url.JoinPath("http://", runCmdOptions.managerAddr, manager.TaskRunEndpoint)
	if err != nil {
		return err
	}
	body := bytes.NewReader(marshaledTask)

	_, err = http.Post(url, "application/json", body)
	return err
}
