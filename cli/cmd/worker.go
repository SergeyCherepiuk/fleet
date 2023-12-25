package cmd

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/docker"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	backend "github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/spf13/cobra"
)

type WorkerCmdOptions struct {
	managerAddr string
}

var (
	WorkerCmd = &cobra.Command{
		Use:     "worker",
		PreRunE: workerPreRun,
		RunE:    workerRun,
	}
	workerCmdOptions WorkerCmdOptions
	workerRuntime    c14n.Runtime
)

func init() {
	WorkerCmd.PersistentFlags().StringVar(&workerCmdOptions.managerAddr, "manager", "", "Address and port of the manager node")
}

func workerPreRun(_ *cobra.Command, _ []string) error {
	if workerCmdOptions.managerAddr == "" {
		return errors.New("manager address is not provided")
	}

	var err error
	workerRuntime, err = docker.New()
	return err
}

func workerRun(_ *cobra.Command, _ []string) error {
	worker := worker.New(Node, workerRuntime, workerCmdOptions.managerAddr)
	return backend.StartServer(Node.Addr.String(), worker)
}
