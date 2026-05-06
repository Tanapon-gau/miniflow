package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Tanapon-gau/miniflow/scheduler/internal/constants"
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
			log.Printf("process run %s failed: %v", run.ID, err)
		}
	}
	return nil
}

func (s *Scheduler) processRun(ctx context.Context, run model.Run) error {
	if run.Status == constants.StatusPending {
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

	for _, task := range ReadyTasks(tasks, deps) {
		if err := s.dispatch(ctx, task); err != nil {
			log.Printf("dispatch task %s failed: %v", task.ID, err)
		}
	}

	tasks, err = s.db.TasksForRun(ctx, run.ID)
	if err != nil {
		return err
	}
	if outcome := RunOutcome(tasks); outcome != "" {
		return s.db.MarkRunFinished(ctx, run.ID, outcome)
	}
	return nil
}

func (s *Scheduler) dispatch(ctx context.Context, task model.Task) error {
	if err := s.db.MarkTaskQueued(ctx, task.ID); err != nil {
		return err
	}
	return s.queue.Push(ctx, queue.TaskMessage{
		TaskID:         task.ID,
		RunID:          task.RunID,
		Type:           task.Type,
		Payload:        task.Payload,
		TimeoutSeconds: task.TimeoutSeconds,
		MaxRetries:     task.MaxRetries,
		Attempt:        task.Attempt + 1,
	})
}
