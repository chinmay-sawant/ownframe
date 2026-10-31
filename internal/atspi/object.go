package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

// Object is one AT-SPI accessible built from a page node.
type Object struct {
	Path, Parent        string
	Children            []string
	ID                  host.AccessibilityNodeID
	Role                Role
	Name, Description   string
	Value               string
	State               State
	Actions             []string
	Bounds              host.AccessibilityBounds
	HasBounds, Password bool
}

// State is the AT-SPI state set used in this slice.
type State struct {
	Enabled, Sensitive bool
	Focusable, Focused bool
	Checkable, Checked bool
	Editable, ReadOnly bool
	Active             bool
}

// Application is the published tree for one page snapshot.
type Application struct {
	Name, RootPath     string
	Document, Revision uint64
	Focused            host.AccessibilityNodeID
	Objects            []Object
}

// Event is a bus notification. It holds no accessible text beyond the
// node reference an AT needs to query the tree.
type Event struct {
	Path, Kind string
	Detail     string
	Child      string
	Index      int32
	Enabled    bool
}
