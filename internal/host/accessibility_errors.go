package host

import "errors"

var (
	ErrAccessibilityUnavailable = errors.New("ownframe: accessibility tree unavailable")
	ErrAccessibilityStale       = errors.New("ownframe: stale accessibility document")
	ErrAccessibilityNode        = errors.New("ownframe: accessibility node unavailable")
	ErrAccessibilityAction      = errors.New("ownframe: accessibility action unsupported")
	ErrAccessibilityDisabled    = errors.New("ownframe: accessibility node disabled")
	ErrAccessibilityRange       = errors.New("ownframe: invalid accessibility value")
)
