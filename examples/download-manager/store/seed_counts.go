package store

func dummyCounts(n int) (completed, failed, cancelled, queued int) {
	failed = n / seedFailDiv
	cancelled = n / seedCancelDiv
	queued = n / seedQueuedDiv
	paused := n / seedPausedDiv

	completed = n - failed - cancelled - queued - paused
	if completed < 0 {
		completed = 0
	}

	return completed, failed, cancelled, queued
}
