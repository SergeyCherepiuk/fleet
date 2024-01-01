package cmd

import (
	"context"

	cmdcontext "github.com/SergeyCherepiuk/fleet/cli/cmd/context"
	"github.com/SergeyCherepiuk/fleet/cli/cmd/manager"
	"github.com/SergeyCherepiuk/fleet/cli/cmd/task"
	"github.com/SergeyCherepiuk/fleet/cli/cmd/worker"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:               "fleet",
		PersistentPreRunE: rootPreRun,
	}
)

func init() {
	RootCmd.AddCommand(manager.ManagerCmd)
	RootCmd.AddCommand(worker.WorkerCmd)
	RootCmd.AddCommand(task.TaskCmd)
}

func rootPreRun(cmd *cobra.Command, _ []string) error {
	ip, err := node.LocalIPv4()
	if err != nil {
		return err
	}

	port, err := node.RandomPort()
	if err != nil {
		return err
	}

	n := node.Node{
		Addr: node.Addr{Addr: ip, Port: port},
	}
	ctx := context.WithValue(cmd.Context(), cmdcontext.NodeKey, n)
	cmd.SetContext(ctx)
	return nil
}
