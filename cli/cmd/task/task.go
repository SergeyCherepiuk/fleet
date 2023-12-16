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
		Use:     "task",
		PreRunE: taskPreRun,
	}
	taskCmdOptions TaskCmdOptions
)

func init() {
	TaskCmd.PersistentFlags().StringVar(&taskCmdOptions.managerAddr, "manager", "", "Address and port of the manager node")
	TaskCmd.AddCommand(RunCmd)
	TaskCmd.AddCommand(StopCmd)
	TaskCmd.AddCommand(ListCmd)
}

func taskPreRun(_ *cobra.Command, _ []string) error {
	if taskCmdOptions.managerAddr == "" {
		return errors.New("manager address is not provided")
	}
	return nil
}
