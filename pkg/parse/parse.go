package parse

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/image"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"gopkg.in/yaml.v3"
)

type ManifestEntry struct {
	Image             string
	Env               map[string]string
	Labels            container.Labels
	ExposedPorts      []uint16                    `yaml:"exposedPorts"`
	RestartPolicy     container.RestartPolicy     `yaml:"restartPolicy"`
	RequiredResources container.RequiredResources `yaml:"requiredResources"`
}

func (me *ManifestEntry) validate() error {
	if strings.TrimSpace(me.Image) == "" {
		return errors.New("image is not provided for one of the tasks")
	}

	knownRestartPolicy := me.RestartPolicy == "never" ||
		me.RestartPolicy == "on-failure" ||
		me.RestartPolicy == "always"

	if me.RestartPolicy == "" {
		me.RestartPolicy = container.Never
	} else if !knownRestartPolicy {
		return fmt.Errorf(
			"unknown restart policy, available options: %q, %q, %q",
			container.Never, container.OnFailure, container.Always,
		)
	}

	if me.Labels == nil {
		me.Labels = make(container.Labels)
	}

	if me.ExposedPorts == nil {
		me.ExposedPorts = make([]uint16, 0)
	}

	return nil
}

func (me *ManifestEntry) toTask() task.Task {
	image := image.Image{Ref: me.Image}
	container := container.New(image, container.Config{
		ExposedPorts:      me.ExposedPorts,
		Env:               joinEnvs(me.Env),
		Labels:            me.Labels,
		RestartPolicy:     me.RestartPolicy,
		RequiredResources: me.RequiredResources,
	})
	return *task.New(*container)
}

func joinEnvs(env map[string]string) []string {
	joined := make([]string, 0, len(env))
	for k, v := range env {
		joined = append(joined, fmt.Sprintf("%s=%s", k, v))
	}
	return joined
}

func Parse(filepath string) ([]task.Task, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(content)
	d := yaml.NewDecoder(r)
	d.KnownFields(true)

	var entries []struct{ Task ManifestEntry }
	if err := d.Decode(&entries); err != nil {
		return nil, err
	}

	tasks := make([]task.Task, 0, len(entries))
	for _, entry := range entries {
		if err := entry.Task.validate(); err != nil {
			return nil, err
		}

		tasks = append(tasks, entry.Task.toTask())
	}
	return tasks, nil
}
