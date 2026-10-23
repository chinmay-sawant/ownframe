//go:build linux && !android && !ios && !js

package atspi

func liveReason() string {
	return "linux AT-SPI mapping is experimental; D-Bus client is not implemented"
}
