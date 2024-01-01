package task

import (
	"fmt"
	"net/url"
	"time"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/format"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/spf13/cobra"
)

var (
	ListCmd = &cobra.Command{
		Use:  "list",
		RunE: listRun,
	}

	listCmdOptions struct {
		workerId string
	}
)

func init() {
	ListCmd.Flags().StringVarP(&listCmdOptions.workerId, "worker", "w", "", "Worker ID to list the tasks of a specific worker")
}

func listRun(_ *cobra.Command, _ []string) error {
	endpoint, _ := url.JoinPath("/task/list", listCmdOptions.workerId)
	resp, err := httpclient.Get(taskCmdOptions.managerAddr, endpoint)
	if err != nil {
		return err
	}

	var tasks []task.Task
	if err := httpinternal.Body(resp, &tasks); err != nil {
		return err
	}

	headers := []string{"TASK ID", "IMAGE", "STATE", "RESTARTS", "START TIME", "FINISH TIME"}
	accessMap := format.AccessMap[task.Task]{
		"TASK ID":     func(t task.Task) any { return t.Id },
		"IMAGE":       func(t task.Task) any { return t.Container.Image.Ref },
		"STATE":       func(t task.Task) any { return t.State },
		"RESTARTS":    func(t task.Task) any { return len(t.Restarts) },
		"START TIME":  func(t task.Task) any { return formatTime(t.StartedAt) },
		"FINISH TIME": func(t task.Task) any { return formatTime(t.FinishedAt) },
	}
	fmt.Print(format.Table[task.Task](headers, accessMap, tasks))
	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.DateTime)
}
