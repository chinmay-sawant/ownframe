package collector

import (
	"context"
	"errors"
)

// worker runs queued calls with a deadline. A call that overruns the deadline
// is abandoned; the worker returns when the call finally ends.
func (p *Pool) worker() {
	defer p.wg.Done()

	for j := range p.jobs {
		p.busy.Add(1)
		ctx, cancel := context.WithTimeout(j.ctx, p.timeout)
		err := j.fn(ctx)
		cancel()
		p.busy.Add(-1)
		p.completed.Add(1)

		if errors.Is(err, context.DeadlineExceeded) {
			p.timedOut.Add(1)
		}
		if j.onDone != nil {
			j.onDone(err)
		}
	}
}
