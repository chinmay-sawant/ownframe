package insights

import (
	"fmt"
	"html/template"
)

// monthsFor returns the month labels that overlap a window starting at
// start. A band clipped to fewer than three columns is skipped, so the
// first and last partial months stay unlabeled.
func monthsFor(start int) []Month {
	out := make([]Month, 0, len(streakMonths))
	col := 0

	for _, band := range streakMonths {
		lo, hi := col, col+band.Weeks-1
		col = hi + 1

		visLo, visHi := max(lo, start), min(hi, start+streakWindow-1)
		if visHi-visLo < minSpan {
			continue
		}

		out = append(out, Month{
			Label:  band.Label,
			Column: template.CSS(fmt.Sprintf("%d / span %d", visLo-start+1, visHi-visLo+1)), //nolint:gosec // reason: column is two internally formatted integers, no user input reaches the CSS.
		})
	}

	return out
}
