package page

import (
	"context"
	"testing"
)

// TestHoverFlowSingleCascadeNoCapture is the agnostic end-to-end pin for the
// hover readout pattern every builder app shares (inventory bins, loglens
// rows): a handler text patch plus the hover's own scoped patch, on a page
// whose theme also carries :has rules and id-less transitioning controls.
//
// It requires, per hover: one cascade total (the handler patch relays it,
// the Redraw reuses it, so StyleTime does not move), no full reparse, no
// transition capture (TransitionsNoID stays 0), no bound run, and exactly
// two scoped commits (patch plus hover patch).
func TestHoverFlowSingleCascadeNoCapture(t *testing.T) {
	source := `<style>.cell{width:40px;height:20px}` +
		`.cell:hover{outline:2px solid red}` +
		`.btn{transition: opacity 0.15s}` +
		`.m:has(:hover) .s{display:block}` +
		`#detail{width:200px;height:20px}</style>` +
		`<div class="m"><div class="s">s</div></div>` +
		`<button class="btn">go</button>` +
		`<div id="a" class="cell">a</div><div id="b" class="cell">b</div>` +
		`<div id="detail">sum</div>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	p.SetPerf(true)
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.Handle(Handlers{
		Hover: func(ctx context.Context, _, next Box) (bool, error) {
			text := "hover " + next.ID
			if err := p.PatchElement(ctx, "detail", ElementEdit{Text: &text}); err != nil {
				return false, err
			}
			return false, nil
		},
	})

	base := p.Stats()
	hoverAt(t, p, "a")
	hoverAt(t, p, "b")
	got := p.Stats()

	if got.Parses != base.Parses {
		t.Fatalf("parses = %d, want %d (a full reparse ran on hover)", got.Parses, base.Parses)
	}
	if got.Layouts != base.Layouts+4 {
		t.Fatalf("layouts = %d, want %d (want one patch plus one hover patch per hover)", got.Layouts, base.Layouts+4)
	}
	// StyleTime is per-Redraw: BeginSample zeroes it, and only a cascade
	// adds to it. The last hover Redraw must have reused the handler
	// patch's document, so it stays exactly zero.
	if got.StyleTime != 0 {
		t.Fatalf("style time = %v on the last hover redraw, want 0 (a second cascade ran)", got.StyleTime)
	}
	if got := p.Motion().TransitionsNoID; got != 0 {
		t.Fatalf("TransitionsNoID = %d, want 0 (capture ran outside every transition)", got)
	}
	if got := p.Motion().TransitionsBound; got != 0 {
		t.Fatalf("TransitionsBound = %d, want 0 (no run should start)", got)
	}
}
