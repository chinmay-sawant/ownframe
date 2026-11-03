package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// Kbd displays one keyboard key or shortcut.
func Kbd(label string) *ui.Element {
	return ui.Span().Class("of-kbd").Text(label)
}
