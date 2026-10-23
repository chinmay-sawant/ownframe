//go:build !linux || android || ios || js

package atspi

func liveReason() string {
	return "AT-SPI is Linux only; UIA, VoiceOver, and TalkBack are not shipped"
}
