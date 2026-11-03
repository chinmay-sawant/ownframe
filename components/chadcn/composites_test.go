package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderCompositeTree(t *testing.T, root *ui.Element) string {
	t.Helper()
	markup, err := ui.Render(root)
	if err != nil {
		t.Fatal(err)
	}
	return markup
}

func compositeWant(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(markup, fragment) {
			t.Fatalf("missing %q in %s", fragment, markup)
		}
	}
}

func compositeDeny(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(markup, fragment) {
			t.Fatalf("unexpected %q in %s", fragment, markup)
		}
	}
}

// compositeTag returns the opening tag that contains needle.
func compositeTag(t *testing.T, markup, needle string) string {
	t.Helper()
	index := strings.Index(markup, needle)
	if index < 0 {
		t.Fatalf("missing %q in %s", needle, markup)
	}
	start := strings.LastIndex(markup[:index], "<")
	end := strings.Index(markup[index:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("malformed tag around %q", needle)
	}
	return markup[start : index+end+1]
}

func TestScrollAreaRendersViewportAndBar(t *testing.T) {
	t.Parallel()
	state := &chadcn.ScrollState{}
	state.SetRange(40, 10)
	markup := renderCompositeTree(t, chadcn.ScrollArea(chadcn.ScrollAreaOptions{
		ID:          "tags",
		Label:       "Tags",
		State:       state,
		TrackPixels: 200,
		Height:      ui.Px(240),
		Children:    []*ui.Element{ui.Text("v1.2.0-beta.50")},
	}))
	compositeWant(t, markup,
		`class="of-scope of-scroll-area"`,
		`id="tags"`,
		`role="region"`,
		`aria-label="Tags"`,
		`class="of-scroll-area-viewport"`,
		`id="tags-viewport"`,
		`style="bottom:0px;left:0px;position:absolute;right:0px;top:0px;"`,
		`style="height:240px;"`,
		`role="scrollbar"`,
		`id="tags-thumb"`,
		`aria-controls="tags-viewport"`,
		`aria-valuemax="30"`,
		`v1.2.0-beta.50`,
	)
}

func TestScrollAreaWithoutStateSkipsBar(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.ScrollArea(chadcn.ScrollAreaOptions{
		ID:       "empty",
		Children: []*ui.Element{ui.Text("Nothing to scroll")},
	}))
	compositeWant(t, markup, `class="of-scope of-scroll-area"`, `id="empty-viewport"`, `aria-label="Scroll area"`, "Nothing to scroll")
	compositeDeny(t, markup, `role="scrollbar"`, `id="empty-thumb"`, "height:240px")
}

func TestResizableSplitsPanes(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.Resizable(chadcn.ResizableOptions{
		ID:     "split",
		Ratio:  0.25,
		First:  ui.Text("One"),
		Second: ui.Text("Two"),
	}))
	compositeWant(t, markup,
		`class="of-scope of-resizable of-resizable--horizontal"`,
		`class="of-resizable-pane of-resizable-pane--first"`,
		`id="split-first"`,
		`style="flex-grow:0.25;"`,
		`class="of-resizable-pane of-resizable-pane--second"`,
		`style="flex-grow:0.75;"`,
		`role="separator"`,
		`aria-orientation="vertical"`,
		`aria-valuenow="25"`,
		`id="split-divider"`,
		`tabindex="0"`,
		`class="of-resizable-grip"`,
		"One",
		"Two",
	)
}

func TestResizableOrientationAndRatioFallback(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.Resizable(chadcn.ResizableOptions{
		Ratio:       4,
		Orientation: chadcn.ResizableColumn,
		First:       ui.Text("Top"),
		Second:      ui.Text("Bottom"),
	}))
	compositeWant(t, markup,
		`class="of-scope of-resizable of-resizable--vertical"`,
		`aria-orientation="horizontal"`,
		`aria-valuenow="50"`,
		`style="flex-grow:0.5;"`,
	)
	compositeDeny(t, markup, `id="split-divider"`)
}

func TestSidebarRendersNavItems(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.Sidebar(chadcn.SidebarOptions{
		ID:     "app-nav",
		Label:  "Application",
		Open:   true,
		Header: ui.Text("Acme Inc"),
		Content: ui.Div().Children(
			chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Playground", Active: true, Action: "nav:playground", Badge: "3"}),
			chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Settings", Icon: ui.Text("⚙")}),
		),
		Footer: ui.Text("m@example.com"),
		Width:  ui.Px(240),
	}))
	compositeWant(t, markup,
		`<nav aria-label="Application" class="of-scope of-sidebar of-sidebar--open"`,
		`data-state="open"`,
		`id="app-nav"`,
		`style="width:240px;"`,
		`class="of-sidebar-header"`,
		`class="of-sidebar-content"`,
		`class="of-sidebar-footer"`,
		`class="of-sidebar-item of-sidebar-item--active"`,
		`aria-current="page"`,
		`data-action="nav:playground"`,
		`class="of-sidebar-item-label"`,
		`class="of-sidebar-item-badge"`,
		`class="of-sidebar-item-icon"`,
		"Acme Inc",
		"m@example.com",
	)
	compositeDeny(t, markup, "of-sidebar--collapsed", "of-sidebar--closed")
}

func TestSidebarCollapsedAndClosed(t *testing.T) {
	t.Parallel()
	collapsed := renderCompositeTree(t, chadcn.Sidebar(chadcn.SidebarOptions{
		Open:      true,
		Collapsed: true,
		Content:   chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Models"}),
	}))
	compositeWant(t, collapsed, `aria-label="Sidebar"`, "of-sidebar--collapsed", "of-sidebar--open")
	closed := renderCompositeTree(t, chadcn.Sidebar(chadcn.SidebarOptions{
		Header: ui.Text("Acme Inc"),
	}))
	compositeWant(t, closed, "of-sidebar--closed", `aria-hidden="true"`, `data-state="closed"`, "Acme Inc")
}

func TestCarouselShowsOneSlide(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{
		ID:    "quotes",
		Label: "Quotes",
		Index: 1,
		Slides: []*ui.Element{
			ui.Text("alpha"),
			ui.Text("beta"),
			ui.Text("gamma"),
		},
	}))
	compositeWant(t, markup,
		`<section aria-label="Quotes" aria-roledescription="carousel" class="of-scope of-carousel"`,
		`id="quotes"`,
		`role="region"`,
		`aria-label="2 of 3"`,
		`aria-roledescription="slide"`,
		`aria-label="Previous slide"`,
		`aria-label="Next slide"`,
		`class="of-carousel-stage"`,
		`id="quotes-stage"`,
		`aria-controls="quotes-stage"`,
		"beta",
	)
	compositeDeny(t, markup, "alpha", "gamma")
	if got := strings.Count(markup, `class="of-carousel-dot"`); got != 2 {
		t.Fatalf("inactive dot count=%d want 2", got)
	}
	if got := strings.Count(markup, "of-carousel-dot--active"); got != 1 {
		t.Fatalf("active dot count=%d want 1", got)
	}
}

func TestCarouselDisablesControlsAtBounds(t *testing.T) {
	t.Parallel()
	first := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{
		Slides: []*ui.Element{ui.Text("one"), ui.Text("two")},
	}))
	compositeWant(t, first, `aria-label="1 of 2"`)
	if tag := compositeTag(t, first, `aria-label="Previous slide"`); !strings.Contains(tag, "disabled") {
		t.Fatalf("previous control at the first slide is enabled: %s", tag)
	}
	if tag := compositeTag(t, first, `aria-label="Next slide"`); strings.Contains(tag, "disabled") {
		t.Fatalf("next control at the first slide is disabled: %s", tag)
	}
	last := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{
		Index:  1,
		Slides: []*ui.Element{ui.Text("one"), ui.Text("two")},
	}))
	if tag := compositeTag(t, last, `aria-label="Previous slide"`); strings.Contains(tag, "disabled") {
		t.Fatalf("previous control at the last slide is disabled: %s", tag)
	}
	if tag := compositeTag(t, last, `aria-label="Next slide"`); !strings.Contains(tag, "disabled") {
		t.Fatalf("next control at the last slide is enabled: %s", tag)
	}
}

func TestCarouselWithoutSlides(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{Index: 3}))
	compositeWant(t, markup, `class="of-scope of-carousel"`, `aria-label="Carousel"`)
	compositeDeny(t, markup, "of-carousel-slide", "of-carousel-dot", `aria-roledescription="slide"`)
	if tag := compositeTag(t, markup, `aria-label="Next slide"`); !strings.Contains(tag, "disabled") {
		t.Fatalf("next control without slides is enabled: %s", tag)
	}
}

func TestDataTableRendersRowsAndSelection(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		ID:      "invoices",
		Caption: "Recent invoices",
		Columns: []chadcn.DataColumn{
			{Label: "Customer"},
			{Label: "Amount", Align: "right", Width: ui.Px(120)},
			{Label: "Status", Align: "center"},
		},
		Rows: [][]string{
			{"Ada Lovelace", "$316.00", "Success"},
			{"Grace Hopper", "$242.00", "Processing"},
		},
		Selected: map[int]bool{1: true},
		OnSelect: func(_ context.Context, _ int) error { return nil },
	}))
	compositeWant(t, markup,
		`class="of-scope of-data-table"`,
		`id="invoices"`,
		`class="of-data-table-table"`,
		`<caption class="of-data-table-caption">Recent invoices</caption>`,
		`class="of-data-table-header"`,
		`scope="col"`,
		`class="of-data-table-header of-data-table--align-right"`,
		`class="of-data-table-header of-data-table--align-center"`,
		`style="width:120px;"`,
		`of-data-table-select-cell`,
		`class="of-data-table-select"`,
		`class="of-data-table-select of-data-table-select--checked"`,
		`role="checkbox"`,
		`data-action=`,
		`aria-label="Select row 1"`,
		`aria-checked="false"`,
		`aria-label="Select row 2"`,
		`aria-checked="true"`,
		`id="invoices-select-1"`,
		`class="of-data-table-row of-data-table-row--selected"`,
		"Ada Lovelace",
		"Grace Hopper",
	)
}

func TestDataTableEmptyStateAndCaption(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		Columns: []chadcn.DataColumn{{Label: "Name"}, {Label: "Amount"}},
	}))
	compositeWant(t, markup,
		`class="of-data-table-cell of-data-table-empty"`,
		`colspan="2"`,
		"No results.",
	)
	compositeDeny(t, markup, "role=\"checkbox\"", "of-data-table-caption")
	withSelection := renderCompositeTree(t, chadcn.DataTable(chadcn.DataTableOptions{
		Columns:  []chadcn.DataColumn{{Label: "Name"}},
		Empty:    "No invoices.",
		OnSelect: func(_ context.Context, _ int) error { return nil },
	}))
	compositeWant(t, withSelection, "No invoices.", `colspan="2"`)
}

func TestCompositesZeroValueDefaults(t *testing.T) {
	t.Parallel()
	for name, root := range map[string]*ui.Element{
		"scroll-area": chadcn.ScrollArea(chadcn.ScrollAreaOptions{}),
		"resizable":   chadcn.Resizable(chadcn.ResizableOptions{}),
		"sidebar":     chadcn.Sidebar(chadcn.SidebarOptions{}),
		"carousel":    chadcn.Carousel(chadcn.CarouselOptions{}),
		"data-table":  chadcn.DataTable(chadcn.DataTableOptions{}),
	} {
		if root == nil {
			t.Fatalf("%s returned nil", name)
		}
		if _, err := ui.Render(root); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
