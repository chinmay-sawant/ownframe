package page

import (
	"context"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// hoverBoxWidth returns the CSS-pixel width of the box with the given id.
func hoverBoxWidth(t *testing.T, p *Page, id string) float64 {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			return b.W
		}
	}

	t.Fatalf("no box %q", id)

	return 0
}

// TestStyleTriggerHoverSkipsCascade proves a hover on a hoverless sheet
// skips the style stage: no parse, no cascade, no layout, and no repaint.
// Counters prove it, not timing.
func TestStyleTriggerHoverSkipsCascade(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#e{width:100px;height:20px}</style></head><body><div id="e">x</div></body></html>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := p.Stats()
	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats(); !reflect.DeepEqual(got, before) {
		t.Fatalf("hover restyled a hoverless sheet: %+v -> %+v", before, got)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	after := p.Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("redraw recascaded a hoverless sheet: %+v -> %+v", before, after)
	}
	if after.Layouts != before.Layouts || after.Repaints != before.Repaints {
		t.Fatalf("redraw relaid a hoverless sheet: %+v -> %+v", before, after)
	}
}

// TestStyleTriggerStateNoopReusesDocument proves the style fast path: a
// redraw that only moves state the sheets cannot match reuses the cached
// styled document instead of re-cascading. A move on a used pseudo still
// takes the full relayout.
func TestStyleTriggerStateNoopReusesDocument(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#e{width:100px;height:20px}#e:hover{width:120px}</style></head><body><div id="e">x</div></body></html>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.statePseudos.Hover {
		t.Fatal("sheets should use :hover")
	}

	cached := p.cache.Styled()
	if cached == nil {
		t.Fatal("no cached document")
	}

	// Focus is unused by these sheets, so moving it must not restyle.
	focused := p.renderState()
	focused.Focus = "e"
	doc, ok := p.styleNoop(focused)
	if !ok {
		t.Fatal("unused focus move was not a style noop")
	}
	if doc != cached {
		t.Fatal("unused focus move replaced the styled document")
	}
	if routed, err := p.styleStaged(ctx, false, p.source, focused); err != nil {
		t.Fatal(err)
	} else if routed != cached {
		t.Fatal("styleStaged re-cascaded an unused focus move")
	}

	// Hover is used, so moving it must restyle.
	hovered := p.renderState()
	hovered.Hover = "e"
	if _, ok := p.styleNoop(hovered); ok {
		t.Fatal("used hover move skipped the restyle")
	}
	if doc, err := p.styleStaged(ctx, false, p.source, hovered); err != nil {
		t.Fatal(err)
	} else if doc == cached {
		t.Fatal("used hover move reused the styled document")
	}

	// The noop leaves the cascade counters alone.
	if got := p.Stats(); got.Parses != 1 || got.Cascades != 1 {
		t.Fatalf("noop restyle moved the cascade counters: %+v", got)
	}
}

// TestStyleTriggerThemeSwapApplies proves a same-source theme swap paints
// the new sheet even when the style cache is warm: the swap takes a fresh
// Apply instead of the cacheFresh relayout, which would keep the old extra
// sheet. It also refreshes which pseudos the page reports.
func TestStyleTriggerThemeSwapApplies(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#swatch{width:100px;height:40px}</style></head><body><div id="swatch"></div></body></html>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := hoverBoxWidth(t, p, "swatch"); got != 100 {
		t.Fatalf("plain width = %v, want 100", got)
	}
	if p.statePseudos.Hover {
		t.Fatal("plain sheets should not use :hover")
	}

	// Swap the theme without dropping the parsed tree, the way a
	// tree-preserving SetTheme would. The next redraw must still paint it.
	sheet, err := css.Parse(`#swatch{width:150px}#swatch:hover{width:160px}`)
	if err != nil {
		t.Fatal(err)
	}
	p.theme = sheet
	p.themeSrc = `#swatch{width:150px}#swatch:hover{width:160px}`

	before := p.Stats()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := hoverBoxWidth(t, p, "swatch"); got != 150 {
		t.Fatalf("themed width = %v, want 150", got)
	}
	if !p.statePseudos.Hover {
		t.Fatal("theme :hover was not reported after the swap")
	}
	if got := p.Stats(); got.Cascades == before.Cascades {
		t.Fatalf("theme swap skipped the cascade: %+v", got)
	}
}

// TestStyleTriggerThemeParity proves the SetTheme path paints exactly what a
// full rebuild with the same themed document paints.
func TestStyleTriggerThemeParity(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#swatch{width:100px;height:40px}</style></head><body><div id="swatch"></div></body></html>`
	theme := `#swatch{width:150px;background:#00ff00}`

	ctx := context.Background()

	swapped, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := swapped.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := swapped.SetTheme(theme); err != nil {
		t.Fatal(err)
	}
	if err := swapped.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	built, err := New(Config{HTML: source, Theme: theme, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := built.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got, want := hoverBoxWidth(t, swapped, "swatch"), hoverBoxWidth(t, built, "swatch"); got != want {
		t.Fatalf("swapped width = %v, rebuilt width = %v", got, want)
	}
}

// TestStyleTriggerKeyComparisons pins the redraw key behavior the style
// fast path relies on: only real input changes refuse the skip.
func TestStyleTriggerKeyComparisons(t *testing.T) {
	t.Parallel()

	view := draw.Viewport{Width: 320, Height: 200}
	state := render.State{Hover: "e"}
	pseudos := render.StatePseudos{Hover: true}

	last := draw.KeyForHashed("a", draw.HashString("a"), 1, 1, view, state, pseudos, 0)
	if !draw.Unchanged(true, last, last) {
		t.Fatal("identical keys report changed")
	}
	if draw.Unchanged(false, last, last) {
		t.Fatal("missing last frame reports unchanged")
	}

	same := last
	if !draw.Unchanged(true, last, same) {
		t.Fatal("equal keys report changed")
	}

	for name, mutate := range map[string]func(*draw.Key){
		"width":  func(k *draw.Key) { k.Width++ },
		"height": func(k *draw.Key) { k.Height++ },
		"hover":  func(k *draw.Key) { k.Hover = "other" },
		"focus":  func(k *draw.Key) { k.Focus = "f" },
		"active": func(k *draw.Key) { k.Active = "a" },
		"images": func(k *draw.Key) { k.Images++ },
		"rev":    func(k *draw.Key) { k.SourceRev++ },
		"hash":   func(k *draw.Key) { k.SourceHash++ },
		"len":    func(k *draw.Key) { k.SourceLen++ },
	} {
		next := last
		mutate(&next)
		if draw.Unchanged(true, last, next) {
			t.Fatalf("%s change reports unchanged", name)
		}
	}

	// Equal revision, hash, and length never memcmp the source: different
	// Source bytes with the same identity still report unchanged.
	collision := last
	collision.Source = "b"
	if len(collision.Source) != collision.SourceLen {
		t.Fatal("fixture sources must share one length")
	}
	collision.SourceHash = last.SourceHash
	if !draw.Unchanged(true, last, collision) {
		t.Fatal("revision-identical keys memcmp the source")
	}

	// Without revisions, different sources still report changed.
	a := draw.KeyFor("aaa", view, render.State{}, render.StatePseudos{}, 0)
	b := draw.KeyFor("bbb", view, render.State{}, render.StatePseudos{}, 0)
	if draw.Unchanged(true, a, b) {
		t.Fatal("different sources report unchanged")
	}
	if !draw.Unchanged(true, a, a) {
		t.Fatal("identical sourceless keys report changed")
	}

	// An unmatched state drops out of the key, so hovering a hoverless
	// sheet never changes the frame.
	bare := draw.KeyForHashed("a", draw.HashString("a"), 1, 1, view,
		render.State{Hover: "e"}, render.StatePseudos{}, 0)
	plain := draw.KeyForHashed("a", draw.HashString("a"), 1, 1, view,
		render.State{}, render.StatePseudos{}, 0)
	if !draw.Unchanged(true, bare, plain) {
		t.Fatal("unused hover state changes the key")
	}
}
