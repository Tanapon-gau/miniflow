package scheduler

import (
	"github.com/Tanapon-gau/miniflow/scheduler/internal/constants"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/model"
)

func DepsFromDAG(defs []model.DAGTask) map[string][]string {
	deps := make(map[string][]string, len(defs))
	for _, definition := range defs {
		deps[definition.Name] = definition.DependsOn
	}
	return deps
}

func ReadyTasks(tasks []model.Task, deps map[string][]string) []model.Task {
	statusByName := make(map[string]string, len(tasks))
	for _, task := range tasks {
		statusByName[task.Name] = task.Status
	}
	var ready []model.Task
	for _, task := range tasks {
		if task.Status != constants.StatusPending {
			continue
		}
		allDepsSucceeded := true
		for _, upstreamName := range deps[task.Name] {
			if statusByName[upstreamName] != constants.StatusSuccess {
				allDepsSucceeded = false
				break
			}
		}
		if allDepsSucceeded {
			ready = append(ready, task)
		}
	}
	return ready
}

func RunOutcome(tasks []model.Task) string {
	for _, task := range tasks {
		switch task.Status {
		case constants.StatusFailed:
			return constants.StatusFailed
		case constants.StatusPending, constants.StatusQueued, constants.StatusRunning, constants.StatusRetrying:
			return ""
		}
	}
	return constants.StatusSuccess
}
