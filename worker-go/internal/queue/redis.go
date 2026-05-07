package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/constants"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/model"
)

type Queue struct {
	client *redis.Client
}

func New(addr string) *Queue {
	return &Queue{client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (q *Queue) Close() error {
	return q.client.Close()
}

func (q *Queue) BRPop(ctx context.Context) ([]byte, string, error) {
	result, err := q.client.BRPop(ctx, constants.BRPopTimeout, constants.QueueShell, constants.QueueHTTP).Result()
	if err != nil {
		return nil, "", err
	}
	// BRPop returns [key, value]
	return []byte(result[1]), result[0], nil
}

func (q *Queue) Push(ctx context.Context, message model.TaskMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal task message for task %s: %w", message.TaskID, err)
	}
	var queueName string
	switch message.Type {
	case constants.TaskTypeShell:
		queueName = constants.QueueShell
	case constants.TaskTypeHTTP:
		queueName = constants.QueueHTTP
	default:
		return fmt.Errorf("task %s: unknown type %q, cannot re-queue", message.TaskID, message.Type)
	}
	return q.client.LPush(ctx, queueName, data).Err()
}

func (q *Queue) PublishEvent(ctx context.Context, event model.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event for task %s: %w", event.TaskID, err)
	}
	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.EventsStream,
		Values: map[string]any{"data": string(data)},
	}).Err()
}
