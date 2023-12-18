package task

import (
	"net/http"
	"net/url"

	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:  "stop",
	RunE: stopRun,
}

func stopRun(_ *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	url, err := url.JoinPath(
		"http://",
		taskCmdOptions.managerAddr,
		manager.TaskStopEndpoint,
		id.String(),
	)
	if err != nil {
		return err
	}

	// TODO(SergeyCherepiuk): Process response
	_, err = http.Post(url, "", nil)
	return err
}
