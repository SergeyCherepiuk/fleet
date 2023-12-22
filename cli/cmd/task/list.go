package task

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const Template = `
{{range $worker_id, $tasks := .}}worker {{$worker_id}}:{{range $tasks}}
{{.ID}} (state: {{.State}}, restarts: {{len .Restarts}}){{end}}
{{end}}
`

var ListCmd = &cobra.Command{
	Use:  "list",
	RunE: listRun,
}

func listRun(_ *cobra.Command, _ []string) error {
	resp, err := httpclient.Get(taskCmdOptions.managerAddr, "/task/list")
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

	text := strings.TrimSpace(Template)
	tmpl := template.Must(template.New("list").Parse(text))
	return tmpl.Execute(os.Stdout, stat)
}
