package page

import (
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/page/patch"
)

// displayOpPool rents operation buffers for display patches and frame ticks.
// It stays in this file so the gochecknoglobals allowlist still covers it.
//
// Retention note: unlike the single-slot dirty arenas, a sync.Pool drains
// across garbage collections, so a giant frame's grown Ops backing array
// pins only transiently between collections. Dropping over-cap buffers
// instead of pooling them would belong in patch.Rent/Free and their
// patch_wire.go call sites, which live outside this file.
var displayOpPool = sync.Pool{
	New: func() any {
		return patch.NewOpBuffer()
	},
}
