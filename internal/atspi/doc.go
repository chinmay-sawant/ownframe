// Package atspi maps host.AccessibilityTree to AT-SPI objects.
//
// This is the Linux first slice: role and action mapping plus a bus
// seam. Tests use FakeBus. NewLiveBus never reports Live; a D-Bus
// AT-SPI client is not implemented. Windows UI Automation, macOS
// VoiceOver, and Android TalkBack are not shipped.
package atspi
