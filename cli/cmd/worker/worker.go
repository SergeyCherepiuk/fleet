package worker

import (
	"errors"

	"github.com/SergeyCherepiuk/fleet/cli/cmd/context"
	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/docker"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	backend "github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/spf13/cobra"
)

var (
	WorkerCmd = &cobra.Command{
		Use:     "worker",
		PreRunE: workerPreRun,
		RunE:    workerRun,
	}

	workerCmdOptions struct {
		managerAddr string
	}

	workerRuntime c14n.Runtime
)

func init() {
	WorkerCmd.PersistentFlags().StringVar(&workerCmdOptions.managerAddr, "manager", "", "Address and port of the manager node")
	WorkerCmd.AddCommand(ListCmd)
}

func workerPreRun(_ *cobra.Command, _ []string) error {
	if workerCmdOptions.managerAddr == "" {
		return errors.New("manager address is not provided")
	}

	var err error
	workerRuntime, err = docker.New()
	return err
}

func workerRun(cmd *cobra.Command, _ []string) error {
	n := cmd.Context().Value(context.NodeKey).(node.Node)
	worker := backend.New(n, workerRuntime, workerCmdOptions.managerAddr)
	return backend.StartServer(n.Addr.String(), worker)
}
