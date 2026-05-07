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
	if shift > 4 { // 2^4 * 1s = 16s; next doubling would exceed 30s cap
		shift = 4
	}
	delay := constants.RetryBaseDelay << uint(shift)
	if delay > constants.RetryMaxDelay {
		return constants.RetryMaxDelay
	}
	return delay
}
