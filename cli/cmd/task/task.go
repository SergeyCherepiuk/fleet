package task

import (
	"errors"

	"github.com/spf13/cobra"
)

type TaskCmdOptions struct {
	managerAddr string
}

var (
	TaskCmd = &cobra.Command{
		Use:               "task",
		PersistentPreRunE: taskPreRun,
	}
	taskCmdOptions TaskCmdOptions
)

func init() {
	TaskCmd.PersistentFlags().StringVar(&taskCmdOptions.managerAddr, "manager", "", "Address and port of the manager node")
	TaskCmd.AddCommand(RunCmd)
	TaskCmd.AddCommand(StopCmd)
}

func taskPreRun(_ *cobra.Command, _ []string) error {
	if taskCmdOptions.managerAddr == "" {
		return errors.New("manager address is not provided")
	}
	return nil
}
