package snip

import "context"

// darwinFullscreen runs screencapture to a temp PNG file.
func darwinFullscreen(ctx context.Context) ([]byte, error) {
	return shotToFile(ctx, "screencapture", "-x", "-t", "png")
}
