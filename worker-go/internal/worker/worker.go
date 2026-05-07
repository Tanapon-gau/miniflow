package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/constants"
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

func (w *Worker) Run(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		data, _, err := w.queue.BRPop(ctx)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			log.Printf("brpop failed: %v", err)
			continue
		}
		w.handle(ctx, data)
	}
}

func (w *Worker) handle(ctx context.Context, data []byte) {
	var message model.TaskMessage
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("unmarshal task message failed: %v — raw: %.200s", err, data)
		return
	}

	if err := w.db.MarkTaskRunning(ctx, message.TaskID, message.Attempt); err != nil {
		log.Printf("mark task %s running failed: %v", message.TaskID, err)
		return
	}
	w.publishEvent(ctx, message, constants.EventStarted, nil)

	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(message.TimeoutSeconds)*time.Second)
	defer cancel()

	var result Result
	switch message.Type {
	case constants.TaskTypeShell:
		result = ExecShell(taskCtx, message)
	case constants.TaskTypeHTTP:
		result = ExecHTTP(taskCtx, message)
	default:
		log.Printf("task %s: unknown type %q", message.TaskID, message.Type)
		return
	}

	w.publishEvent(ctx, message, constants.EventLog, map[string]string{"output": result.Output})

	if result.Err != nil {
		log.Printf("task %s attempt %d failed: %v", message.TaskID, message.Attempt, result.Err)
		if message.Attempt <= message.MaxRetries {
			w.scheduleRetry(ctx, message)
			return
		}
		if err := w.db.MarkTaskDone(ctx, message.TaskID, constants.StatusFailed); err != nil {
			log.Printf("mark task %s done failed: %v", message.TaskID, err)
		}
		w.publishEvent(ctx, message, constants.EventFailed, map[string]string{"status": constants.StatusFailed})
		return
	}

	if err := w.db.MarkTaskDone(ctx, message.TaskID, constants.StatusSuccess); err != nil {
		log.Printf("mark task %s done failed: %v", message.TaskID, err)
	}
	w.publishEvent(ctx, message, constants.EventSucceeded, map[string]string{"status": constants.StatusSuccess})
}

func (w *Worker) scheduleRetry(ctx context.Context, message model.TaskMessage) {
	nextAttempt := message.Attempt + 1
	delay := RetryDelay(message.Attempt)

	if err := w.db.MarkTaskRetrying(ctx, message.TaskID, message.Attempt); err != nil {
		log.Printf("mark task %s retrying failed: %v", message.TaskID, err)
	}
	w.publishEvent(ctx, message, constants.EventRetrying,
		map[string]string{"attempt": fmt.Sprintf("%d", message.Attempt), "next_in": delay.String()})

	log.Printf("task %s: attempt %d/%d failed, retrying in %s",
		message.TaskID, message.Attempt, message.MaxRetries+1, delay)

	retryMessage := message
	retryMessage.Attempt = nextAttempt

	// sleep and re-queue in a goroutine so the consumer loop stays unblocked
	go func() {
		select {
		case <-time.After(delay):
			if err := w.queue.Push(ctx, retryMessage); err != nil {
				log.Printf("re-queue task %s attempt %d failed: %v", message.TaskID, nextAttempt, err)
			}
		case <-ctx.Done():
		}
	}()
}

func (w *Worker) publishEvent(ctx context.Context, message model.TaskMessage, eventType string, data any) {
	event := model.Event{
		TaskID:    message.TaskID,
		RunID:     message.RunID,
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	if err := w.queue.PublishEvent(ctx, event); err != nil {
		log.Printf("publish %s event for task %s failed: %v", eventType, message.TaskID, err)
	}
}
