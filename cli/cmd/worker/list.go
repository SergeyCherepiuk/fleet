package worker

import (
	"fmt"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/format"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:  "list",
	RunE: listRun,
}

func listRun(_ *cobra.Command, _ []string) error {
	resp, err := httpclient.Get(workerCmdOptions.managerAddr, "/worker/list")
	if err != nil {
		return err
	}

	var workers []worker.Info
	if err := httpinternal.Body(resp, &workers); err != nil {
		return err
	}

	headers := []string{"WORKER ID", "IP ADDRESS", "MANAGER IP", "TASKS COUNT", "RUNTIME"}
	accessMap := format.AccessMap[worker.Info]{
		"WORKER ID":   func(i worker.Info) any { return i.Id },
		"IP ADDRESS":  func(i worker.Info) any { return i.Addr },
		"MANAGER IP":  func(i worker.Info) any { return i.ManagerAddr },
		"TASKS COUNT": func(i worker.Info) any { return i.TasksCount },
		"RUNTIME":     func(i worker.Info) any { return i.RuntimeName },
	}
	fmt.Print(format.Table[worker.Info](headers, accessMap, workers))
	return nil
}
