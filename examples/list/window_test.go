package list

import "testing"

func TestWindowClamps(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                        string
		offsetY, viewH, gridY, rowH int
		total, over                 int
		start, count, top, bottom   int
	}{
		{
			name:   "top",
			viewH:  50,
			rowH:   10,
			total:  20,
			over:   2,
			count:  10,
			bottom: 100,
		},
		{
			name:    "mid",
			offsetY: 80,
			viewH:   50,
			rowH:    10,
			total:   20,
			over:    2,
			start:   6,
			count:   10,
			top:     60,
			bottom:  40,
		},
		{
			name:    "bottom",
			offsetY: 150,
			viewH:   50,
			rowH:    10,
			total:   20,
			over:    2,
			start:   10,
			count:   10,
			top:     100,
		},
		{
			name:    "past end",
			offsetY: 500,
			viewH:   50,
			rowH:    10,
			total:   20,
			over:    2,
			start:   10,
			count:   10,
			top:     100,
		},
		{
			name:  "empty",
			viewH: 50,
			rowH:  10,
			over:  2,
		},
		{
			name:  "zero row height",
			viewH: 50,
			total: 20,
			over:  2,
		},
		{
			name:   "no overscan",
			viewH:  50,
			rowH:   10,
			total:  20,
			count:  6,
			bottom: 140,
		},
		{
			name:    "grid origin",
			offsetY: 40,
			viewH:   50,
			gridY:   40,
			rowH:    10,
			total:   20,
			over:    2,
			count:   10,
			bottom:  100,
		},
		{
			name:    "above grid",
			offsetY: 0,
			viewH:   50,
			gridY:   40,
			rowH:    10,
			total:   20,
			over:    2,
			count:   10,
			bottom:  100,
		},
		{
			name:  "overscan larger than list",
			viewH: 50,
			rowH:  10,
			total: 3,
			over:  8,
			count: 3,
		},
		{
			name:   "negative overscan",
			viewH:  50,
			rowH:   10,
			total:  20,
			over:   -4,
			count:  6,
			bottom: 140,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Window(tc.offsetY, tc.viewH, tc.gridY, tc.rowH, tc.total, tc.over)
			if got.Start != tc.start || got.Count != tc.count {
				t.Fatalf("start/count = %d/%d, want %d/%d", got.Start, got.Count, tc.start, tc.count)
			}
			if got.Top != tc.top || got.Bottom != tc.bottom {
				t.Fatalf("spacers = %d/%d, want %d/%d", got.Top, got.Bottom, tc.top, tc.bottom)
			}
			end := got.Start + got.Count
			if got.Start < 0 || got.Count < 0 || end > tc.total {
				t.Fatalf("range %d+%d outside [0,%d]", got.Start, got.Count, tc.total)
			}
		})
	}
}
