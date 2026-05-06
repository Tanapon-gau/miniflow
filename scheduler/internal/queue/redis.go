package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/Tanapon-gau/miniflow/scheduler/internal/constants"
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

// Workers consume via BRPOP (rightmost element = oldest = FIFO).
func (q *Queue) Push(ctx context.Context, message TaskMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal task message for task %s: %w", message.TaskID, err)
	}
	return q.client.LPush(ctx, constants.TaskQueuePrefix+message.Type, data).Err()
}
