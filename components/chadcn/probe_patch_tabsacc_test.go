package chadcn_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeTabsAccRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

// TestProbePatchTabsAcc proves open->open content edits on tabs, accordion,
// collapsible, and carousel take the ui.View.Update patch path (parses+0).
// Host content carries its own stable leaf ids; the tab triggers/panels,
// accordion triggers/panels/titles, collapsible content, and carousel stage
// and dots already derive theirs from host/item IDs, and the carousel
// controls gained ID+"-prev"/"-next" so they no longer block a patch.
func TestProbePatchTabsAcc(t *testing.T) {
	t.Parallel()

	t.Run("TabsPanelEdit", func(t *testing.T) {
		t.Parallel()
		build := func(body string) *ui.Element {
			return probeTabsAccRoot(chadcn.Tabs(chadcn.TabsOptions{
				Active: "tab-a",
				Items: []chadcn.TabItem{
					{ID: "tab-a", Label: "Alpha"},
					{ID: "tab-b", Label: "Beta"},
				},
				Render: func(id string) *ui.Element {
					if id == "tab-a" {
						return ui.P().ID("pp-tabs-panel").Text(body)
					}
					return ui.P().ID("pp-tabs-other").Text("other")
				},
			}))
		}
		v, ctx := mountProbeView(t, build("panel v1"))
		expectProbePatch(t, v, ctx, build("panel v2"))
	})

	t.Run("AccordionPanelEdit", func(t *testing.T) {
		t.Parallel()
		build := func(body string) *ui.Element {
			return probeTabsAccRoot(chadcn.Accordion(chadcn.AccordionOptions{
				Open: []string{"a1"},
				Items: []chadcn.AccordionItem{
					{ID: "a1", Title: "One", Content: ui.P().ID("pp-acc-a1").Text(body)},
					{ID: "a2", Title: "Two", Content: ui.P().ID("pp-acc-a2").Text("c2")},
				},
			}))
		}
		v, ctx := mountProbeView(t, build("c1 v1"))
		expectProbePatch(t, v, ctx, build("c1 v2"))
	})

	t.Run("AccordionTitleEdit", func(t *testing.T) {
		t.Parallel()
		// Same-length titles: a longer title that resizes the trigger is
		// refused by the engine and falls back to FULL, which is engine
		// behavior, not a missing id.
		build := func(title string) *ui.Element {
			return probeTabsAccRoot(chadcn.Accordion(chadcn.AccordionOptions{
				Open: []string{"a1"},
				Items: []chadcn.AccordionItem{
					{ID: "a1", Title: title, Content: ui.P().ID("pp-acc-a1").Text("c1")},
					{ID: "a2", Title: "Two", Content: ui.P().ID("pp-acc-a2").Text("c2")},
				},
			}))
		}
		v, ctx := mountProbeView(t, build("One"))
		expectProbePatch(t, v, ctx, build("Two"))
	})

	t.Run("CollapsibleBodyEdit", func(t *testing.T) {
		t.Parallel()
		build := func(body string) *ui.Element {
			return probeTabsAccRoot(chadcn.Collapsible(chadcn.CollapsibleOptions{
				ID:      "pp-col",
				Open:    true,
				Trigger: chadcn.Button("More", chadcn.ButtonOptions{}).ID("pp-col-trigger"),
			}, ui.P().ID("pp-col-body").Text(body)))
		}
		v, ctx := mountProbeView(t, build("body v1"))
		expectProbePatch(t, v, ctx, build("body v2"))
	})

	t.Run("CarouselSlideEdit", func(t *testing.T) {
		t.Parallel()
		build := func(body string) *ui.Element {
			return probeTabsAccRoot(chadcn.Carousel(chadcn.CarouselOptions{
				ID:    "pp-car",
				Label: "Shots",
				Index: 1,
				Slides: []*ui.Element{
					ui.P().ID("pp-car-s0").Text("s0"),
					ui.P().ID("pp-car-s1").Text(body),
					ui.P().ID("pp-car-s2").Text("s2"),
				},
			}))
		}
		v, ctx := mountProbeView(t, build("s1 v1"))
		expectProbePatch(t, v, ctx, build("s1 v2"))
	})

	t.Run("TabsIdenticalNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return probeTabsAccRoot(chadcn.Tabs(chadcn.TabsOptions{
				Active: "tab-a",
				Items: []chadcn.TabItem{
					{ID: "tab-a", Label: "Alpha", Content: ui.P().ID("pp-tabs-a").Text("a")},
					{ID: "tab-b", Label: "Beta", Content: ui.P().ID("pp-tabs-b").Text("b")},
				},
			}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})
}

// TestProbePatchTabsAccStructuralFull documents the transitions that stay
// FULL by design: tab switches, accordion open-set changes, collapsible and
// carousel mounts, and pagination page changes all mount different nodes or
// remove attributes, which the diff cannot patch. Tabs and accordion panels
// mount lazily (only the active/open panel renders), so switching the active
// set always reparses.
// requireProbeFull fails when the Update does not take the full-render
// fallback: exactly one generation and at least one reparse.
func requireProbeFull(t *testing.T, v *ui.View, ctx context.Context, next *ui.Element) {
	t.Helper()
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, next); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	if got := v.Page.Probe().Generation(); got != gen0+1 {
		t.Fatalf("generation = %d, want %d", got, gen0+1)
	}
	if s1.Parses <= s0.Parses {
		t.Fatalf("expected full render: parses %d -> %d", s0.Parses, s1.Parses)
	}
}

func TestProbePatchTabsAccStructuralFull(t *testing.T) {
	t.Parallel()

	t.Run("TabsSwitch", func(t *testing.T) {
		t.Parallel()
		build := func(active string) *ui.Element {
			return probeTabsAccRoot(chadcn.Tabs(chadcn.TabsOptions{
				Active: active,
				Items: []chadcn.TabItem{
					{ID: "tab-a", Label: "Alpha"},
					{ID: "tab-b", Label: "Beta"},
				},
				Render: func(id string) *ui.Element {
					return ui.P().ID("pp-tabsfull-" + id).Text("panel " + id)
				},
			}))
		}
		v, ctx := mountProbeView(t, build("tab-a"))
		requireProbeFull(t, v, ctx, build("tab-b"))
	})

	t.Run("AccordionOpenChange", func(t *testing.T) {
		t.Parallel()
		build := func(open ...string) *ui.Element {
			return probeTabsAccRoot(chadcn.Accordion(chadcn.AccordionOptions{
				Open: open,
				Items: []chadcn.AccordionItem{
					{ID: "a1", Title: "One", Content: ui.P().ID("pp-accfull-a1").Text("c1")},
					{ID: "a2", Title: "Two", Content: ui.P().ID("pp-accfull-a2").Text("c2")},
				},
			}))
		}
		v, ctx := mountProbeView(t, build("a1"))
		requireProbeFull(t, v, ctx, build("a1", "a2"))
	})

	t.Run("CollapsibleToggle", func(t *testing.T) {
		t.Parallel()
		build := func(open bool) *ui.Element {
			return probeTabsAccRoot(chadcn.Collapsible(chadcn.CollapsibleOptions{
				ID:      "pp-colfull",
				Open:    open,
				Trigger: chadcn.Button("More", chadcn.ButtonOptions{}).ID("pp-colfull-trigger"),
			}, ui.P().ID("pp-colfull-body").Text("body")))
		}
		v, ctx := mountProbeView(t, build(false))
		requireProbeFull(t, v, ctx, build(true))
	})

	t.Run("CarouselIndex", func(t *testing.T) {
		t.Parallel()
		build := func(index int) *ui.Element {
			return probeTabsAccRoot(chadcn.Carousel(chadcn.CarouselOptions{
				ID:    "pp-carfull",
				Label: "Shots",
				Index: index,
				Slides: []*ui.Element{
					ui.P().ID("pp-carfull-s0").Text("s0"),
					ui.P().ID("pp-carfull-s1").Text("s1"),
					ui.P().ID("pp-carfull-s2").Text("s2"),
				},
			}))
		}
		v, ctx := mountProbeView(t, build(0))
		requireProbeFull(t, v, ctx, build(1))
	})

	// PaginationPage stays FULL: PaginationOptions carries no host ID, so
	// page-number ids would collide across instances and are not invented.
	// Even with ids, a page change removes aria-current/disabled attributes,
	// which the diff cannot patch, and wide ranges change the slot child
	// count. FULL is by design here.
	t.Run("PaginationPage", func(t *testing.T) {
		t.Parallel()
		build := func(page int) *ui.Element {
			return probeTabsAccRoot(chadcn.Pagination(chadcn.PaginationOptions{
				Page: page, TotalPages: 5,
			}))
		}
		v, ctx := mountProbeView(t, build(1))
		requireProbeFull(t, v, ctx, build(2))
	})
}
