package main

import (
	"context"
	"testing"
)

// TestEveryTabStaysOnTheVectorPath: a page the replay painter declines turns
// into a bitmap, and a bitmap page refuses element patches ("no cache").
// Every tab must keep a retained display list so the live demos can patch it.
func TestEveryTabStaysOnTheVectorPath(t *testing.T) {
	app, err := NewApp(catalogPath())
	if err != nil {
		t.Fatal(err)
	}

	for _, tab := range allTabs() {
		page, err := app.NewPage(context.Background(), tab.ID, 1100, 800)
		if err != nil {
			t.Fatalf("%s: %v", tab.ID, err)
		}

		if page.Probe().Display() == nil {
			t.Errorf("tab %s fell back to a bitmap", tab.ID)
		}
	}
}

// TestTopBlockIsPinned: the title bar, the header and the tab bar are fixed to
// the viewport. The section's ops stay in the scrolling flow under them.
func TestTopBlockIsPinned(t *testing.T) {
	app, err := NewApp(catalogPath())
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"shapes", "motion", "nav"} {
		page, err := app.NewPage(context.Background(), id, 1100, 800)
		if err != nil {
			t.Fatal(err)
		}

		display := page.Probe().Display()
		sticky, other := 0, 0

		for index, op := range display.Ops {
			if index >= len(display.OpOwners) || display.OpOwners[index] < 0 {
				continue
			}

			owner := display.Boxes[display.OpOwners[index]]
			if owner.ID != "top" && owner.ID != "tab-btn-"+id && owner.ID != "sec-"+id {
				continue
			}

			if owner.ID != "sec-"+id && op.Fixed {
				sticky++
			}

			if owner.ID != "sec-"+id && !op.Fixed {
				t.Errorf("%s: an op of %s is not fixed", id, owner.ID)
			}

			if owner.ID == "sec-"+id && op.Fixed {
				other++
			}
		}

		if sticky == 0 {
			t.Errorf("%s: the top block carries no fixed op", id)
		}

		if other != 0 {
			t.Errorf("%s: the section body must scroll, but %d of its ops are fixed", id, other)
		}
	}
}
