package task

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:  "list",
	RunE: listRun,
}

func listRun(cmd *cobra.Command, args []string) error {
	url, err := url.JoinPath("http://", taskCmdOptions.managerAddr, manager.StatEndpoint)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var stat map[uuid.UUID][]task.Task
	if err := json.Unmarshal(body, &stat); err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, format(stat))
	return nil
}

func format(stat map[uuid.UUID][]task.Task) (s string) {
	for wid, tasks := range stat {
		s += fmt.Sprintf("worker %s:\n", wid)
		for i, task := range tasks {
			s += fmt.Sprintf("%d: %s (%s)\n", i+1, task.ID, task.State)
		}
	}
	return
}
