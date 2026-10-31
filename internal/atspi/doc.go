// Package atspi maps host.AccessibilityTree to AT-SPI objects.
//
// On Linux, NewLiveBus connects to the accessibility bus, claims an
// ownframe name, registers with org.a11y.atspi.Registry, and serves
// the mapped tree over D-Bus. FakeBus records calls for tests.
// Windows UI Automation, macOS VoiceOver, and Android TalkBack are
// not shipped.
package atspi
