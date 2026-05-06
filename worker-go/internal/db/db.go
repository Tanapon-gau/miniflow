package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (d *DB) MarkTaskRunning(ctx context.Context, id uuid.UUID) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE tasks SET status = 'running', started_at = $1 WHERE id = $2`,
		time.Now().UTC(), id)
	return err
}

func (d *DB) MarkTaskDone(ctx context.Context, id uuid.UUID, status string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE tasks SET status = $1, finished_at = $2 WHERE id = $3`,
		status, time.Now().UTC(), id)
	return err
}
