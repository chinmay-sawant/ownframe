package ui

// errPreviewLen caps the error text in the status line, in runes.
const errPreviewLen = 80

// detailStatus words the selection's lifecycle and load state.
func detailStatus(sel selection) string {
	switch {
	case sel.id == "":
		return "no selection"
	case sel.trk.Err != "":
		return "error: " + truncate(sel.trk.Err, errPreviewLen)
	case sel.trk.Loading && !sel.trk.OK:
		return "loading\u2026"
	case !sel.found:
		return "no longer running"
	case sel.trk.OK:
		return "running"
	default:
		return "waiting for data"
	}
}

// nonEmpty returns s or the n/a placeholder.
func nonEmpty(s string) string {
	if s == "" {
		return unavailable
	}

	return s
}

// countOrNA renders a count that may be unknown.
func countOrNA(n int, ok bool) string {
	if !ok {
		return unavailable
	}

	return formatCount(n)
}

// countU64 renders a uint64 count that may be unknown.
func countU64(n uint64, ok bool) string {
	if !ok {
		return unavailable
	}

	return formatCount(int(n)) //nolint:gosec // reason: rendered counters stay far below MaxInt64; int is 64-bit on every supported platform.
}
