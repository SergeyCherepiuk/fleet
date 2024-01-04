package task

import (
	"errors"
	"net/http"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/parse"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:  "run",
	RunE: runRun,
}

func runRun(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("no manifest file provided")
	}

	tasks, err := parse.Parse(args[0])
	if err != nil {
		return err
	}

	resp, err := httpclient.Post(taskCmdOptions.managerAddr, "/task/run", tasks)

	if resp.StatusCode != http.StatusCreated {
		message := httpinternal.ErrorMessage(resp.Body)
		return errors.New(message)
	}

	return nil
}
