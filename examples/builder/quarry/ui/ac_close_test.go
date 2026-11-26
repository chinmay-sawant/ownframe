package ui

import (
	"strings"
	"testing"
)

func (r *rig) openPopover() {
	r.t.Helper()
	r.focusEditor()
	r.typ("SELECT * FROM us")
	if !r.a.ac.open || !r.has("ac-popover") {
		r.t.Fatal("popover should be open")
	}
}

func TestAutocompleteClosesOnRunExplainTabEscapeBlurAndCaretMove(t *testing.T) {
	cases := map[string]func(r *rig){
		"run":      func(r *rig) { r.tap("ed-run") },
		"explain":  func(r *rig) { r.tap("ed-explain") },
		"newtab":   func(r *rig) { r.tap("tab-new") },
		"escape":   func(r *rig) { r.key("escape") },
		"blur":     func(r *rig) { r.tap("sb-title") },
		"caret":    func(r *rig) { r.key("home") },
		"f5":       func(r *rig) { r.key("f5") },
		"overflow": func(r *rig) { r.tap("ov-more-ed") },
		"ctrlente": func(r *rig) { r.chord("enter", true) },
	}
	for name, act := range cases {
		t.Run(name, func(t *testing.T) {
			r := newRig(t, 1440, 960, "")
			r.openPopover()
			act(r)
			r.settle()
			if r.a.ac.open || r.has("ac-popover") {
				t.Fatalf("%s must close the autocomplete popover", name)
			}
		})
	}
}

func TestSQLiteMonitorShowsSizeChartInsteadOfZeroCacheSeries(t *testing.T) {
	for _, dark := range []bool{true, false} {
		r := newRig(t, 1440, 960, "")
		if !dark {
			r.tap("tb-theme")
		}
		r.openMonitor()
		txt := r.texts()
		if strings.Contains(txt, "Cache Hit Ratio (%)") {
			t.Fatal("SQLite cannot report a cache hit ratio; the card must not claim one")
		}
		if !strings.Contains(txt, "Database Size (MB)") || !r.has("chart-cache") {
			t.Fatal("SQLite shows database size in that slot")
		}
	}
}
