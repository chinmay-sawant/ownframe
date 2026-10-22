package window

import (
	"os"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// pacePage is the wrapper's page-driven state: the deadline captured from the
// page, the timer kick channel, and the debug counters. The pacing wrapper embeds it.
type pacePage struct {
	want          time.Time
	kick          chan struct{}
	debug         *os.File
	debugAt       time.Time
	debugPresents uint64
}

// captureFrameWant reads the page's next-frame request once per tick, the way
// captureFrameOptIn reads the frame-dirty opt-in. The page clears the request
// when it is read, so a page that wants another frame asks again on every
// frame. A screen without host.FrameWanter never asks, and the window keeps
// its keep-alive fallback.
func (g *paceGame) captureFrameWant() {
	g.want = time.Time{}

	if wanter, ok := g.shell.app.(host.FrameWanter); ok {
		g.want = wanter.NextFrameAt()
	}
}

// publish copies the pacer's state to the shell for the devtools Frame tab.
// The game goroutine writes it; the timer goroutine owns the policy lock.
func (g *paceGame) publish(now time.Time) {
	g.mu.Lock()
	next, skipped := g.policy.next(), g.policy.skipped
	presents, idle := g.policy.presents, !g.policy.fast
	g.mu.Unlock()

	g.shell.pace.next, g.shell.pace.skipped, g.shell.pace.idle = next, skipped, idle
	g.debugLine(now, presents, skipped, idle)
}

// kickTimer asks the timer goroutine to recompute its sleep after the policy
// moved. The channel holds one pending kick, so the game loop never blocks.
func (g *paceGame) kickTimer() {
	select {
	case g.kick <- struct{}{}:
	default:
	}
}
