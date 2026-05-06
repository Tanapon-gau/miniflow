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

// Result holds the output of a task execution.
type Result struct {
	Output string
	Err    error
}

// ExecShell runs the shell command specified in msg.Payload.
// The context carries the caller-supplied deadline (from timeout_seconds).
func ExecShell(ctx context.Context, msg model.TaskMessage) Result {
	var p model.ShellPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return Result{Err: fmt.Errorf("parse shell payload: %w", err)}
	}
	if p.Command == "" {
		return Result{Err: fmt.Errorf("shell payload missing 'command'")}
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", p.Command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return Result{Output: buf.String(), Err: fmt.Errorf("command failed: %w", err)}
	}
	return Result{Output: buf.String()}
}

// ExecHTTP performs the HTTP request specified in msg.Payload.
func ExecHTTP(ctx context.Context, msg model.TaskMessage) Result {
	var p model.HTTPPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return Result{Err: fmt.Errorf("parse http payload: %w", err)}
	}
	if p.URL == "" {
		return Result{Err: fmt.Errorf("http payload missing 'url'")}
	}
	if p.Method == "" {
		p.Method = http.MethodGet
	}

	var bodyReader io.Reader
	if p.Body != "" {
		bodyReader = strings.NewReader(p.Body)
	}

	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, bodyReader)
	if err != nil {
		return Result{Err: fmt.Errorf("build request: %w", err)}
	}
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: time.Duration(msg.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Err: fmt.Errorf("http request: %w", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	output := fmt.Sprintf("%s %s", resp.Status, string(body))

	if resp.StatusCode >= 400 {
		return Result{Output: output, Err: fmt.Errorf("http %d", resp.StatusCode)}
	}
	return Result{Output: output}
}
