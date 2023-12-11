package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/SergeyCherepiuk/fleet/pkg/collections/queue"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	backend "github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type ErrManagerAddrMissing error

var (
	WorkerCmd = &cobra.Command{
		Use:     "worker",
		PreRunE: workerPreRunE,
		RunE:    workerRunE,
	}
	managerAddr string
)

func init() {
	WorkerCmd.Flags().StringVar(&managerAddr, "manager", "", "Address and port of the manager node")
}

func workerPreRunE(cmd *cobra.Command, args []string) error {
	if err := RootCmd.RunE(cmd, args); err != nil {
		return err
	}

	if managerAddr == "" {
		return ErrManagerAddrMissing(errors.New("manager address is not provided"))
	}

	return nil
}

func workerRunE(cmd *cobra.Command, args []string) error {
	worker := worker.Worker{
		Node:  Node,
		ID:    uuid.New(),
		Tasks: make(map[uuid.UUID]task.Task),
		Queue: queue.New[task.Task](0),
	}

	addr := fmt.Sprintf("%s:%d", Node.Addr.Addr, Node.Addr.Port)

	if err := notifyManager(); err != nil {
		return err
	}

	return backend.StartServer(addr, worker)
}

func notifyManager() error {
	workerAddr, err := json.Marshal(Node.Addr)
	if err != nil {
		return err
	}

	// TODO(SergeyCherepiuk): Process response
	url, err := url.JoinPath("http://"+managerAddr, "worker")
	if err != nil {
		return err
	}
	body := bytes.NewReader(workerAddr)

	_, err = http.Post(url, "application/json", body)
	return err
}
