package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Tanapon-gau/miniflow/scheduler/internal/db"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/model"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/queue"
)

type Scheduler struct {
	db       *db.DB
	queue    *queue.Queue
	interval time.Duration
}

func New(database *db.DB, q *queue.Queue, interval time.Duration) *Scheduler {
	return &Scheduler{db: database, queue: q, interval: interval}
}

// Run blocks, ticking on interval until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				log.Printf("tick error: %v", err)
			}
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	runs, err := s.db.ActiveRuns(ctx)
	if err != nil {
		return err
	}
	for _, run := range runs {
		if err := s.processRun(ctx, run); err != nil {
			log.Printf("run %s: %v", run.ID, err)
		}
	}
	return nil
}

func (s *Scheduler) processRun(ctx context.Context, run model.Run) error {
	if run.Status == "pending" {
		if err := s.db.MarkRunRunning(ctx, run.ID); err != nil {
			return err
		}
	}

	dagBytes, err := s.db.WorkflowDAG(ctx, run.WorkflowID)
	if err != nil {
		return err
	}
	var dagDef model.DAGDef
	if err := json.Unmarshal(dagBytes, &dagDef); err != nil {
		return err
	}
	deps := DepsFromDAG(dagDef.Tasks)

	tasks, err := s.db.TasksForRun(ctx, run.ID)
	if err != nil {
		return err
	}

	for _, t := range ReadyTasks(tasks, deps) {
		if err := s.dispatch(ctx, t); err != nil {
			log.Printf("dispatch task %s: %v", t.ID, err)
		}
	}

	// re-fetch after dispatch to get updated statuses
	tasks, err = s.db.TasksForRun(ctx, run.ID)
	if err != nil {
		return err
	}
	if outcome := RunOutcome(tasks); outcome != "" {
		return s.db.MarkRunFinished(ctx, run.ID, outcome)
	}
	return nil
}

func (s *Scheduler) dispatch(ctx context.Context, t model.Task) error {
	if err := s.db.MarkTaskQueued(ctx, t.ID); err != nil {
		return err
	}
	return s.queue.Push(ctx, queue.TaskMessage{
		TaskID:         t.ID,
		RunID:          t.RunID,
		Type:           t.Type,
		Payload:        t.Payload,
		TimeoutSeconds: t.TimeoutSeconds,
		MaxRetries:     t.MaxRetries,
		Attempt:        t.Attempt + 1,
	})
}
