package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TaskMessage struct {
	TaskID         uuid.UUID       `json:"task_id"`
	RunID          uuid.UUID       `json:"run_id"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	TimeoutSeconds int             `json:"timeout_seconds"`
	MaxRetries     int             `json:"max_retries"`
	Attempt        int             `json:"attempt"`
}

type Queue struct {
	client *redis.Client
}

func New(addr string) *Queue {
	return &Queue{client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (q *Queue) Close() error {
	return q.client.Close()
}

// Push serialises msg and prepends it to "tasks:{type}".
// Workers consume via BRPOP (rightmost element = oldest = FIFO).
func (q *Queue) Push(ctx context.Context, msg TaskMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal task message: %w", err)
	}
	return q.client.LPush(ctx, "tasks:"+msg.Type, data).Err()
}
