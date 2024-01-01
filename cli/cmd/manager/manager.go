package manager

import (
	"github.com/SergeyCherepiuk/fleet/cli/cmd/context"
	backend "github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/spf13/cobra"
)

var ManagerCmd = &cobra.Command{
	Use:  "manager",
	RunE: managerRun,
}

func managerRun(cmd *cobra.Command, _ []string) error {
	n := cmd.Context().Value(context.NodeKey).(node.Node)
	scheduler := scheduler.NewEpvm(scheduler.EpvmStrategyBestFit)
	manager := backend.New(n, scheduler)
	return backend.StartServer(n.Addr.String(), manager)
}
