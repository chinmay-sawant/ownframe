package ownframe

import "github.com/chinmay-sawant/ownframe/internal/host"

// Host is the inner screen the window and web loops call.
func (t Testing) Host() host.Screen {
	return innerPage(t.page)
}
