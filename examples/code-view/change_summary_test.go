package main

import "testing"

func TestChangeSummaryCounts(t *testing.T) {
	t.Parallel()

	cs := changeSet{Files: []changedFile{
		{Path: "a.go", Status: "M", Add: 3, Del: 1},
		{Path: "b.go", Status: "A", Add: 10},
		{Path: "c.go", Status: "D", Del: 7},
		{Path: "new.go", OldPath: "old.go", Status: "R", Add: 1, Del: 1},
		{Path: "notes.md", Status: "U", Add: 5},
		{Path: "copy.go", Status: "C", Binary: true},
	}}

	got := summarizeChanges(cs)
	want := changeSummary{
		Added:     1,
		Modified:  1,
		Deleted:   1,
		Renamed:   1,
		Untracked: 1,
		Other:     1,
		Total:     6,
		Add:       19,
		Del:       9,
	}

	if got != want {
		t.Fatalf("summary = %+v, want %+v", got, want)
	}
}

func TestChangeSummaryLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		summary changeSummary
		label   string
		short   string
	}{
		{
			name:    "sample diff",
			summary: summarizeChanges(sampleChangeSet()),
			label:   "7 changed · 4 modified · 2 added · 1 deleted · +562 −109",
			short:   "4M 2A 1D",
		},
		{
			name:    "one file",
			summary: summarizeChanges(changeSet{Files: []changedFile{{Path: "x.go", Status: "M", Add: 2}}}),
			label:   "1 change · 1 modified · +2 −0",
			short:   "1M",
		},
		{
			name: "every bucket",
			summary: changeSummary{
				Modified:  3,
				Added:     2,
				Deleted:   1,
				Renamed:   1,
				Untracked: 2,
				Other:     1,
				Total:     10,
				Add:       5,
				Del:       4,
			},
			label: "10 changed · 3 modified · 2 added · 1 deleted · 1 renamed · 2 untracked · 1 other · +5 −4",
			short: "3M 2A 1D 1R 2U 1?",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.summary.Label(); got != tc.label {
				t.Errorf("Label() = %q, want %q", got, tc.label)
			}

			if got := tc.summary.Short(); got != tc.short {
				t.Errorf("Short() = %q, want %q", got, tc.short)
			}
		})
	}
}

func TestChangeSummaryEmpty(t *testing.T) {
	t.Parallel()

	got := summarizeChanges(changeSet{})
	if got != (changeSummary{}) {
		t.Fatalf("empty summary = %+v", got)
	}

	if label := got.Label(); label != noChanges {
		t.Errorf("Label() = %q, want %q", label, noChanges)
	}

	if short := got.Short(); short != "" {
		t.Errorf("Short() = %q, want empty", short)
	}
}
