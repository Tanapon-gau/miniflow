package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tanapon-gau/miniflow/scheduler/internal/model"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (d *DB) Close() {
	d.pool.Close()
}

func (d *DB) ActiveRuns(ctx context.Context) ([]model.Run, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT id, workflow_id, status, triggered_at, started_at, finished_at
		 FROM runs WHERE status IN ('pending', 'running')`)
	if err != nil {
		return nil, fmt.Errorf("query active runs: %w", err)
	}
	defer rows.Close()

	var runs []model.Run
	for rows.Next() {
		var run model.Run
		var id, workflowID string
		if err := rows.Scan(&id, &workflowID, &run.Status, &run.TriggeredAt, &run.StartedAt, &run.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan run row: %w", err)
		}
		run.ID, _ = uuid.Parse(id)
		run.WorkflowID, _ = uuid.Parse(workflowID)
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (d *DB) WorkflowDAG(ctx context.Context, workflowID uuid.UUID) ([]byte, error) {
	var dag []byte
	if err := d.pool.QueryRow(ctx,
		`SELECT dag FROM workflows WHERE id = $1`, workflowID).Scan(&dag); err != nil {
		return nil, fmt.Errorf("query dag for workflow %s: %w", workflowID, err)
	}
	return dag, nil
}

func (d *DB) TasksForRun(ctx context.Context, runID uuid.UUID) ([]model.Task, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT id, run_id, name, type, status, payload, attempt, max_retries, timeout_seconds
		 FROM tasks WHERE run_id = $1`, runID)
	if err != nil {
		return nil, fmt.Errorf("query tasks for run %s: %w", runID, err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		var id, taskRunID string
		var payload []byte
		if err := rows.Scan(&id, &taskRunID, &task.Name, &task.Type, &task.Status, &payload,
			&task.Attempt, &task.MaxRetries, &task.TimeoutSeconds); err != nil {
			return nil, fmt.Errorf("scan task row for run %s: %w", runID, err)
		}
		task.ID, _ = uuid.Parse(id)
		task.RunID, _ = uuid.Parse(taskRunID)
		task.Payload = payload
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (d *DB) MarkRunRunning(ctx context.Context, runID uuid.UUID) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE runs SET status = 'running', started_at = $1 WHERE id = $2`,
		time.Now().UTC(), runID)
	if err != nil {
		return fmt.Errorf("mark run %s running: %w", runID, err)
	}
	return nil
}

func (d *DB) MarkRunFinished(ctx context.Context, runID uuid.UUID, taskStatus string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE runs SET status = $1, finished_at = $2 WHERE id = $3`,
		taskStatus, time.Now().UTC(), runID)
	if err != nil {
		return fmt.Errorf("mark run %s finished with status %q: %w", runID, taskStatus, err)
	}
	return nil
}

func (d *DB) MarkTaskQueued(ctx context.Context, taskID uuid.UUID) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE tasks SET status = 'queued' WHERE id = $1`, taskID)
	if err != nil {
		return fmt.Errorf("mark task %s queued: %w", taskID, err)
	}
	return nil
}
