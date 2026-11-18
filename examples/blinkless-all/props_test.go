package main

import (
	"strings"
	"testing"
)

// TestEveryGenPropHasItsOwnSpec: a generated row must have a demo built for
// its property, not the plain fallback box.
func TestEveryGenPropHasItsOwnSpec(t *testing.T) {
	taken := handNames()

	for _, r := range genProps() {
		if taken[r.Name] {
			continue
		}

		if specFor(r).HTML == defaultSpecHTML {
			t.Errorf("%s (%s) has no demo spec", r.Name, r.Group)
		}
	}
}

// TestVisibleSpecsChangeTheLayout: a spec that says Visible must show a
// difference between the control and the demo in the engine's own output.
func TestVisibleSpecsChangeTheLayout(t *testing.T) {
	taken := handNames()
	byGroup := map[string][]genProp{}

	for _, r := range genProps() {
		if !taken[r.Name] {
			byGroup[r.Group] = append(byGroup[r.Group], r)
		}
	}

	for group, rows := range byGroup {
		set := newGenSet(rows)
		set.compute()

		for _, r := range rows {
			if !specFor(r).Visible {
				continue
			}

			if res := set.res[r.Name]; !strings.Contains(res, "layout or paint differs") {
				t.Errorf("%s (%s): marked Visible but %s", r.Name, group, res)
			}
		}
	}
}

func handNames() map[string]bool {
	taken := map[string]bool{}

	for _, tab := range []Tab{navTab(), shapesTab(), motionTab(), fragTab(), readTab(), svgTab()} {
		for _, n := range tab.Names() {
			taken[n] = true
		}
	}

	return taken
}
