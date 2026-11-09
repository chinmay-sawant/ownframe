package page

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
)

func TestAdoptPatchedSourceKeepsHTMLInStep(t *testing.T) {
	t.Parallel()

	oldSource := builderDocHead + "Adopt" + builderDocMid +
		`<div id="x">old</div>` + builderDocTail
	newSource := builderDocHead + "Adopt" + builderDocMid +
		`<div id="x">new</div>` + builderDocTail

	p, err := New(Config{HTML: oldSource, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	doc := p.nav.Doc()

	text := "new"
	if err := p.PatchElement(ctx, "x", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	if err := AdoptPatchedSource(p, newSource); err != nil {
		t.Fatal(err)
	}
	// The adopt path hashes the source once and hands the same value to
	// history and to the source memo, so every recorded hash agrees with
	// a fresh fold of the adopted bytes.
	wantHash := draw.HashString(newSource)
	if got := p.HTML(); got != newSource {
		t.Fatalf("HTML() = %q, want the adopted source", got)
	}
	if got := p.source; got != newSource {
		t.Fatalf("source = %q, want the adopted source", got)
	}
	if got := p.lastSyncBody; got != newSource {
		t.Fatalf("lastSyncBody = %q, want the adopted source", got)
	}
	if got := p.nav.Current(); got != newSource {
		t.Fatalf("nav current = %q, want the adopted source", got)
	}
	if got := p.nav.CurrentHash(); got != nav.Hash(newSource) || got != wantHash {
		t.Fatalf("nav hash = %d, want %d", got, wantHash)
	}
	if got := p.nav.CurrentDelim(); got != strings.Contains(newSource, "{{") {
		t.Fatalf("nav delim = %v, want %v", got, strings.Contains(newSource, "{{"))
	}
	if got := p.sourceHash; got != wantHash {
		t.Fatalf("source hash = %d, want %d", got, wantHash)
	}
	if got := p.lastBody; got.source != newSource || got.hash != wantHash {
		t.Fatalf("body memo = (%q, %d), want (%q, %d)", got.source, got.hash, newSource, wantHash)
	}
	if p.nav.Doc() != doc {
		t.Fatalf("adopt moved history: doc %d, want %d", p.nav.Doc(), doc)
	}

	// The remembered key covers the adopted source, so the next redraw is
	// the unchanged fast path and does not lay the page out again.
	if got := p.last.SourceHash; got != wantHash || p.last.SourceLen != len(newSource) ||
		p.last.SourceRev != p.sourceRev {
		t.Fatalf("render key source = (%d, %d, %d), want (%d, %d, %d)",
			got, p.last.SourceLen, p.last.SourceRev, wantHash, len(newSource), p.sourceRev)
	}
	if got := p.last; got != p.keyFor(p.renderState()) {
		t.Fatal("render key does not match the adopted source")
	}
	layouts := p.Stats().Layouts
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != layouts {
		t.Fatalf("post-adopt redraw relaid out: %d -> %d", layouts, got)
	}
	if got := p.HTML(); got != newSource {
		t.Fatalf("HTML() after redraw = %q, want the adopted source", got)
	}

	// Adopting the same bytes again keeps the source revision and the
	// render key: the shared hash proves the source did not move.
	rev, key := p.sourceRev, p.last
	if err := AdoptPatchedSource(p, newSource); err != nil {
		t.Fatal(err)
	}
	if p.sourceRev != rev {
		t.Fatalf("re-adopt bumped source revision %d -> %d", rev, p.sourceRev)
	}
	if p.last != key {
		t.Fatal("re-adopt changed the render key for identical bytes")
	}
}

func TestAdoptPatchedSourceRefusesHandWrittenHTML(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<p>plain</p>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := AdoptPatchedSource(p, `<p>other</p>`); !errors.Is(err, ErrAdoptSource) {
		t.Fatalf("err = %v, want %v", err, ErrAdoptSource)
	}
	if got := p.HTML(); got != `<p>plain</p>` {
		t.Fatalf("HTML() = %q, want the original source", got)
	}
}
