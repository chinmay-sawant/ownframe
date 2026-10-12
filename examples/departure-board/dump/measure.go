package dump

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// Scenarios lists the scenario names Measure accepts, in report order.
func Scenarios() []string {
	return []string{"warmup", "flap", "idle", "loop", "cascade"}
}

// Measure runs one scenario against app and returns its report. The caller
// passes a freshly built app per scenario, and the shapes are in
// measure_scenario.go. Measure is headless and opens no window.
func Measure(ctx context.Context, app *board.App, scenario string) (Report, error) {
	plan, ok := scenarioPlan(scenario)
	if !ok {
		return Report{}, fmt.Errorf("dump: unknown scenario %q", scenario)
	}

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	rep := Report{Scenario: scenario, Kinds: map[string]int{}}
	if plan.draw {
		start := time.Now()
		var err error
		if plan.advance {
			err = app.Advance(ctx)
		} else {
			err = app.Redraw(ctx)
		}
		// Advance embeds its Redraw, so either branch is the draw time.
		rep.RedrawNS = int64(time.Since(start))
		if err != nil {
			return Report{}, err
		}
	}

	for i := 0; i < plan.ticks; i++ {
		pace()

		if plan.midAt > 0 && i == plan.midAt {
			start := time.Now()
			if err := app.Advance(ctx); err != nil {
				return Report{}, err
			}
			rep.RedrawNS += int64(time.Since(start))
		}

		f, err := tickFrame(ctx, app)
		if err != nil {
			return Report{}, err
		}
		rep.Frames = append(rep.Frames, f)
	}

	runtime.ReadMemStats(&after)
	rep.summarize(app, before, after)

	return rep, nil
}

// percentile returns the nearest-rank percentile p of ns (0 to 100), sorting
// ns in place; an empty slice yields 0.
func percentile(ns []int64, p int) int64 {
	if len(ns) == 0 {
		return 0
	}
	sort.Slice(ns, func(i, j int) bool { return ns[i] < ns[j] })
	rank := (p*len(ns) + 99) / 100
	if rank < 1 {
		rank = 1
	}
	return ns[rank-1]
}
