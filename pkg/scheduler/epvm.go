package scheduler

import (
	"math"

	httpinternal "github.com/SergeyCherepiuk/fleet/internal/http"
	mapsinternal "github.com/SergeyCherepiuk/fleet/internal/maps"
	"github.com/SergeyCherepiuk/fleet/pkg/consensus"
	"github.com/SergeyCherepiuk/fleet/pkg/container"
	"github.com/SergeyCherepiuk/fleet/pkg/httpclient"
	"github.com/SergeyCherepiuk/fleet/pkg/node"
	"github.com/SergeyCherepiuk/fleet/pkg/task"
	"github.com/google/uuid"
)

const Leib = 1.539600717839002

type epvm struct {
	strategy EpvmStrategy
}

func NewEpvm(stragery EpvmStrategy) *epvm {
	return &epvm{strategy: stragery}
}

type EpvmStrategy string

const (
	EpvmStrategyBestFit  EpvmStrategy = "EpvmStrategyBestFit"
	EpvmStrategyWorstFit EpvmStrategy = "EpvmStrategyWorstFit"
)

func (e *epvm) SelectWorker(t task.Task, ws map[uuid.UUID]consensus.Worker) (uuid.UUID, consensus.Worker, error) {
	if len(ws) == 0 {
		return uuid.Nil, consensus.Worker{}, ErrNoAvailableWorkers
	}

	workersResources := queryResources(ws)
	dropUnableWorkers(t.Container.Config.RequiredResources, workersResources)

	if len(workersResources) == 0 {
		return uuid.Nil, consensus.Worker{}, ErrNoCapableWorkers
	}

	costs := costs(t.Container.Config.RequiredResources, workersResources)
	id := pick(costs, e.strategy)
	return id, ws[id], nil
}

func queryResources(workers map[uuid.UUID]consensus.Worker) map[uuid.UUID]node.Resources {
	workersResources := make(map[uuid.UUID]node.Resources)
	for id, worker := range workers {
		resp, err := httpclient.Get(worker.Addr.String(), "/resources/available")
		if err != nil {
			continue
		}

		var resources node.Resources
		if err := httpinternal.Body(resp, &resources); err != nil {
			continue
		}

		workersResources[id] = resources
	}
	return workersResources
}

func dropUnableWorkers(
	requiredResources container.RequiredResources,
	workersResources map[uuid.UUID]node.Resources,
) {
	for id, resoures := range workersResources {
		notEnoughMem := resoures.Memory.Available < requiredResources.Memory
		notEnoughCpu := float64(resoures.CPU.Cores) < requiredResources.CPU
		if notEnoughMem || notEnoughCpu {
			delete(workersResources, id)
		}
	}
}

func costs(
	requiredResources container.RequiredResources,
	workersResources map[uuid.UUID]node.Resources,
) map[uuid.UUID]float64 {
	costs := make(map[uuid.UUID]float64)
	for id, resources := range workersResources {
		costs[id] = cost(requiredResources, resources)
	}
	return costs
}

func cost(
	requiredResources container.RequiredResources,
	workerResources node.Resources,
) float64 {
	memoryBefore := workerResources.Memory.Available
	memoryAfter := workerResources.Memory.Available - requiredResources.Memory
	memoryCost := math.Pow(Leib, float64(memoryBefore)) - math.Pow(Leib, float64(memoryAfter))
	return memoryCost
}

func pick(costs map[uuid.UUID]float64, strategy EpvmStrategy) uuid.UUID {
	switch strategy {
	case EpvmStrategyBestFit:
		return mapsinternal.KeyWithMaxValue(costs)
	case EpvmStrategyWorstFit:
		return mapsinternal.KeyWithMinValue(costs)
	}
	return uuid.Nil
}
