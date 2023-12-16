package cmd

import (
	"runtime"

	"github.com/SergeyCherepiuk/fleet/cli/cmd/task"
	"github.com/SergeyCherepiuk/fleet/internal/memory"
	"github.com/SergeyCherepiuk/fleet/internal/net"
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
	ip, err := net.LocalIPv4()
	if err != nil {
		return err
	}

	port, err := net.RandomPort()
	if err != nil {
		return err
	}

	memory, err := memory.Total()
	if err != nil {
		return err
	}

	Node = node.Node{
		Addr: node.Addr{Addr: ip, Port: port},
		Resources: node.Resources{
			CPU:    float64(runtime.NumCPU()),
			Memory: int64(memory),
		},
	}
	return nil
}
