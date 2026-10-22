//go:build (!linux && !windows && !darwin) || android || ios

package filepick

import "context"

// Pick returns ErrNoDialog where the system has no file dialog.
func Pick(ctx context.Context, _ string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}

	return "", ErrNoDialog
}
