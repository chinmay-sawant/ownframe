package collector

import (
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// NewLive returns the platform's live source. Platforms without an adapter
// get a source that reports everything as unsupported, so the UI shows
// unavailable values instead of zeroes.
func NewLive() domain.Source { return liveSource() }
