package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
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

type ShellPayload struct {
	Command string `json:"command"`
}

type HTTPPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Event struct {
	TaskID    uuid.UUID `json:"task_id"`
	RunID     uuid.UUID `json:"run_id"`
	EventType string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}
