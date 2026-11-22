// Package uia maps host.AccessibilityTree to Windows UI Automation (UIA).
//
// The package has two layers. The portable layer (snapshot.go, provider.go)
// turns a detached accessibility snapshot into UIA elements: control types,
// names, states, screen bounds, and the Invoke, Value, Toggle, and
// SelectionItem patterns. It has no platform code, so Linux tests drive it
// with a fake UIA client. The Windows layer (*_windows.go) exposes that model
// to the operating system as COM server-side providers, built from syscall
// callbacks without cgo, and answers WM_GETOBJECT on the window.
//
// Password values never leave the snapshot builder.
package uia
