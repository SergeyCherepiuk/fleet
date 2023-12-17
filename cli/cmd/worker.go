package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SergeyCherepiuk/fleet/pkg/c14n"
	"github.com/SergeyCherepiuk/fleet/pkg/docker"
	"github.com/SergeyCherepiuk/fleet/pkg/manager"
	"github.com/SergeyCherepiuk/fleet/pkg/worker"
	backend "github.com/SergeyCherepiuk/fleet/pkg/worker"
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
	runtime, err := docker.New()
	if err != nil {
		return err
	}

	worker := worker.New(Node, runtime)

	url, err := url.JoinPath(
		"http://",
		workerCmdOptions.managerAddr,
		manager.WorkerEndpoint,
		worker.ID.String(),
	)
	if err != nil {
		return err
	}

	if err := registerWorker(url); err != nil {
		return err
	}
	go startSendingHeartbeats(url)

	addr := fmt.Sprintf("%s:%d", Node.Addr.Addr, Node.Addr.Port)
	return backend.StartServer(addr, worker)
}

func registerWorker(url string) error {
	workerAddr, err := json.Marshal(Node.Addr)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(workerAddr))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return errors.New("worker is up: failed to notify manager")
	}
	return nil
}

func startSendingHeartbeats(url string) {
	client := http.Client{}
	for {
		req, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			continue
		}

		client.Do(req)
		time.Sleep(worker.HeartbeatInterval)
	}
}
