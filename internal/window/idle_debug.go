package window

import (
	"fmt"
	"os"
	"time"
)

// paceDebugEnv turns on one pacing line per second on stderr, so a headless
// CPU run can report the present rate without opening the devtools dock.
// It is dev-only and off unless explicitly set; keep it off in shipped builds.
const paceDebugEnv = "OWNFRAME_PACE_DEBUG"

// paceDebugWriter returns stderr when the debug environment variable is set
// and nil otherwise. It is read once, when the wrapper starts.
func paceDebugWriter() *os.File {
	if os.Getenv(paceDebugEnv) == "" {
		return nil
	}

	return os.Stderr
}

// debugLine writes presents per second, the skipped total, and the idle flag
// once a second while the debug variable is set. The counts come from the
// publish that just read them under the policy lock.
func (g *paceGame) debugLine(now time.Time, presents, skipped uint64, idle bool) {
	if g.debug == nil || now.Sub(g.debugAt) < time.Second {
		return
	}

	dp := presents - g.debugPresents
	g.debugAt, g.debugPresents = now, presents

	_, _ = fmt.Fprintf(g.debug, "pace: presents=%d/s skipped=%d idle=%v\n", dp, skipped, idle)
}
