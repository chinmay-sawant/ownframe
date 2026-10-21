package page

import "github.com/chinmay-sawant/ownframe/internal/host"

// pageScrollState is the scroll state of a page. The window writes it and
// the page loop reads it.
type pageScrollState struct {
	scroll    host.Scroll
	hasScroll bool
	windowing windowingState
}
