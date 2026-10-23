package mobile

import "github.com/chinmay-sawant/ownframe/examples/lg-remote/bridge"

// Accessibility returns JSON for this app's virtual nodes, not OwnFrame 1.0 accessibility.
func Accessibility() string { return bridge.Accessibility() }

// QueueAction puts a native accessibility action on the game loop.
func QueueAction(action string) bool { return bridge.QueueAction(action) }
