// Package androidaccess lets the Android host read a page's accessibility
// tree as JSON and run the actions TalkBack asks for.
//
// A bound Android package cannot export Go structs, so the tree crosses the
// language boundary as one string. An app's bound package wraps Tree and
// Perform in two exported functions; the host's accessibility provider
// (examples/android-host, class AccessibilityBridge) reads the string and
// answers TalkBack from it.
//
// The window loop publishes the tree and runs the actions, because both
// belong to the page's owning loop. Tree returns the latest published
// snapshot and Perform queues the action for the loop, so both are safe to
// call from any thread, including Android's main thread and its binder
// threads. They report an empty tree and a refusal before a page is attached.
package androidaccess

import "github.com/chinmay-sawant/ownframe/internal/a11ybridge"

// Action numbers for Perform: 1 focus, 2 activate (tap), 3 set the value of a
// text field.
const (
	ActionFocus    = a11ybridge.ActionFocus
	ActionActivate = a11ybridge.ActionActivate
	ActionSetValue = a11ybridge.ActionSetValue
)

// Tree returns the page's latest accessibility snapshot as JSON, in document
// order, with bounds in CSS pixels of the page content. Password values are
// never included. Before the first snapshot it returns a tree with no nodes,
// so a host can always parse the result.
func Tree() string { return a11ybridge.Tree() }

// Perform runs one action TalkBack asked for. document is the token from the
// Tree the host read, so an action aimed at a node of a page that has since
// navigated is refused instead of landing on a different node. It returns an
// empty string on success and a short message otherwise.
func Perform(document int64, node string, action int, value string) string {
	return a11ybridge.Perform(document, node, action, value)
}
