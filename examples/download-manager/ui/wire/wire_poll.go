package wire

import (
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/scheduler"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
)

// Poll returns up to budget results without waiting: engine events first,
// then worker answers.
func (b *Backend) Poll(budget int) []ui.Update {
	if budget <= 0 {
		return nil
	}

	out := make([]ui.Update, 0, budget)
	for _, ev := range b.eng.Drain(budget) {
		out = append(out, b.event(ev))
	}

	for len(out) < budget {
		select {
		case u := <-b.results:
			out = append(out, u)
		default:
			return out
		}
	}

	return out
}

// event maps one engine event. Every event feeds the speed tracker; a state
// event also resets the rate baseline, so a pause or resume gap cannot drag
// the smoothed rate toward zero. A state event carries the full job,
// including a terminal state.
func (b *Backend) event(ev scheduler.Event) ui.Update {
	b.track(ev)

	row := b.mapJob(ev.Job)

	return ui.Update{Kind: ui.UpdateProgress, Row: &row}
}

// track folds one event into the per-job speed. Progress samples come from
// domain.Rate, which refuses a gap longer than MaxSampleGap; a state event
// resets the baseline.
func (b *Backend) track(ev scheduler.Event) {
	at := ev.At
	if at.IsZero() {
		at = time.Now()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	prev := b.speeds[ev.Job.ID]

	if !prev.at.IsZero() {
		if inst, ok := domain.Rate(prev.done, prev.at, ev.Job.Done, at, domain.MaxSampleGap); ok {
			if prev.bytes <= 0 {
				prev.bytes = inst
			} else {
				prev.bytes = speedAlpha*prev.bytes + speedBeta*inst
			}
		}
	}

	if ev.Kind == scheduler.EventState {
		prev.bytes = 0
	}

	prev.done, prev.at = ev.Job.Done, at
	b.speeds[ev.Job.ID] = prev
}
