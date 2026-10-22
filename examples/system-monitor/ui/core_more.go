package ui

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/collector"
)

// problemErrLen caps the error text in the problems list.
const problemErrLen = 90

// Problems lists the collector's latest failure per step.
func (c *Collector) Problems(context.Context) []string {
	errs := c.mgr.Errors()
	out := make([]string, 0, len(errs))

	for _, e := range errs {
		out = append(out, e.Step+": "+truncate(e.Err, problemErrLen))
	}

	return out
}

// SetLive switches the manager between the dummy and live sources. The
// manager clears its baselines and graph buffers on the switch.
func (c *Collector) SetLive(live bool) error {
	if live {
		return c.mgr.SetMode(collector.ModeLive)
	}

	return c.mgr.SetMode(collector.ModeDummy)
}

// Close stops the manager's loops and pool within the close budget.
func (c *Collector) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), closeBudget)
	defer cancel()

	return c.mgr.Close(ctx)
}
