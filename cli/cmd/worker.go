package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/docker"
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	backend "github.com/SergeyCherepiuk/fleet/pkg/worker"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type WorkerCmdOptions struct {
	managerAddr string
}

var (
	WorkerCmd = &cobra.Command{
		Use:     "worker",
		PreRunE: workerPreRun,
		RunE:    workerRun,
	}
	workerCmdOptions WorkerCmdOptions
	workerRuntime    c14n.Runtime
)

func init() {
	WorkerCmd.PersistentFlags().StringVar(&workerCmdOptions.managerAddr, "manager", "", "Address and port of the manager node")
}

func workerPreRun(_ *cobra.Command, _ []string) error {
	if workerCmdOptions.managerAddr == "" {
		return errors.New("manager address is not provided")
	}

	var err error
	workerRuntime, err = docker.New()
	return err
}

func workerRun(_ *cobra.Command, _ []string) error {
	worker := worker.Worker{
		Node:    Node,
		Runtime: workerRuntime,
		ID:      uuid.New(),
		Tasks:   make(map[uuid.UUID]task.Task),
	}

	url, err := url.JoinPath("http://", workerCmdOptions.managerAddr, manager.WorkerEndpoint)
	if err != nil {
		return err
	}

	workerAddr, err := json.Marshal(Node.Addr)
	if err != nil {
		return err
	}

	if err := notifyManagerWorkerUp(url, workerAddr); err != nil {
		return err
	}
	go notifyManagerWorkerDown(url, workerAddr) // NOTE(SergeyCherepiuk): Works as defer

	addr := fmt.Sprintf("%s:%d", Node.Addr.Addr, Node.Addr.Port)
	return backend.StartServer(addr, &worker)
}

func notifyManagerWorkerUp(url string, body []byte) error {
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return errors.New("worker is up: failed to notify manager")
	}
	return nil
}

func notifyManagerWorkerDown(url string, body []byte) error {
	sigint := make(chan os.Signal)
	signal.Notify(sigint, syscall.SIGINT)
	<-sigint
	defer os.Exit(0)

	req, err := http.NewRequest("DELETE", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return errors.New("worker is down: failed to notify manager")
	}

	return nil
}
