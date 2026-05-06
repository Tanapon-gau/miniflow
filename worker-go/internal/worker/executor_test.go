package worker_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/model"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/worker"
)

func shellMsg(command string) model.TaskMessage {
	p, _ := json.Marshal(model.ShellPayload{Command: command})
	return model.TaskMessage{
		TaskID:         uuid.New(),
		RunID:          uuid.New(),
		Type:           "shell",
		Payload:        p,
		TimeoutSeconds: 10,
	}
}

func httpMsg(method, url, body string) model.TaskMessage {
	p, _ := json.Marshal(model.HTTPPayload{Method: method, URL: url, Body: body})
	return model.TaskMessage{
		TaskID:         uuid.New(),
		RunID:          uuid.New(),
		Type:           "http",
		Payload:        p,
		TimeoutSeconds: 10,
	}
}

// --- shell ---

func TestExecShell_Success(t *testing.T) {
	res := worker.ExecShell(context.Background(), shellMsg("echo hello"))
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Output != "hello\n" {
		t.Fatalf("unexpected output: %q", res.Output)
	}
}

func TestExecShell_NonZeroExit(t *testing.T) {
	res := worker.ExecShell(context.Background(), shellMsg("exit 1"))
	if res.Err == nil {
		t.Fatal("expected error for exit 1")
	}
}

func TestExecShell_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	msg := shellMsg("sleep 10")
	msg.TimeoutSeconds = 10 // outer timeout from context wins
	res := worker.ExecShell(ctx, msg)
	if res.Err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestExecShell_MissingCommand(t *testing.T) {
	p, _ := json.Marshal(model.ShellPayload{Command: ""})
	msg := model.TaskMessage{TaskID: uuid.New(), Type: "shell", Payload: p, TimeoutSeconds: 5}
	res := worker.ExecShell(context.Background(), msg)
	if res.Err == nil {
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

	res := worker.ExecHTTP(context.Background(), httpMsg(http.MethodGet, srv.URL, ""))
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestExecHTTP_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	res := worker.ExecHTTP(context.Background(), httpMsg(http.MethodGet, srv.URL, ""))
	if res.Err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestExecHTTP_DefaultMethodIsGET(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p, _ := json.Marshal(model.HTTPPayload{URL: srv.URL}) // no method
	msg := model.TaskMessage{TaskID: uuid.New(), Type: "http", Payload: p, TimeoutSeconds: 5}
	worker.ExecHTTP(context.Background(), msg)
	if gotMethod != http.MethodGet {
		t.Fatalf("expected GET, got %s", gotMethod)
	}
}

func TestExecHTTP_MissingURL(t *testing.T) {
	p, _ := json.Marshal(model.HTTPPayload{})
	msg := model.TaskMessage{TaskID: uuid.New(), Type: "http", Payload: p, TimeoutSeconds: 5}
	res := worker.ExecHTTP(context.Background(), msg)
	if res.Err == nil {
		t.Fatal("expected error for missing url")
	}
}
