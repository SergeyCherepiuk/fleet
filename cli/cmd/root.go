package cmd

import (
	"github.com/SergeyCherepiuk/fleet/cli/cmd/task"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:               "fleet",
		PersistentPreRunE: rootPreRun,
	}
	Node node.Node
)

func init() {
	RootCmd.AddCommand(ManagerCmd)
	RootCmd.AddCommand(WorkerCmd)
	RootCmd.AddCommand(task.TaskCmd)
}

func rootPreRun(_ *cobra.Command, _ []string) error {
	ip, err := Node.LocalIPv4()
	if err != nil {
		return err
	}

	port, err := Node.RandomPort()
	if err != nil {
		return err
	}

	Node = node.Node{
		Addr: node.Addr{Addr: ip, Port: port},
	}
	return nil
}
