package cmd

import (
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	backend "github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/spf13/cobra"
)

var ManagerCmd = &cobra.Command{
	Use:  "manager",
	RunE: managerRun,
}

func managerRun(_ *cobra.Command, _ []string) error {
	scheduler := scheduler.NewEpvm(scheduler.EpvmStrategyBestFit)
	manager := manager.New(Node, scheduler)
	return backend.StartServer(Node.Addr.String(), manager)
}
