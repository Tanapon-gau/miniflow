package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/model"
)

type Result struct {
	Output string
	Err    error
}

func ExecShell(ctx context.Context, message model.TaskMessage) Result {
	var payload model.ShellPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return Result{Err: fmt.Errorf("task %s: parse shell payload: %w", message.TaskID, err)}
	}
	if payload.Command == "" {
		return Result{Err: fmt.Errorf("task %s: shell payload missing required field 'command'", message.TaskID)}
	}

	// context carries the deadline set by timeout_seconds
	cmd := exec.CommandContext(ctx, "sh", "-c", payload.Command)
	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	if err := cmd.Run(); err != nil {
		return Result{Output: outputBuf.String(), Err: fmt.Errorf("task %s: command %q failed: %w", message.TaskID, payload.Command, err)}
	}
	return Result{Output: outputBuf.String()}
}

func ExecHTTP(ctx context.Context, message model.TaskMessage) Result {
	var payload model.HTTPPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return Result{Err: fmt.Errorf("task %s: parse http payload: %w", message.TaskID, err)}
	}
	if payload.URL == "" {
		return Result{Err: fmt.Errorf("task %s: http payload missing required field 'url'", message.TaskID)}
	}
	if payload.Method == "" {
		payload.Method = http.MethodGet
	}

	var bodyReader io.Reader
	if payload.Body != "" {
		bodyReader = strings.NewReader(payload.Body)
	}

	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, bodyReader)
	if err != nil {
		return Result{Err: fmt.Errorf("task %s: build %s request to %s: %w", message.TaskID, payload.Method, payload.URL, err)}
	}
	for headerName, headerValue := range payload.Headers {
		req.Header.Set(headerName, headerValue)
	}

	client := &http.Client{Timeout: time.Duration(message.TimeoutSeconds) * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return Result{Err: fmt.Errorf("task %s: %s %s failed: %w", message.TaskID, payload.Method, payload.URL, err)}
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)
	output := fmt.Sprintf("%s %s", response.Status, string(responseBody))

	if response.StatusCode >= 400 {
		return Result{Output: output, Err: fmt.Errorf("task %s: %s %s returned status %d", message.TaskID, payload.Method, payload.URL, response.StatusCode)}
	}
	return Result{Output: output}
}
