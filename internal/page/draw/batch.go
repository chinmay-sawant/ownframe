package draw

// OpenBatch starts one nesting level. The window wraps a press, focus,
// click, and release so the final state draws once.
func OpenBatch(count *int) { *count++ }

// QueueBatch records a draw while a batch is open and reports that the
// caller should return without drawing. skipSame false also forces the
// draw when the outermost batch closes.
func QueueBatch(open int, pending, force *bool, skipSame bool) bool {
	if open <= 0 {
		return false
	}

	*pending = true
	if !skipSame {
		*force = true
	}

	return true
}

// CloseBatch ends one nesting level. run is true only for the outermost
// batch that still has a pending draw. forced is that draw's force flag.
// Pending and the flag are cleared when run is true. An inner level leaves
// both flags for the outer close.
func CloseBatch(count *int, pending, force *bool) (bool, bool) {
	if *count == 0 {
		return false, false
	}

	*count--
	if *count > 0 || !*pending {
		return false, false
	}

	forced := *force
	*pending = false
	*force = false

	return true, forced
}

// RestoreBatch puts back a draw CloseBatch cleared after that draw failed,
// so the next batch commits it.
func RestoreBatch(pending, force *bool, forced bool) {
	*pending = true
	*force = forced
}
