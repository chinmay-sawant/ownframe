//go:build darwin && !ios

package filepick

import (
	"context"
	"os/exec"
	"strings"
)

const script = `on run argv
POSIX path of (choose file with prompt (item 1 of argv))
end run`

// Pick shows the macOS open dialog through osascript.
func Pick(ctx context.Context, title string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	if _, err := exec.LookPath("osascript"); err != nil {
		return "", ErrNoDialog
	}
	out, err := exec.CommandContext(ctx, "osascript", "-e", script, title).Output()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxCanceled(ctxErr)
		}
		debugf("osascript failed: %v", err)

		return "", ErrCanceled
	}
	picked := strings.TrimRight(string(out), "\r\n")
	if picked == "" {
		return "", ErrCanceled
	}

	return picked, nil
}
