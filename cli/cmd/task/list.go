package task

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/format"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
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
		"TASK ID":     func(t task.Task) any { return formatId(t.Id) },
		"IMAGE":       func(t task.Task) any { return trimImageRef(t.Container.Image.Ref) },
		"STATE":       func(t task.Task) any { return t.State },
		"RESTARTS":    func(t task.Task) any { return max(0, len(t.StartedAt)-1) },
		"START TIME":  func(t task.Task) any { return formatLastTime(t.StartedAt) },
		"FINISH TIME": func(t task.Task) any { return formatLastTime(t.FinishedAt) },
	}
	fmt.Print(format.Table[task.Task](headers, accessMap, tasks))
	return nil
}

func formatId(id uuid.UUID) string {
	if id == uuid.Nil {
		return "-"
	}
	return id.String()
}

func trimImageRef(ref string) string {
	index := strings.LastIndexByte(ref, '/')
	if index == -1 || index == len(ref)-1 {
		return ref
	}
	return ref[index+1:]
}

func formatLastTime(ts []time.Time) string {
	if len(ts) == 0 {
		return "-"
	}
	return ts[len(ts)-1].Format(time.DateTime)
}
