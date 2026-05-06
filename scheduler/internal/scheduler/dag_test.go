package scheduler_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/constants"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/model"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/scheduler"
)

func task(name, taskStatus string) model.Task {
	return model.Task{ID: uuid.New(), Name: name, Status: taskStatus}
}

func noDeps(names ...string) map[string][]string {
	m := make(map[string][]string)
	for _, name := range names {
		m[name] = nil
	}
	return m
}

func TestReadyTasks_NoDeps_AllPending(t *testing.T) {
	tasks := []model.Task{task("a", constants.StatusPending), task("b", constants.StatusPending)}
	ready := scheduler.ReadyTasks(tasks, noDeps("a", "b"))
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready, got %d", len(ready))
	}
}

func TestReadyTasks_BlockedByRunningDep(t *testing.T) {
	tasks := []model.Task{task("a", constants.StatusRunning), task("b", constants.StatusPending)}
	deps := map[string][]string{"a": nil, "b": {"a"}}
	ready := scheduler.ReadyTasks(tasks, deps)
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready, got %d", len(ready))
	}
}

func TestReadyTasks_UnblockedAfterDepSucceeds(t *testing.T) {
	tasks := []model.Task{task("a", constants.StatusSuccess), task("b", constants.StatusPending)}
	deps := map[string][]string{"a": nil, "b": {"a"}}
	ready := scheduler.ReadyTasks(tasks, deps)
	if len(ready) != 1 || ready[0].Name != "b" {
		t.Fatalf("expected b to be ready, got %v", ready)
	}
}

func TestReadyTasks_SkipsNonPending(t *testing.T) {
	tasks := []model.Task{
		task("a", constants.StatusQueued),
		task("b", constants.StatusRunning),
		task("c", constants.StatusSuccess),
	}
	ready := scheduler.ReadyTasks(tasks, noDeps("a", "b", "c"))
	if len(ready) != 0 {
		t.Fatalf("expected 0, got %d", len(ready))
	}
}

func TestRunOutcome_AllSuccess(t *testing.T) {
	tasks := []model.Task{task("a", constants.StatusSuccess), task("b", constants.StatusSuccess)}
	if got := scheduler.RunOutcome(tasks); got != constants.StatusSuccess {
		t.Fatalf("expected %q, got %q", constants.StatusSuccess, got)
	}
}

func TestRunOutcome_OneFailed(t *testing.T) {
	tasks := []model.Task{task("a", constants.StatusSuccess), task("b", constants.StatusFailed)}
	if got := scheduler.RunOutcome(tasks); got != constants.StatusFailed {
		t.Fatalf("expected %q, got %q", constants.StatusFailed, got)
	}
}

func TestRunOutcome_StillRunning(t *testing.T) {
	tasks := []model.Task{task("a", constants.StatusSuccess), task("b", constants.StatusRunning)}
	if got := scheduler.RunOutcome(tasks); got != "" {
		t.Fatalf("expected in-progress (empty string), got %q", got)
	}
}

func TestDepsFromDAG(t *testing.T) {
	defs := []model.DAGTask{
		{Name: "a", DependsOn: nil},
		{Name: "b", DependsOn: []string{"a"}},
	}
	deps := scheduler.DepsFromDAG(defs)
	if len(deps["b"]) != 1 || deps["b"][0] != "a" {
		t.Fatalf("unexpected deps for b: %v", deps["b"])
	}
	if deps["a"] != nil {
		t.Fatalf("expected nil deps for a, got %v", deps["a"])
	}
}
