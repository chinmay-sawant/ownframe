package main

import (
	"strings"
	"testing"
	"time"
)

// TestCatalogStartRunsEachPreviewKind clicks one card of each kind and checks
// the app records the state and the runtime takes the preview.
func TestCatalogStartRunsEachPreviewKind(t *testing.T) {
	for _, name := range []string{"-webkit-animation", "transition-property", "animation-composition"} {
		t.Run(name, func(t *testing.T) {
			p, d, v := fixturePage(t)
			row := catalogRowByName(t, v, name)
			clickID(t, p, row.Button)
			if !v.Started[row.Button] || v.Selected != name {
				t.Fatalf("start state not recorded: %+v", v)
			}
			if !strings.Contains(v.Readout, row.Value) {
				t.Fatalf("readout = %q, want the card value %q", v.Readout, row.Value)
			}
			if row.SampleKey == "" {
				return
			}
			time.Sleep(5 * time.Millisecond)
			tickPage(t, p)
			if got := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel); got == 0 {
				manual := d.manual[row.SampleKey]
				if manual == nil || !manual.active {
					t.Fatalf("%s started neither an automatic nor a manual preview", name)
				}
			}
		})
	}
}
