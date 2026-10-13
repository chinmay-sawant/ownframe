package page

import "time"

// frameState is the page's per-frame opt-ins: the dirty-rect opt-in the
// window reads, and the next-frame deadline a page sets to pace an idle
// window. Both clear when the window reads them. It is embedded in Page so
// the fields promote and page_types.go stays inside the file size limit.
type frameState struct {
	frameDirty bool
	frameAt    time.Time
}
