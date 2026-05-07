package constants

import "time"

const (
	StatusRunning   = "running"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"

	TaskTypeShell = "shell"
	TaskTypeHTTP  = "http"

	QueueShell = "tasks:shell"
	QueueHTTP  = "tasks:http"

	EventStarted   = "started"
	EventLog       = "log"
	EventSucceeded = "succeeded"
	EventFailed    = "failed"
	EventRetrying  = "retrying"

	EventsStream = "events"
	BRPopTimeout = 5 * time.Second

	RetryBaseDelay = 1 * time.Second
	RetryMaxDelay  = 30 * time.Second
)
