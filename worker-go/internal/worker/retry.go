package worker

import (
	"time"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/constants"
)

// RetryDelay returns the backoff duration before the next attempt.
// attempt is the attempt number that just failed (1-based).
// Formula: base * 2^(attempt-1), capped at RetryMaxDelay.
func RetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return constants.RetryBaseDelay
	}
	shift := attempt - 1
	if shift > 30 { // guard against int64 overflow on pathological attempt counts
		return constants.RetryMaxDelay
	}
	delay := constants.RetryBaseDelay << uint(shift)
	if delay > constants.RetryMaxDelay {
		return constants.RetryMaxDelay
	}
	return delay
}
