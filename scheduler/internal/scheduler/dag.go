package scheduler

import "github.com/Tanapon-gau/miniflow/scheduler/internal/model"

// DepsFromDAG builds a name→[]upstream map from parsed DAG task definitions.
func DepsFromDAG(defs []model.DAGTask) map[string][]string {
	m := make(map[string][]string, len(defs))
	for _, d := range defs {
		m[d.Name] = d.DependsOn
	}
	return m
}

// ReadyTasks returns tasks whose status is "pending" and all upstream deps are "success".
func ReadyTasks(tasks []model.Task, deps map[string][]string) []model.Task {
	statusByName := make(map[string]string, len(tasks))
	for _, t := range tasks {
		statusByName[t.Name] = t.Status
	}
	var ready []model.Task
	for _, t := range tasks {
		if t.Status != "pending" {
			continue
		}
		ok := true
		for _, up := range deps[t.Name] {
			if statusByName[up] != "success" {
				ok = false
				break
			}
		}
		if ok {
			ready = append(ready, t)
		}
	}
	return ready
}

// RunOutcome returns "success", "failed", or "" (still in progress).
func RunOutcome(tasks []model.Task) string {
	for _, t := range tasks {
		switch t.Status {
		case "failed":
			return "failed"
		case "pending", "queued", "running", "retrying":
			return ""
		}
	}
	return "success"
}
