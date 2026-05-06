package constants

import "time"

const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"

	TaskTypeShell = "shell"
	TaskTypeHTTP  = "http"

	QueueShell = "tasks:shell"
	QueueHTTP  = "tasks:http"

	EventStarted   = "started"
	EventLog       = "log"
	EventSucceeded = "succeeded"
	EventFailed    = "failed"

	EventsStream = "events"
	BRPopTimeout = 5 * time.Second
)
