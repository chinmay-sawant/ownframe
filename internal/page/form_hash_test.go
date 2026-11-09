package page

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
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

// TestSourceHashFoldParity proves the FNV-1a fold is identical in the three
// places a source hash crosses: the draw key, the nav history, and the form
// sync memo, so one computed value can be shared by all of them.
func TestSourceHashFoldParity(t *testing.T) {
	t.Parallel()

	for _, s := range []string{
		"",
		"a",
		"{",
		"{{",
		"<p>plain</p>",
		"{{.X}}",
		strings.Repeat("body {{.X}} ", 64),
	} {
		want := form.HashSource(s)
		if got := draw.HashString(s); got != want {
			t.Fatalf("draw.HashString(%q) = %d, want %d", s, got, want)
		}
		if got := nav.Hash(s); got != want {
			t.Fatalf("nav.Hash(%q) = %d, want %d", s, got, want)
		}
	}
}

// TestFormSyncBodyMemoSkipsCleanHash proves the body memo: once a publish
// pairs the template body with its hash, a clean redraw that reuses the same
// body (the focused-field case, which cannot take the early unchanged skip)
// reuses that hash instead of folding the bytes again, and publishes the
// same source, body, and live values.
func TestFormSyncBodyMemoSkipsCleanHash(t *testing.T) {
	ctx := context.Background()
	const doc = `<!doctype html><html><head><title>memo</title></head><body>` +
		`<input id="field" type="text" value="abc"></body></html>`

	p, err := New(Config{HTML: doc, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "field"); err != nil {
		t.Fatal(err)
	}

	beforeSource := p.source
	beforeBody := p.lastSyncBody
	beforeValue := p.FormValue("field")
	beforeLayouts := p.Stats().Layouts
	memo := p.lastBody
	if memo.source != beforeBody || memo.hash != form.HashSource(beforeBody) {
		t.Fatalf("body memo = (%q, %d), want (%q, %d)",
			memo.source, memo.hash, beforeBody, form.HashSource(beforeBody))
	}

	calls := syncHashCalls.Load()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// The body hash is reused from the memo; the one remaining call hashes
	// the freshly built control rewrite output, a different string on every
	// redraw that the body memo cannot cover.
	if d := syncHashCalls.Load() - calls; d != 1 {
		t.Fatalf("clean focused redraw hashed %d times, want 1 (the rewrite output)", d)
	}
	if p.source != beforeSource || p.lastSyncBody != beforeBody {
		t.Fatal("clean focused redraw changed the synced source")
	}
	if p.FormValue("field") != beforeValue {
		t.Fatalf("field value = %q, want %q", p.FormValue("field"), beforeValue)
	}
	if p.Stats().Layouts != beforeLayouts {
		t.Fatalf("clean focused redraw relaid out: %d -> %d", beforeLayouts, p.Stats().Layouts)
	}
	if p.lastBody != memo {
		t.Fatal("clean focused redraw replaced the body memo")
	}
}

// TestSameSourceBytes proves the pointer fast path only accepts equal
// lengths over one backing array, so equal content in a fresh allocation
// still takes the hashing path.
func TestSameSourceBytes(t *testing.T) {
	t.Parallel()

	const text = "same bytes"
	fresh := text[:5] + text[5:]
	long := "X" + text
	offset := long[1:]

	if !sameSourceBytes(text, text[:]) {
		t.Fatal("same backing array and length not detected")
	}
	if sameSourceBytes(text, fresh) {
		t.Fatal("a fresh allocation was reported identical")
	}
	if sameSourceBytes(text, offset) {
		t.Fatal("an equal-length slice at another offset was reported identical")
	}
	if sameSourceBytes(text, text[:len(text)-1]) {
		t.Fatal("a shorter string was reported identical")
	}
	if sameSourceBytes("", "") {
		t.Fatal("empty strings must take the hashing path")
	}
}

// TestFormHashOncePerChangedFrame proves a body is folded once the first
// time it is seen and never again while the memo still holds it, using
// counters not timing. A simple input-free, emoji-free doc keeps every
// stage on the same string, so the old path hashed it three times (sync,
// emoji, noteSource) and the body memo now hashes it once on the first
// sync and reuses that value on every later frame, dirty or clean.
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
	// A dirty frame on a source with no template actions reuses
	// nav.Current as the body, so the memoized hash still applies and no
	// fold runs. The memo is keyed by body bytes, not by the dirty flag.
	if d := syncHashCalls.Load() - h1; d != 0 {
		t.Fatalf("dirty frame hashed an unchanged body %d times, want 0", d)
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
