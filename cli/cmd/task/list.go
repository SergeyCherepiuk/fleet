package task

import "github.com/spf13/cobra"

var ListCmd = &cobra.Command{
	Use:  "list",
	RunE: listRun,
}

func listRun(cmd *cobra.Command, args []string) error {
	return nil
}
