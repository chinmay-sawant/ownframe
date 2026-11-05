package chadcn_test

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func TestCarouselDotsSelectIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var chosen int
	var calls int
	view, err := ui.Mount(ownframe.Config{Width: 640, Height: 480, Theme: chadcn.Stylesheet}, chadcn.Carousel(chadcn.CarouselOptions{
		ID:    "quotes",
		Index: 0,
		Slides: []*ui.Element{
			ui.Text("alpha"),
			ui.Text("beta"),
			ui.Text("gamma"),
		},
		OnSelect: func(_ context.Context, index int) error {
			calls++
			chosen = index
			return nil
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	markup := view.Page.HTML()
	if !strings.Contains(markup, `<button`) || !strings.Contains(markup, `aria-label="Go to slide 3"`) {
		t.Fatalf("dots are not buttons: %s", markup)
	}
	if err := pagetest.Click(ctx, view.Page, "quotes-dot-2"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || chosen != 2 {
		t.Fatalf("OnSelect calls=%d index=%d", calls, chosen)
	}
}

func TestCarouselDisabledEndsDropHandlers(t *testing.T) {
	t.Parallel()
	noop := func(context.Context) error { return nil }
	slides := []*ui.Element{ui.Text("one"), ui.Text("two"), ui.Text("three")}
	first := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{Slides: slides, OnPrev: noop, OnNext: noop}))
	prev := compositeTag(t, first, `aria-label="Previous slide"`)
	next := compositeTag(t, first, `aria-label="Next slide"`)
	if !strings.Contains(prev, "disabled") || strings.Contains(prev, "data-action") {
		t.Fatalf("first previous kept a handler: %s", prev)
	}
	if strings.Contains(next, "disabled") || !strings.Contains(next, "data-action") {
		t.Fatalf("first next dropped its handler: %s", next)
	}
	last := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{Index: 2, Slides: slides, OnPrev: noop, OnNext: noop}))
	prev = compositeTag(t, last, `aria-label="Previous slide"`)
	next = compositeTag(t, last, `aria-label="Next slide"`)
	if strings.Contains(prev, "disabled") || !strings.Contains(prev, "data-action") {
		t.Fatalf("last previous dropped its handler: %s", prev)
	}
	if !strings.Contains(next, "disabled") || strings.Contains(next, "data-action") {
		t.Fatalf("last next kept a handler: %s", next)
	}
	middle := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{Index: 1, Slides: slides, OnPrev: noop, OnNext: noop}))
	for _, label := range []string{`aria-label="Previous slide"`, `aria-label="Next slide"`} {
		tag := compositeTag(t, middle, label)
		if strings.Contains(tag, "disabled") || !strings.Contains(tag, "data-action") {
			t.Fatalf("middle control: %s", tag)
		}
	}
	empty := renderCompositeTree(t, chadcn.Carousel(chadcn.CarouselOptions{OnPrev: noop, OnNext: noop}))
	for _, label := range []string{`aria-label="Previous slide"`, `aria-label="Next slide"`} {
		tag := compositeTag(t, empty, label)
		if !strings.Contains(tag, "disabled") || strings.Contains(tag, "data-action") {
			t.Fatalf("empty control kept a handler: %s", tag)
		}
	}
}

func TestMessageScrollerAppliesScrollOffset(t *testing.T) {
	t.Parallel()
	var state chadcn.ScrollState
	state.SetRange(900, 300)
	if !state.ScrollTo(160) {
		t.Fatal("offset did not move")
	}
	markup, err := ui.Render(chadcn.MessageScroller(chadcn.MessageScrollerOptions{
		Height:   ui.Px(300),
		State:    &state,
		Children: []*ui.Element{ui.Text("older"), ui.Text("newer")},
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`class="of-message-scroller-content"`,
		"translateY(-160px)",
		"height:300px",
		"older",
		"newer",
	} {
		if !strings.Contains(markup, want) {
			t.Fatalf("missing %q in %s", want, markup)
		}
	}
	state.ScrollTo(0)
	markup, err = ui.Render(chadcn.MessageScroller(chadcn.MessageScrollerOptions{
		State:    &state,
		Children: []*ui.Element{ui.Text("older")},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(markup, "translateY") || strings.Contains(markup, "style=") {
		t.Fatalf("zero offset should not shift or style the viewport: %s", markup)
	}
}

func TestMessageScrollerStylesheetDoesNotScroll(t *testing.T) {
	t.Parallel()
	const rule = ".of-message-scroller-viewport {"
	start := strings.Index(chadcn.Stylesheet, rule)
	if start < 0 {
		t.Fatal("missing viewport rule")
	}
	end := strings.Index(chadcn.Stylesheet[start:], "}")
	if end < 0 {
		t.Fatal("viewport rule did not close")
	}
	block := chadcn.Stylesheet[start : start+end]
	if strings.Contains(block, "overflow-y: auto") || strings.Contains(block, "overflow: auto") {
		t.Fatalf("viewport pretends to scroll: %s", block)
	}
	if !strings.Contains(block, "overflow: hidden") {
		t.Fatalf("viewport does not clip: %s", block)
	}
}

func TestResizableRatioAtAndDividerID(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.Resizable(chadcn.ResizableOptions{
		ID:     "split",
		Ratio:  0.25,
		First:  ui.Text("One"),
		Second: ui.Text("Two"),
	}))
	compositeWant(t, markup, `id="split-divider"`, `role="separator"`)
	for _, tt := range []struct {
		name                  string
		pointer, origin, span float64
		want                  float64
	}{
		{"quarter", 50, 0, 200, 0.25},
		{"half from origin", 140, 40, 200, 0.5},
		{"low clamp", -10, 0, 200, 0.1},
		{"high clamp", 190, 0, 200, 0.9},
		{"nan", math.NaN(), 0, 200, 0.5},
		{"zero span", 10, 0, 0, 0.5},
	} {
		got := chadcn.ResizableRatioAt(tt.pointer, tt.origin, tt.span)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Fatalf("%s: ratio=%v want=%v", tt.name, got, tt.want)
		}
	}
	var got float64
	options := chadcn.ResizableOptions{
		OnChange: func(_ context.Context, ratio float64) error {
			got = ratio
			return nil
		},
	}
	ratio := chadcn.ResizableRatioAt(50, 0, 200)
	if err := options.OnChange(context.Background(), ratio); err != nil {
		t.Fatal(err)
	}
	if got != ratio {
		t.Fatalf("OnChange got %v want %v", got, ratio)
	}
}

func TestScrollAreaHonorsOffsetAndOnChange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var state chadcn.ScrollState
	state.SetRange(1000, 200)
	var view *ui.View
	var tree func() *ui.Element
	tree = func() *ui.Element {
		return chadcn.ScrollArea(chadcn.ScrollAreaOptions{
			ID:          "tags",
			State:       &state,
			TrackPixels: 200,
			Height:      ui.Px(200),
			OnChange: func(ctx context.Context) error {
				return view.Update(ctx, tree())
			},
			Children: []*ui.Element{ui.Text("row")},
		})
	}
	markup, err := ui.Render(tree())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(markup, "translateY") {
		t.Fatalf("zero offset translated content: %s", markup)
	}
	if !state.ScrollTo(80) {
		t.Fatal("offset did not move")
	}
	markup, err = ui.Render(tree())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(markup, "translateY(-80px)") || !strings.Contains(markup, `id="tags-bar-after"`) {
		t.Fatalf("offset or track missing: %s", markup)
	}
	view, err = ui.Mount(ownframe.Config{Width: 400, Height: 320, Theme: theme.Dark + chadcn.Stylesheet}, tree())
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "tags-bar-after"); err != nil {
		t.Fatal(err)
	}
	if state.Offset() != 280 {
		t.Fatalf("track click offset=%d want 280", state.Offset())
	}
}

func TestScrollTrackPlacesThumb(t *testing.T) {
	t.Parallel()
	scrolling := chadcn.ScrollMetrics{Content: 1000, Viewport: 200}
	for _, tt := range []struct {
		name    string
		metrics chadcn.ScrollMetrics
		localY  float64
		track   int
		want    int
	}{
		{"start", scrolling, 0, 200, 0},
		{"center", scrolling, 100, 200, 400},
		{"end", scrolling, 200, 200, 800},
		{"past end", scrolling, 240, 200, 800},
		{"negative", scrolling, -20, 200, 0},
		{"nan", scrolling, math.NaN(), 200, 0},
		{"no track", scrolling, 10, 0, 0},
		{"fits", chadcn.ScrollMetrics{Content: 100, Viewport: 100}, 10, 100, 0},
	} {
		if got := tt.metrics.OffsetAt(tt.localY, tt.track); got != tt.want {
			t.Fatalf("%s: offset=%d want=%d", tt.name, got, tt.want)
		}
	}
	var state chadcn.ScrollState
	state.SetRange(1000, 200)
	if !state.Track(100, 200) || state.Offset() != 400 {
		t.Fatalf("track offset=%d", state.Offset())
	}
	if state.Track(100, 200) {
		t.Fatal("repeat track moved")
	}
}
