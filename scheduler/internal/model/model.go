package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Workflow struct {
	ID  uuid.UUID
	DAG json.RawMessage
}

type Run struct {
	ID          uuid.UUID
	WorkflowID  uuid.UUID
	Status      string
	TriggeredAt time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
}

type Task struct {
	ID             uuid.UUID
	RunID          uuid.UUID
	Name           string
	Type           string
	Status         string
	Payload        json.RawMessage
	Attempt        int
	MaxRetries     int
	TimeoutSeconds int
}

// DAGDef is the parsed workflow dag JSON.
type DAGDef struct {
	Tasks []DAGTask `json:"tasks"`
}

type DAGTask struct {
	Name           string   `json:"name"`
	DependsOn      []string `json:"depends_on"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxRetries     int      `json:"max_retries"`
}
