package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

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

// BRPop blocks until a message arrives on tasks:shell or tasks:http.
// Returns the raw JSON bytes and the queue name it came from.
func (q *Queue) BRPop(ctx context.Context) ([]byte, string, error) {
	res, err := q.client.BRPop(ctx, 5*time.Second, "tasks:shell", "tasks:http").Result()
	if err != nil {
		return nil, "", err
	}
	// BRPop returns [key, value]
	return []byte(res[1]), res[0], nil
}

// PublishEvent appends an event to the Redis stream "events".
func (q *Queue) PublishEvent(ctx context.Context, ev model.Event) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "events",
		Values: map[string]any{"data": string(data)},
	}).Err()
}
