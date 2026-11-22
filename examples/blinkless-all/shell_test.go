package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// handCounts is how many properties each hand-built tab covers. They sum to
// 61. The generated tabs (ids starting "all-") cover the rest of the
// implemented properties.
var handCounts = map[string]int{"nav": 20, "shapes": 18, "motion": 6, "frag": 7, "read": 5, "svg": 5}

func newTestApp(t *testing.T) *App {
	t.Helper()

	app, err := NewApp(catalogPath())
	if err != nil {
		t.Fatal(err)
	}

	return app
}

func TestTabCounts(t *testing.T) {
	app := newTestApp(t)

	hand, total := 0, 0

	for _, tab := range app.tabs {
		total += tab.Count()

		if want, ok := handCounts[tab.ID]; ok {
			hand++

			if got := tab.Count(); got != want {
				t.Errorf("tab %s covers %d properties, want %d", tab.ID, got, want)
			}

			continue
		}

		if !strings.HasPrefix(tab.ID, "all-") {
			t.Fatalf("unexpected tab %q", tab.ID)
		}
	}

	if hand != len(handCounts) {
		t.Fatalf("%d hand-built tabs, want %d", hand, len(handCounts))
	}

	if want := app.cat.Count(StatusImplemented); total != want {
		t.Fatalf("tabs cover %d properties, the catalog lists %d implemented", total, want)
	}
}

// TestNoPropertyTwice checks a property is on exactly one tab, once.
func TestNoPropertyTwice(t *testing.T) {
	app := newTestApp(t)
	seen := map[string]string{}

	for _, tab := range app.tabs {
		for _, name := range tab.Names() {
			if prev, dup := seen[name]; dup {
				t.Errorf("%s is on tab %s and on tab %s", name, prev, tab.ID)
			}

			seen[name] = tab.ID
		}
	}
}

// TestOnlyOpenTabIsInThePage proves toggling hides the rest: the page for one
// tab holds that tab's section and none of the other five.
func TestOnlyOpenTabIsInThePage(t *testing.T) {
	app := newTestApp(t)

	for _, open := range app.tabs {
		doc := app.Page(open.ID)

		for _, other := range app.tabs {
			has := strings.Contains(doc, fmt.Sprintf(`id="sec-%s"`, other.ID))
			if has != (other.ID == open.ID) {
				t.Errorf("page for %s: section %s present = %v", open.ID, other.ID, has)
			}

			// Every page still has all six buttons, each with its count.
			label := fmt.Sprintf("%s (%d)", other.Title, other.Count())
			if !strings.Contains(doc, label) {
				t.Errorf("page for %s lacks the tab label %q", open.ID, label)
			}
		}
	}
}

func TestHeaderCounts(t *testing.T) {
	app := newTestApp(t)
	head := app.header()

	for _, want := range []string{
		fmt.Sprintf("lists %d CSS properties", app.cat.Total()),
		fmt.Sprintf("%d are implemented", app.cat.Count(StatusImplemented)),
		"A family is a group of properties that one example covers",
	} {
		if !strings.Contains(head, want) {
			t.Errorf("header lacks %q", want)
		}
	}
}

// TestEveryTabRenders lays each tab out through the engine and draws it.
func TestEveryTabRenders(t *testing.T) {
	app := newTestApp(t)

	for _, tab := range app.tabs {
		t.Run(tab.ID, func(t *testing.T) {
			t.Parallel()

			page, err := app.NewPage(context.Background(), tab.ID, 1000, 800)
			if err != nil {
				t.Fatalf("tab %s: %v", tab.ID, err)
			}

			if len(page.Probe().PNG()) == 0 {
				t.Errorf("tab %s drew an empty picture", tab.ID)
			}
		})
	}
}

// TestEveryRowHasItsShowcaseRightBelow checks the layout the user asked for:
// each example's showcase sits directly under its own row, before the next
// row starts, instead of one block after all the rows.
func TestEveryRowHasItsShowcaseRightBelow(t *testing.T) {
	app := newTestApp(t)

	for _, tab := range app.tabs {
		doc := app.Page(tab.ID)
		last := -1

		for _, p := range tab.Props {
			if p.Demo == nil {
				t.Errorf("tab %s: %s has no Demo", tab.ID, p.Name)

				continue
			}

			row := strings.Index(doc, fmt.Sprintf(`id="row-%s"`, p.Name))
			demo := strings.Index(doc, fmt.Sprintf(`id="demo-%s"`, p.Name))

			if row < 0 || demo < row || row < last {
				t.Errorf("tab %s: %s row at %d, demo at %d, previous row at %d", tab.ID, p.Name, row, demo, last)
			}

			if strings.TrimSpace(p.Demo()) == "" {
				t.Errorf("tab %s: %s has an empty showcase", tab.ID, p.Name)
			}

			last = row
		}
	}
}

// TestClickingATabLoadsThatTab clicks each tab button the way the window does
// and checks the page now holds that tab's section and none of the others.
func TestClickingATabLoadsThatTab(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	for _, target := range app.tabs {
		t.Run(target.ID, func(t *testing.T) {
			t.Parallel()

			page, err := app.NewPage(ctx, app.FirstTab(), 1000, 800)
			if err != nil {
				t.Fatal(err)
			}

			var x, y float64

			found := false

			for _, b := range page.Probe().Display().Boxes {
				if b.ID == "tab-btn-"+target.ID {
					x, y, found = b.X+b.W/2, b.Y+b.H/2, true
				}
			}

			if !found {
				t.Fatalf("no button for tab %s", target.ID)
			}

			if err = page.Click(ctx, x, y); err != nil {
				t.Fatalf("click %s: %v", target.ID, err)
			}

			doc := page.HTML()
			for _, other := range app.tabs {
				has := strings.Contains(doc, fmt.Sprintf(`id="sec-%s"`, other.ID))
				if has != (other.ID == target.ID) {
					t.Errorf("after clicking %s: section %s present = %v", target.ID, other.ID, has)
				}
			}

			if len(page.Probe().PNG()) == 0 { //nolint:contextcheck // Probe().PNG() takes no context
				t.Errorf("after clicking %s: empty picture", target.ID)
			}
		})
	}
}

func TestUnknownTabHandling(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if _, err := app.NewPage(ctx, "no-such-tab", 1000, 800); err == nil {
		t.Fatal("NewPage with unknown tab succeeded, want error")
	}

	doc := app.Page("no-such-tab")
	if !strings.Contains(doc, fmt.Sprintf(`id="sec-%s"`, app.FirstTab())) {
		t.Errorf("Page with unknown tab did not fall back to first tab %q", app.FirstTab())
	}
}

func TestHeaderEmptyCatalogNoNaN(t *testing.T) {
	t.Parallel()

	app := &App{cat: &Catalog{status: map[string]string{}}}
	head := app.header()
	if strings.Contains(head, "NaN") {
		t.Fatalf("empty catalog header contains NaN: %s", head)
	}
	if !strings.Contains(head, "(0.0%)") {
		t.Fatalf("empty catalog header missing (0.0%%): %s", head)
	}
}

func TestGenTabsComputeLazily(t *testing.T) {
	computed := 0
	rows := []genProp{{Name: "color", Accepted: "<color>", Sample: "#c0392b", Group: "fx"}}
	set := newGenSet(rows)
	tab := Tab{
		ID:      "all-test",
		Title:   "Test",
		Prepare: func() { set.once.Do(func() { computed++; set.compute() }) },
	}

	if computed != 0 || set.res != nil {
		t.Fatalf("genSet computed eagerly before tab render: computed=%d res=%v", computed, set.res)
	}

	tab.Prepare()
	tab.Prepare()

	if computed != 1 || len(set.res) != 1 {
		t.Fatalf("genSet did not compute once lazily: computed=%d res=%v", computed, set.res)
	}
}
