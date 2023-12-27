package task

import (
	"errors"
	"fmt"
	"net/http"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:  "stop",
	RunE: stopRun,
}

func stopRun(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("no task id provided")
	}

	id, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/task/stop/%s", id)
	resp, err := httpclient.Post(taskCmdOptions.managerAddr, endpoint, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		message := httpinternal.ErrorMessage(resp.Body)
		return errors.New(message)
	}

	return nil
}
