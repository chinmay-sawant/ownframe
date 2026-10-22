package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// retryBusy runs fn and retries only lock failures, within the context
// deadline. Every caller's operation is idempotent.
func retryBusy(ctx context.Context, fn func() error) error {
	delay := retryDelay

	for {
		err := fn()
		if err == nil || !isBusy(err) {
			return err
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()

			return fmt.Errorf("store: locked: %w", ctx.Err())
		case <-timer.C:
		}

		if delay < 200*time.Millisecond {
			delay *= 2
		}
	}
}

// isBusy reports SQLite lock contention. A busy statement did not commit,
// so a retry cannot replay a half-applied write.
func isBusy(err error) bool {
	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "sqlite_busy") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database table is locked")
}

// clampLimit bounds a score page size.
func clampLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}

	return min(limit, maxLimit)
}
