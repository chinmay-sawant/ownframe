//go:build linux && !android && !ios && !js

package atspi

func liveReason() string {
	return "at-spi bus socket not found"
}
