package cmd

import (
	"fmt"

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
	manager := manager.New(Node, scheduler.AlwaysFirst{})
	addr := fmt.Sprintf("%s:%d", Node.Addr.Addr, Node.Addr.Port)
	return backend.StartServer(addr, manager)
}
