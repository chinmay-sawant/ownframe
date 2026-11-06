package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// TestFormHashParityRichDoc proves the threaded-hash sync stays
// byte-identical on a generated doc with controls, buttons, spans, and
// emoji: the published source and the sync bookkeeping match the pre-change
// goldens, and the old emoji wrapper agrees with the hashed variant.
func TestFormHashParityRichDoc(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: formSyncGoldenDoc, Width: 800, Height: 600})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(&formSyncGoldenData{Name: "Bob"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source != formSyncGoldenFinal {
		t.Fatalf("redraw source mismatch:\n%s", p.source)
	}
	if p.lastSyncBody != formSyncGoldenLastSync {
		t.Fatalf("lastSyncBody mismatch:\n%s", p.lastSyncBody)
	}

	q, err := New(Config{HTML: `<p id="m">Good 😂 and plain</p>`, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if q.source != formSyncGoldenSkip {
		t.Fatalf("skip source mismatch:\n%s", q.source)
	}

	wrapped := emojiPass(formSyncGoldenState)
	hashed, _ := draw.EmojiPassWithHash(formSyncGoldenState, form.HashSource(formSyncGoldenState))
	if wrapped != hashed || wrapped != formSyncGoldenFinal {
		t.Fatalf("emoji wrapper mismatch:\nwrapped=%s\nhashed =%s", wrapped, hashed)
	}
}

// TestFormHashOncePerChangedFrame proves the executed body is hashed once
// per changed frame and never on a clean frame, using counters not timing.
// A simple input-free, emoji-free doc keeps every stage on the same string,
// so the old path hashed it three times (sync, emoji, noteSource) and the
// new path hashes it once in the sync and reuses the hash downstream.
func TestFormHashOncePerChangedFrame(t *testing.T) {
	ctx := context.Background()
	const doc = `<p id="formhashonce">plain no emoji 12345</p>`
	p, err := New(Config{HTML: doc, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	h0 := syncHashCalls.Load()
	f0 := form.Telemetry()
	e0 := draw.EmojiStats()
	r0 := rewriteButtonCalls.Load()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.source
	if before == "" {
		t.Fatal("empty source after redraw")
	}
	if d := syncHashCalls.Load() - h0; d != 1 {
		t.Fatalf("changed frame hashes = %d, want 1", d)
	}
	if f1 := form.Telemetry(); f1.Probes != f0.Probes+1 || f1.Scans != f0.Scans+1 ||
		f1.Heads != f0.Heads || f1.Rewrites != f0.Rewrites {
		t.Fatalf("changed frame rescanned: %+v -> %+v", f0, f1)
	}
	if e1 := draw.EmojiStats(); e1.Probes != e0.Probes+1 || e1.Hits != e0.Hits {
		t.Fatalf("changed frame emoji miscounted: %+v -> %+v", e0, e1)
	}
	if r1 := rewriteButtonCalls.Load(); r1 != r0 {
		t.Fatalf("button rewrite ran on input-free body: %d -> %d", r0, r1)
	}

	p.markFormDirty()
	h1 := syncHashCalls.Load()
	f1 := form.Telemetry()
	e1 := draw.EmojiStats()
	r1 := rewriteButtonCalls.Load()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source != before {
		t.Fatalf("dirty redraw changed output:\n%s", p.source)
	}
	if d := syncHashCalls.Load() - h1; d != 1 {
		t.Fatalf("dirty frame hashes = %d, want 1", d)
	}
	if f2 := form.Telemetry(); f2 != f1 {
		t.Fatalf("dirty redraw rescanned: %+v -> %+v", f1, f2)
	}
	if e2 := draw.EmojiStats(); e2.Probes != e1.Probes || e2.Hits != e1.Hits+1 {
		t.Fatalf("dirty redraw emoji not gated: %+v -> %+v", e1, e2)
	}
	if r2 := rewriteButtonCalls.Load(); r2 != r1 {
		t.Fatalf("button rewrite ran on input-free body: %d -> %d", r1, r2)
	}

	h2 := syncHashCalls.Load()
	f2 := form.Telemetry()
	e2 := draw.EmojiStats()
	r2 := rewriteButtonCalls.Load()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source != before {
		t.Fatalf("clean redraw changed output:\n%s", p.source)
	}
	if d := syncHashCalls.Load() - h2; d != 0 {
		t.Fatalf("clean frame hashes = %d, want 0", d)
	}
	if f3 := form.Telemetry(); f3 != f2 {
		t.Fatalf("clean redraw rescanned: %+v -> %+v", f2, f3)
	}
	if e3 := draw.EmojiStats(); e3 != e2 {
		t.Fatalf("clean redraw ran emoji: %+v -> %+v", e2, e3)
	}
	if r3 := rewriteButtonCalls.Load(); r3 != r2 {
		t.Fatalf("button rewrite ran on clean frame: %d -> %d", r2, r3)
	}
}
