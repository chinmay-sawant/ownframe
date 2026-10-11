package ownframe

import "github.com/chinmay-sawant/ownframe/internal/host"

// ScrollbarMode controls whether a page shows scrollbar thumbs.
type ScrollbarMode = host.ScrollbarMode

const (
	// ScrollbarAuto follows the platform default: visible on desktop and web,
	// hidden on Android and iOS.
	ScrollbarAuto = host.ScrollbarAuto
	// ScrollbarVisible shows scrollbar thumbs on every platform.
	ScrollbarVisible = host.ScrollbarVisible
	// ScrollbarHidden hides scrollbar thumbs on every platform.
	ScrollbarHidden = host.ScrollbarHidden
)
