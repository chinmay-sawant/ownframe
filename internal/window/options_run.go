package window

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// RunWithOptions opens a window with optional desktop settings.
func RunWithOptions(ctx context.Context, app host.Screen, options Options) error {
	if ctx == nil {
		return errNilContext
	}

	if app == nil {
		return errNilApp
	}

	return runWindow(ctx, app, options)
}
