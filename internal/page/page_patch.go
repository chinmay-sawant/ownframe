package page

import (
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/page/patch"
)

// displayOpPool rents operation buffers for display patches and frame ticks.
// It stays in this file so the gochecknoglobals allowlist still covers it.
var displayOpPool = sync.Pool{
	New: func() any {
		b := patch.EmptyOps()
		return &b
	},
}
