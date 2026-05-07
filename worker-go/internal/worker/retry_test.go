package worker_test

import (
	"testing"
	"time"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/constants"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/worker"
)

func TestRetryDelay_AttemptOne(t *testing.T) {
	if got := worker.RetryDelay(1); got != constants.RetryBaseDelay {
		t.Fatalf("attempt 1: expected %s, got %s", constants.RetryBaseDelay, got)
	}
}

func TestRetryDelay_DoublesEachAttempt(t *testing.T) {
	prev := worker.RetryDelay(1)
	for attempt := 2; attempt <= 4; attempt++ {
		got := worker.RetryDelay(attempt)
		if got != prev*2 {
			t.Fatalf("attempt %d: expected %s (2× previous), got %s", attempt, prev*2, got)
		}
		prev = got
	}
}

func TestRetryDelay_CappedAtMax(t *testing.T) {
	for attempt := 10; attempt <= 20; attempt++ {
		if got := worker.RetryDelay(attempt); got != constants.RetryMaxDelay {
			t.Fatalf("attempt %d: expected cap %s, got %s", attempt, constants.RetryMaxDelay, got)
		}
	}
}

func TestRetryDelay_NeverExceedsMax(t *testing.T) {
	for attempt := 1; attempt <= 30; attempt++ {
		if got := worker.RetryDelay(attempt); got > constants.RetryMaxDelay {
			t.Fatalf("attempt %d: %s exceeds max %s", attempt, got, constants.RetryMaxDelay)
		}
	}
}

func TestRetryDelay_AlwaysPositive(t *testing.T) {
	for attempt := 1; attempt <= 10; attempt++ {
		if got := worker.RetryDelay(attempt); got <= 0 {
			t.Fatalf("attempt %d: expected positive delay, got %s", attempt, got)
		}
	}
}

func TestRetryDelay_Sequence(t *testing.T) {
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second}, // capped
	}
	for _, tc := range cases {
		if got := worker.RetryDelay(tc.attempt); got != tc.want {
			t.Errorf("attempt %d: expected %s, got %s", tc.attempt, tc.want, got)
		}
	}
}
