package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/db"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/model"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/queue"
)

type Worker struct {
	db    *db.DB
	queue *queue.Queue
}

func New(database *db.DB, q *queue.Queue) *Worker {
	return &Worker{db: database, queue: q}
}

// Run blocks, consuming tasks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		data, _, err := w.queue.BRPop(ctx)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				// timeout with no message — loop again
				continue
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			log.Printf("brpop error: %v", err)
			continue
		}
		w.handle(ctx, data)
	}
}

func (w *Worker) handle(ctx context.Context, data []byte) {
	var msg model.TaskMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("unmarshal task message: %v", err)
		return
	}

	if err := w.db.MarkTaskRunning(ctx, msg.TaskID); err != nil {
		log.Printf("mark running %s: %v", msg.TaskID, err)
		return
	}
	w.publishEvent(ctx, msg, "started", nil)

	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(msg.TimeoutSeconds)*time.Second)
	defer cancel()

	var result Result
	switch msg.Type {
	case "shell":
		result = ExecShell(taskCtx, msg)
	case "http":
		result = ExecHTTP(taskCtx, msg)
	default:
		log.Printf("unknown task type %q for task %s", msg.Type, msg.TaskID)
		return
	}

	w.publishEvent(ctx, msg, "log", map[string]string{"output": result.Output})

	status := "success"
	eventType := "succeeded"
	if result.Err != nil {
		status = "failed"
		eventType = "failed"
		log.Printf("task %s failed: %v", msg.TaskID, result.Err)
	}

	if err := w.db.MarkTaskDone(ctx, msg.TaskID, status); err != nil {
		log.Printf("mark done %s: %v", msg.TaskID, err)
	}
	w.publishEvent(ctx, msg, eventType, map[string]string{"status": status})
}

func (w *Worker) publishEvent(ctx context.Context, msg model.TaskMessage, eventType string, data any) {
	ev := model.Event{
		TaskID:    msg.TaskID,
		RunID:     msg.RunID,
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	if err := w.queue.PublishEvent(ctx, ev); err != nil {
		log.Printf("publish event %s for task %s: %v", eventType, msg.TaskID, err)
	}
}
