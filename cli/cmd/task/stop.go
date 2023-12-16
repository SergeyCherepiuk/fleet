package task

import "github.com/spf13/cobra"

var StopCmd = &cobra.Command{
	Use:  "stop",
	RunE: stopRun,
}

func stopRun(cmd *cobra.Command, args []string) error {
	return nil
}
