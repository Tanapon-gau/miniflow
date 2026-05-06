package scheduler_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/model"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/scheduler"
)

func task(name, status string) model.Task {
	return model.Task{ID: uuid.New(), Name: name, Status: status}
}

func noDeps(names ...string) map[string][]string {
	m := make(map[string][]string)
	for _, n := range names {
		m[n] = nil
	}
	return m
}

func TestReadyTasks_NoDeps_AllPending(t *testing.T) {
	tasks := []model.Task{task("a", "pending"), task("b", "pending")}
	ready := scheduler.ReadyTasks(tasks, noDeps("a", "b"))
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready, got %d", len(ready))
	}
}

func TestReadyTasks_BlockedByRunningDep(t *testing.T) {
	tasks := []model.Task{task("a", "running"), task("b", "pending")}
	deps := map[string][]string{"a": nil, "b": {"a"}}
	ready := scheduler.ReadyTasks(tasks, deps)
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready, got %d", len(ready))
	}
}

func TestReadyTasks_UnblockedAfterDepSucceeds(t *testing.T) {
	tasks := []model.Task{task("a", "success"), task("b", "pending")}
	deps := map[string][]string{"a": nil, "b": {"a"}}
	ready := scheduler.ReadyTasks(tasks, deps)
	if len(ready) != 1 || ready[0].Name != "b" {
		t.Fatalf("expected b to be ready, got %v", ready)
	}
}

func TestReadyTasks_SkipsNonPending(t *testing.T) {
	tasks := []model.Task{task("a", "queued"), task("b", "running"), task("c", "success")}
	ready := scheduler.ReadyTasks(tasks, noDeps("a", "b", "c"))
	if len(ready) != 0 {
		t.Fatalf("expected 0, got %d", len(ready))
	}
}

func TestRunOutcome_AllSuccess(t *testing.T) {
	tasks := []model.Task{task("a", "success"), task("b", "success")}
	if got := scheduler.RunOutcome(tasks); got != "success" {
		t.Fatalf("expected success, got %q", got)
	}
}

func TestRunOutcome_OneFailed(t *testing.T) {
	tasks := []model.Task{task("a", "success"), task("b", "failed")}
	if got := scheduler.RunOutcome(tasks); got != "failed" {
		t.Fatalf("expected failed, got %q", got)
	}
}

func TestRunOutcome_StillRunning(t *testing.T) {
	tasks := []model.Task{task("a", "success"), task("b", "running")}
	if got := scheduler.RunOutcome(tasks); got != "" {
		t.Fatalf("expected in-progress, got %q", got)
	}
}

func TestDepsFromDAG(t *testing.T) {
	defs := []model.DAGTask{
		{Name: "a", DependsOn: nil},
		{Name: "b", DependsOn: []string{"a"}},
	}
	deps := scheduler.DepsFromDAG(defs)
	if len(deps["b"]) != 1 || deps["b"][0] != "a" {
		t.Fatalf("unexpected deps: %v", deps)
	}
	if deps["a"] != nil {
		t.Fatalf("expected nil deps for a, got %v", deps["a"])
	}
}
