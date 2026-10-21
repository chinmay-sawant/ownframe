//go:build darwin && !ios

package filepick

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const script = `on run argv
POSIX path of (choose file with prompt (item 1 of argv))
end run`

// Pick shows the macOS open dialog through osascript.
// ok is false when the person canceled.
func Pick(ctx context.Context, title string) (string, bool) {
	out, err := exec.CommandContext(ctx, "osascript", "-e", script, title).Output()
	if err != nil {
		debugf("osascript failed: %v", err)

		return "", false
	}

	picked := strings.TrimRight(string(out), "\r\n")
	if picked == "" {
		return "", false
	}

	return picked, true
}

// debugf writes to stderr when OWNFRAME_FILEPICK_DEBUG is not empty.
// Values are sanitized to one line. Keep this flag off in shipped builds.
func debugf(format string, args ...any) {
	if os.Getenv("OWNFRAME_FILEPICK_DEBUG") == "" && os.Getenv("GPUI_FILEPICK_DEBUG") == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "filepick: "+format+"\n", cleanArgs(args)...)
}
