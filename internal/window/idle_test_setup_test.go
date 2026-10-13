package window

import (
	"testing"
	"time"
)

// paceStep is one vsync frame at 60 Hz.
const paceStep = 16 * time.Millisecond

// idleAt steps clean frames until the policy idles and returns that time.
func idleAt(t *testing.T, p *pacePolicy, from time.Time) time.Time {
	t.Helper()

	at := from
	for i := 0; i < 100; i++ {
		at = at.Add(paceStep)
		if p.step(at, false, false) == paceToIdle {
			return at
		}
	}

	t.Fatal("policy never idled")
	return at
}
