package worker_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/constants"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/model"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/worker"
)

func shellMsg(command string) model.TaskMessage {
	payload, _ := json.Marshal(model.ShellPayload{Command: command})
	return model.TaskMessage{
		TaskID:         uuid.New(),
		RunID:          uuid.New(),
		Type:           constants.TaskTypeShell,
		Payload:        payload,
		TimeoutSeconds: 10,
	}
}

func httpMsg(method, url, body string) model.TaskMessage {
	payload, _ := json.Marshal(model.HTTPPayload{Method: method, URL: url, Body: body})
	return model.TaskMessage{
		TaskID:         uuid.New(),
		RunID:          uuid.New(),
		Type:           constants.TaskTypeHTTP,
		Payload:        payload,
		TimeoutSeconds: 10,
	}
}

// --- shell ---

func TestExecShell_Success(t *testing.T) {
	result := worker.ExecShell(context.Background(), shellMsg("echo hello"))
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Output != "hello\n" {
		t.Fatalf("unexpected output: %q", result.Output)
	}
}

func TestExecShell_NonZeroExit(t *testing.T) {
	result := worker.ExecShell(context.Background(), shellMsg("exit 1"))
	if result.Err == nil {
		t.Fatal("expected error for exit 1")
	}
}

func TestExecShell_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	message := shellMsg("sleep 10")
	// outer timeout from context wins over message.TimeoutSeconds
	result := worker.ExecShell(ctx, message)
	if result.Err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestExecShell_MissingCommand(t *testing.T) {
	payload, _ := json.Marshal(model.ShellPayload{Command: ""})
	message := model.TaskMessage{
		TaskID:         uuid.New(),
		Type:           constants.TaskTypeShell,
		Payload:        payload,
		TimeoutSeconds: 5,
	}
	result := worker.ExecShell(context.Background(), message)
	if result.Err == nil {
		t.Fatal("expected error for empty command")
	}
}

// --- http ---

func TestExecHTTP_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	result := worker.ExecHTTP(context.Background(), httpMsg(http.MethodGet, srv.URL, ""))
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func TestExecHTTP_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	result := worker.ExecHTTP(context.Background(), httpMsg(http.MethodGet, srv.URL, ""))
	if result.Err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestExecHTTP_DefaultMethodIsGET(t *testing.T) {
	var receivedMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	payload, _ := json.Marshal(model.HTTPPayload{URL: srv.URL})
	message := model.TaskMessage{
		TaskID:         uuid.New(),
		Type:           constants.TaskTypeHTTP,
		Payload:        payload,
		TimeoutSeconds: 5,
	}
	worker.ExecHTTP(context.Background(), message)
	if receivedMethod != http.MethodGet {
		t.Fatalf("expected GET, got %s", receivedMethod)
	}
}

func TestExecHTTP_MissingURL(t *testing.T) {
	payload, _ := json.Marshal(model.HTTPPayload{})
	message := model.TaskMessage{
		TaskID:         uuid.New(),
		Type:           constants.TaskTypeHTTP,
		Payload:        payload,
		TimeoutSeconds: 5,
	}
	result := worker.ExecHTTP(context.Background(), message)
	if result.Err == nil {
		t.Fatal("expected error for missing url")
	}
}
