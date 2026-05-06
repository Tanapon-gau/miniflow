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
		return nil, err
	}
	defer rows.Close()

	var runs []model.Run
	for rows.Next() {
		var r model.Run
		var id, wfID string
		if err := rows.Scan(&id, &wfID, &r.Status, &r.TriggeredAt, &r.StartedAt, &r.FinishedAt); err != nil {
			return nil, err
		}
		r.ID, _ = uuid.Parse(id)
		r.WorkflowID, _ = uuid.Parse(wfID)
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (d *DB) WorkflowDAG(ctx context.Context, id uuid.UUID) ([]byte, error) {
	var dag []byte
	err := d.pool.QueryRow(ctx,
		`SELECT dag FROM workflows WHERE id = $1`, id).Scan(&dag)
	return dag, err
}

func (d *DB) TasksForRun(ctx context.Context, runID uuid.UUID) ([]model.Task, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT id, run_id, name, type, status, payload, attempt, max_retries, timeout_seconds
		 FROM tasks WHERE run_id = $1`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		var id, runID string
		var payload []byte
		if err := rows.Scan(&id, &runID, &t.Name, &t.Type, &t.Status, &payload,
			&t.Attempt, &t.MaxRetries, &t.TimeoutSeconds); err != nil {
			return nil, err
		}
		t.ID, _ = uuid.Parse(id)
		t.RunID, _ = uuid.Parse(runID)
		t.Payload = payload
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (d *DB) MarkRunRunning(ctx context.Context, id uuid.UUID) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE runs SET status = 'running', started_at = $1 WHERE id = $2`,
		time.Now().UTC(), id)
	return err
}

func (d *DB) MarkRunFinished(ctx context.Context, id uuid.UUID, status string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE runs SET status = $1, finished_at = $2 WHERE id = $3`,
		status, time.Now().UTC(), id)
	return err
}

func (d *DB) MarkTaskQueued(ctx context.Context, id uuid.UUID) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE tasks SET status = 'queued' WHERE id = $1`, id)
	return err
}
