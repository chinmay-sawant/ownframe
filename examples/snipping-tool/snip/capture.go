// Package snip captures fullscreen screenshots for the snipping tool.
// Fullscreen shells out to OS tools so the example stays cgo-free.
// DemoImage builds a fallback picture when no tool is available.
package snip

import (
	"context"
	"runtime"
	"time"
)

// Capturer takes a fullscreen shot and returns PNG bytes.
type Capturer func(context.Context) ([]byte, error)

// CaptureWith runs one capturer, or Fullscreen when c is nil.
func CaptureWith(ctx context.Context, c Capturer) ([]byte, error) {
	if c != nil {
		return c(ctx)
	}
	return Fullscreen(ctx)
}

// Fullscreen captures the primary monitor as PNG bytes.
// Linux tries import, gnome-screenshot, scrot, then grim.
// Windows uses PowerShell System.Drawing. Darwin uses screencapture.
func Fullscreen(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	switch runtime.GOOS {
	case "windows":
		return windowsFullscreen(ctx)
	case "darwin":
		return darwinFullscreen(ctx)
	default:
		return linuxFullscreen(ctx)
	}
}
