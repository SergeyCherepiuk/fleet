package cmd

import (
	"fmt"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	backend "github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/scheduler"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var ManagerCmd = &cobra.Command{
	Use:     "manager",
	PreRunE: RootCmd.RunE,
	RunE:    managerRunE,
}

func managerRunE(cmd *cobra.Command, args []string) error {
	manager := manager.Manager{
		Node:         Node,
		ID:           uuid.New(),
		Scheduler:    scheduler.AlwaysFirst{},
		WorkerNodes:  make([]node.Addr, 0),
		PendingTasks: queue.New[task.Task](0),
	}

	addr := fmt.Sprintf("%s:%d", Node.Addr.Addr, Node.Addr.Port)

	return backend.StartServer(addr, manager)
}
