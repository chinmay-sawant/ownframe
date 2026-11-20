package page

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestButtonsFocusAndDisable(t *testing.T) {
	p := bftPage(t, `<button id="a">A</button><button id="b" disabled>B</button><button id="c"><span>C</span></button>`)
	bftDraw(t, p, nil)
	ctx := context.Background()
	for _, want := range []string{"a", "c", "a"} {
		if err := p.FocusNext(ctx); err != nil {
			t.Fatal(err)
		}
		if p.FocusID() != want {
			t.Fatalf("focus %q, want %q", p.FocusID(), want)
		}
		if p.FocusedField() != "" {
			t.Fatal("button focus reported as an editable field")
		}
	}
	calls := 0
	p.Handle(Handlers{Click: func(context.Context, Box) error { calls++; return nil }})
	box := boxByID(p, "b")
	if err := p.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("disabled button invoked handler")
	}
}

func TestButtonRouteAndDisabledRoute(t *testing.T) {
	p := bftPage(t, `<button id="off" disabled data-action="next">Off</button><button id="next" data-action="next"><span>Next</span></button>`)
	bftDraw(t, p, nil)
	if err := p.Route("next", `<p id="destination">Destination</p>`); err != nil {
		t.Fatal(err)
	}
	calls := 0
	p.Handle(Handlers{Click: func(context.Context, Box) error { calls++; return nil }})
	bftClick(t, p, "off")
	if boxByID(p, "next").ID == "" || calls != 0 {
		t.Fatal("disabled button followed route or called handler")
	}
	bftClick(t, p, "next")
	if boxByID(p, "destination").ID == "" || calls != 0 {
		t.Fatal("enabled button failed to route")
	}
}

func TestFocusSetsHostState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := focusFields(t)

	if err := p.FocusNext(ctx); err != nil {
		t.Fatal(err)
	}

	if got := renderState(p).Focus; got != "a" {
		t.Fatalf("render focus %q", got)
	}
}

func TestEscapeClearsFocus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := focusFields(t)

	if err := p.Focus(ctx, "a"); err != nil {
		t.Fatal(err)
	}

	gen := p.Generation()
	if err := p.KeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}

	if p.FocusID() != "" || p.Generation() <= gen {
		t.Fatalf("focus %q generation %d", p.FocusID(), p.Generation())
	}
}

func TestCaretKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd ef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	setCaret(p, 2, false)

	var seen []string
	p.Handle(Handlers{KeyDown: func(_ context.Context, key string) error {
		seen = append(seen, key)
		return nil
	}})

	steps := []struct {
		key  string
		want int
	}{
		{"arrowleft", 1},
		{"arrowright", 2},
		{"end", 7},
		{"home", 0},
		{"ctrl+arrowright", 5},
		{"shift+arrowright", 6},
	}

	for _, step := range steps {
		if err := p.KeyDown(ctx, step.key); err != nil {
			t.Fatal(err)
		}

		if formState(p).Caret != step.want {
			t.Fatalf("%s caret %d want %d", step.key, formState(p).Caret, step.want)
		}
	}

	if formState(p).Anchor != 5 {
		t.Fatalf("shift anchor %d", formState(p).Anchor)
	}

	if len(seen) != len(steps) || seen[0] != "arrowleft" {
		t.Fatalf("handler saw %q", seen)
	}
}

func focusFields(t *testing.T) *Page {
	t.Helper()

	p := bftPage(t, `<input id="a" type="text"`+bedWide+`>`+
		`<input id="b" type="text" disabled`+bedWide+`>`+
		`<input id="c" type="text" tabindex="-1"`+bedWide+`>`+
		`<input id="d" type="text"`+bedWide+`>`)
	bftDraw(t, p, nil)

	return p
}

func TestFocusTraversal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := focusFields(t)

	steps := []struct {
		next func(context.Context) error
		want string
	}{
		{p.FocusNext, "a"},
		{p.FocusNext, "d"},
		{p.FocusNext, "a"},
		{p.FocusPrev, "d"},
		{p.FocusPrev, "a"},
	}

	for i, step := range steps {
		if err := step.next(ctx); err != nil {
			t.Fatal(err)
		}

		if got := p.FocusID(); got != step.want {
			t.Fatalf("step %d focus %q want %q", i, got, step.want)
		}
	}
}

func TestFocusSkipsDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := focusFields(t)

	if err := p.Focus(ctx, "b"); err != nil {
		t.Fatal(err)
	}

	if p.FocusID() != "" {
		t.Fatalf("disabled focus %q", p.FocusID())
	}

	if err := p.Focus(ctx, "c"); err != nil {
		t.Fatal(err)
	}

	if p.FocusID() != "c" {
		t.Fatalf("programmatic focus %q", p.FocusID())
	}

	if err := p.FocusNext(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FocusID() != "a" {
		t.Fatalf("next from tabindex=-1 is %q", p.FocusID())
	}
}

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

	markFormDirty(p)
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

func TestKeyboardButtonActivationAndFormSubmit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="action" type="button" data-action="run">Run</button><input id="field">`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	clicks, submits := 0, 0
	p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
		if box.Action == "run" {
			clicks++
		}
		return nil
	}, Submit: func(context.Context) error { submits++; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "action"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 || submits != 0 {
		t.Fatalf("Space: clicks=%d submits=%d", clicks, submits)
	}
	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if clicks != 2 || submits != 0 {
		t.Fatalf("Enter: clicks=%d submits=%d", clicks, submits)
	}
	if err := p.Focus(ctx, "field"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if clicks != 2 || submits != 1 {
		t.Fatalf("field submit: clicks=%d submits=%d", clicks, submits)
	}
}

func TestKeyboardButtonFollowsRoute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="next" data-action="next">Next</button>`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Route("next", `<p>Next page</p>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "next"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != `<p>Next page</p>` {
		t.Fatal("keyboard activation did not navigate")
	}
}

func TestDisabledFocusedButtonDoesNotSubmit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="action" type="button">Action</button>`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	submits := 0
	p.Handle(Handlers{Submit: func(context.Context) error { submits++; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "action"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceHTML(ctx, p, `<button id="action" type="button" disabled>Action</button>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if submits != 0 {
		t.Fatal("disabled button submitted")
	}
}

func TestExplicitSubmitButtonUsesSubmitHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := New(Config{HTML: `<button id="submit" type="submit">Submit</button>`, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	submits := 0
	p.Handle(Handlers{Submit: func(context.Context) error { submits++; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "submit"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if submits != 1 {
		t.Fatalf("submits=%d", submits)
	}
}

func TestRewriteButtonsEscapedQuotesRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="submit" id="go" value="a&quot;b" data-label="<b>not a tag</b>">`
	got := rewriteButtons(src)

	if !strings.Contains(got, "<button") {
		t.Fatalf("rewriteButtons dropped button: %q", got)
	}

	// Attribute text is part of the serialized attribute value, not executable
	// markup in the rewritten output.
	if got == src {
		t.Fatalf("button was not rewritten: %q", got)
	}
}

func TestRewriteButtonsNestedInputAttrsRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="submit" id="go" value="run" data-x="<input type='text' value='nested'>">`
	got := rewriteButtons(src)

	if strings.Count(got, "<button") != 1 {
		t.Fatalf("expected one button, got %q", got)
	}

	if strings.Count(got, "<button") != 1 {
		t.Fatalf("expected one button, got %q", got)
	}
}

func TestScanControlsNestedInputAttrsRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="text" id="name" value="one" data-x="<input type='text' id='evil'>">`
	spans := form.Scan(src)

	if len(spans) != 1 || spans[0].Control.ID != "name" {
		t.Fatalf("spans = %#v", spans)
	}
}

func TestScanControlsDoubleEscapedQuotesRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="text" id="q" value="&amp;quot;quoted&amp;quot;">`
	spans := form.Scan(src)

	if len(spans) != 1 {
		t.Fatalf("spans = %#v", spans)
	}

	// The value is unescaped exactly once by the scanner, so a double-
	// escaped quote stays escaped rather than becoming a live quote.
	if strings.Contains(spans[0].Control.Value, `"`) {
		t.Fatalf("quote unexpectedly unescaped twice: %q", spans[0].Control.Value)
	}
}

func TestAttrValueDoubleEscapedQuotesRegression(t *testing.T) {
	t.Parallel()

	raw := `<input value="&amp;quot;x&amp;quot;">`
	got := form.AttrValue(raw, "value")

	if got == `"x"` {
		t.Fatalf("attrValue collapsed double escaping: %q", got)
	}
}

func TestSelectAtOutsideBlurs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="ab"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "ab")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectAt(ctx, left+width/2, base-4); err != nil {
		t.Fatal(err)
	}

	if err := p.SelectAt(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}

	if p.FocusID() != "" {
		t.Fatalf("focus %q after outside click", p.FocusID())
	}
}

func TestSelectAtLandsOnGlyph(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="1234567890"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "1234567890")
	left, width := runSpan(p, op)
	base := runBase(p, op)
	want := []int{2, 5, 8}

	for i, frac := range []float64{0.2, 0.5, 0.8} {
		if err := p.SelectAt(ctx, left+width*frac, base-4); err != nil {
			t.Fatal(err)
		}

		st := formState(p)
		if p.FocusID() != "t" || st.Caret != want[i] || st.Anchor != want[i] {
			t.Fatalf("frac %.1f focus %q caret %d anchor %d", frac, p.FocusID(), st.Caret, st.Anchor)
		}
	}
}

func TestDragExtendsRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="1234567890"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "1234567890")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectAt(ctx, left+width*0.2, base-4); err != nil {
		t.Fatal(err)
	}

	if err := p.Drag(ctx, left+width*0.8, base-4); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	start, end := st.Bounds(10)
	if start != 2 || end != 8 || st.Caret != 8 || st.Anchor != 2 {
		t.Fatalf("range [%d,%d) caret %d anchor %d", start, end, st.Caret, st.Anchor)
	}

	if p.FormSelected("t") {
		t.Fatal("partial range reports selected")
	}

	if err := p.Drag(ctx, left+width*0.4, base-4); err != nil {
		t.Fatal(err)
	}

	st = formState(p)
	if start, end := st.Bounds(10); start != 2 || end != 4 {
		t.Fatalf("second drag [%d,%d)", start, end)
	}
}

func TestDragAcrossWrappedLines(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	value := "aaaa bbbb cccc dddd eeee ffff gggg hhhh iiii jjjj"
	p := bftPage(t, `<input id="t" type="text" value="`+value+`" style="display:block;width:120px;height:80px">`)
	bftDraw(t, p, nil)

	box := boxByID(p, "t")
	pt := p.draw.Display.PointsPerPixel

	var first, last *DisplayOp
	for i := range p.draw.Display.Ops {
		op := &p.draw.Display.Ops[i]
		if op.Kind != DisplayOpText || op.Text == "" || op.Font == nil {
			continue
		}

		if op.Y < box.Y*pt || op.Y > (box.Y+box.H)*pt {
			continue
		}

		if first == nil || op.Y < first.Y {
			first = op
		}

		if last == nil || op.Y > last.Y {
			last = op
		}
	}

	if first == nil || last == nil || first == last {
		t.Fatal("value did not wrap")
	}

	if err := p.SelectAt(ctx, first.X/pt, first.Y/pt-4); err != nil {
		t.Fatal(err)
	}

	if err := p.Drag(ctx, (last.X+last.W)/pt+4, last.Y/pt-4); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	start, end := st.Bounds(runeLen(value))
	if start != 0 || end != runeLen(value) {
		t.Fatalf("wrapped range [%d,%d) of %d", start, end, runeLen(value))
	}

	if !p.FormSelected("t") {
		t.Fatal("whole wrapped value not selected")
	}
}

func TestFormSelectedTracksSelectAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	body := `<input id="a" type="text"` + bedWide + `><input id="b" type="text"` + bedWide + `>`
	p := bftPage(t, body)
	bftDraw(t, p, nil)
	bftClick(t, p, "a")
	if err := p.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}
	if p.FormSelected("a") {
		t.Fatal("selected before SelectAll")
	}
	if err := p.SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	if !p.FormSelected("a") || p.FormSelected("b") {
		t.Fatalf("after SelectAll a=%v b=%v", p.FormSelected("a"), p.FormSelected("b"))
	}

	if err := p.Type(ctx, "c"); err != nil {
		t.Fatal(err)
	}

	if p.FormSelected("a") {
		t.Fatal("selected after Type")
	}
}

func TestFormSelectedFalseAfterBlur(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="a" type="text"`+bedWide+`><div id="plain">x</div>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "a")
	if err := p.SelectAll(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.FormSelected("a") {
		t.Fatal("not selected before blur")
	}

	bftClick(t, p, "plain")

	if p.FormSelected("a") {
		t.Fatal("selected after blur")
	}
}

func TestFormSelectedNilForm(t *testing.T) {
	t.Parallel()

	p := &Page{}
	if p.FormSelected("a") {
		t.Fatal("nil form reports selected")
	}
}

// formSyncGoldenDoc is the deterministic fixture for sync parity: controls
// of every kind, a data-bound field, a submit input the button rewrite
// replaces, a plain button, a head block, and one emoji. The golden strings
// below are the exact pre-edit outputs of the sync pipeline for it.
const formSyncGoldenDoc = `<!doctype html><html><head><title>form golden</title></head><body>` +
	`<input id="name" type="text" value="Ada">` +
	`<input id="pw" type="password" value="secret">` +
	`<input id="mail" type="email" value="a@b.c" placeholder="you">` +
	`<input id="bound" type="text" data-bind="Name" value="overridden">` +
	`<input id="ok" type="checkbox" checked>` +
	`<input id="r1" type="radio" name="g" value="1">` +
	`<input id="r2" type="radio" name="g" value="2" checked>` +
	`<textarea id="bio">hi &amp; bye</textarea>` +
	`<select id="city" name="city"><option value="a">A</option><option value="b" selected>B</option></select>` +
	`<input id="go" type="submit" value="Send">` +
	`<button id="plain">Plain</button>` +
	`<p>emoji 😂 here</p>` +
	`</body></html>`

type formSyncGoldenData struct {
	Name string
}

const formSyncGoldenRewritten = "<!doctype html><html><head><title>form golden</title></head><body><input id=\"name\" type=\"text\" value=\"Ada\"><input id=\"pw\" type=\"password\" value=\"secret\"><input id=\"mail\" type=\"email\" value=\"a@b.c\" placeholder=\"you\"><input id=\"bound\" type=\"text\" data-bind=\"Name\" value=\"overridden\"><input id=\"ok\" type=\"checkbox\" checked><input id=\"r1\" type=\"radio\" name=\"g\" value=\"1\"><input id=\"r2\" type=\"radio\" name=\"g\" value=\"2\" checked><textarea id=\"bio\">hi &amp; bye</textarea><select id=\"city\" name=\"city\"><option value=\"a\">A</option><option value=\"b\" selected>B</option></select><button id=\"go\">Send</button><button id=\"plain\">Plain</button><p>emoji 😂 here</p></body></html>"

const formSyncGoldenState = "<!doctype html><html><head><style>[data-ownframe-field]{display:inline-block;border:1px solid #c8c2b4;padding:4px 6px;min-width:10em;min-height:1.6em;white-space:pre-wrap}[data-ownframe-field][data-ownframe-focus=\"1\"]{border:1px solid #1a56db;outline:1px solid #1a56db}[data-ownframe-selection]{background:#d6e2ff}[data-ownframe-composition]{text-decoration:underline}[data-ownframe-field][data-ownframe-placeholder=\"1\"]{color:#6b7280}</style><title>form golden</title></head><body><span id=\"name\" data-ownframe-field=\"input\">Ada</span><span id=\"pw\" data-ownframe-field=\"input\">••••••</span><span id=\"mail\" placeholder=\"you\" data-ownframe-field=\"input\">a@b.c</span><span id=\"bound\" data-bind=\"Name\" data-ownframe-field=\"input\">Bob</span><input id=\"ok\" type=\"checkbox\" checked><input id=\"r1\" name=\"g\" value=\"1\" type=\"radio\"><input id=\"r2\" name=\"g\" value=\"2\" type=\"radio\" checked><span id=\"bio\" data-ownframe-field=\"textarea\">hi &amp; bye</span><select id=\"city\" name=\"city\" data-ownframe-field=\"select\">B</select><button id=\"go\">Send</button><button id=\"plain\">Plain</button><p>emoji 😂 here</p></body></html>"

const formSyncGoldenFinal = "<!doctype html><html><head><style>img[data-ownframe-emoji]{width:1em;height:1em;vertical-align:-0.125em}</style><style>[data-ownframe-field]{display:inline-block;border:1px solid #c8c2b4;padding:4px 6px;min-width:10em;min-height:1.6em;white-space:pre-wrap}[data-ownframe-field][data-ownframe-focus=\"1\"]{border:1px solid #1a56db;outline:1px solid #1a56db}[data-ownframe-selection]{background:#d6e2ff}[data-ownframe-composition]{text-decoration:underline}[data-ownframe-field][data-ownframe-placeholder=\"1\"]{color:#6b7280}</style><title>form golden</title></head><body><span id=\"name\" data-ownframe-field=\"input\">Ada</span><span id=\"pw\" data-ownframe-field=\"input\">••••••</span><span id=\"mail\" placeholder=\"you\" data-ownframe-field=\"input\">a@b.c</span><span id=\"bound\" data-bind=\"Name\" data-ownframe-field=\"input\">Bob</span><input id=\"ok\" type=\"checkbox\" checked><input id=\"r1\" name=\"g\" value=\"1\" type=\"radio\"><input id=\"r2\" name=\"g\" value=\"2\" type=\"radio\" checked><span id=\"bio\" data-ownframe-field=\"textarea\">hi &amp; bye</span><select id=\"city\" name=\"city\" data-ownframe-field=\"select\">B</select><button id=\"go\">Send</button><button id=\"plain\">Plain</button><p>emoji <img data-ownframe-emoji=\"1\" src=\"emoji/1f602\" alt=\"😂\"> here</p></body></html>"

const formSyncGoldenLastSync = "<!doctype html><html><head><title>form golden</title></head><body><input id=\"name\" type=\"text\" value=\"Ada\"><input id=\"pw\" type=\"password\" value=\"secret\"><input id=\"mail\" type=\"email\" value=\"a@b.c\" placeholder=\"you\"><input id=\"bound\" type=\"text\" data-bind=\"Name\" value=\"overridden\"><input id=\"ok\" type=\"checkbox\" checked><input id=\"r1\" type=\"radio\" name=\"g\" value=\"1\"><input id=\"r2\" type=\"radio\" name=\"g\" value=\"2\" checked><textarea id=\"bio\">hi &amp; bye</textarea><select id=\"city\" name=\"city\"><option value=\"a\">A</option><option value=\"b\" selected>B</option></select><button id=\"go\">Send</button><button id=\"plain\">Plain</button><p>emoji 😂 here</p></body></html>"

const formSyncGoldenSkip = "<style>img[data-ownframe-emoji]{width:1em;height:1em;vertical-align:-0.125em}</style><p id=\"m\">Good <img data-ownframe-emoji=\"1\" src=\"emoji/1f602\" alt=\"😂\"> and plain</p>"

// TestFormSyncGoldenParity replays the sync pipeline on the golden fixture
// and compares every stage byte-for-byte with the pre-edit run: the button
// rewrite, the fused scan, the merge, the cached rewrite, the emoji pass,
// and the live redraw (published source plus sync bookkeeping).
func TestFormSyncGoldenParity(t *testing.T) {
	rewritten := rewriteButtons(formSyncGoldenDoc)
	if rewritten != formSyncGoldenRewritten {
		t.Fatalf("rewrite mismatch:\n%s", rewritten)
	}

	hash := form.HashSource(rewritten)
	has, spans, buttons := form.ScanCached(rewritten, hash)
	if !has {
		t.Fatal("probe missed the golden controls")
	}
	if buttons {
		t.Fatal("rewritten source still reports button inputs")
	}

	want := []struct {
		start, end      int
		id, tag, kind   string
		name, bind, val string
		checked         bool
		options         []form.Option
	}{
		{66, 107, "name", "input", "text", "", "", "Ada", false, nil},
		{107, 153, "pw", "input", "password", "", "", "secret", false, nil},
		{153, 215, "mail", "input", "email", "", "", "a@b.c", false, nil},
		{215, 281, "bound", "input", "text", "", "Name", "overridden", false, nil},
		{281, 320, "ok", "input", "checkbox", "", "", "", true, nil},
		{320, 367, "r1", "input", "radio", "g", "", "1", false, nil},
		{367, 422, "r2", "input", "radio", "g", "", "2", true, nil},
		{422, 464, "bio", "textarea", "", "", "", "hi & bye", false, nil},
		{464, 568, "city", "select", "", "city", "", "b", false, []form.Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B", Selected: true},
		}},
		{568, 597, "go", "button", "button", "", "", "", false, nil},
		{597, 630, "plain", "button", "button", "", "", "", false, nil},
	}
	if len(spans) != len(want) {
		t.Fatalf("spans = %d, want %d", len(spans), len(want))
	}
	for i, w := range want {
		c := spans[i].Control
		if spans[i].Start != w.start || spans[i].End != w.end ||
			c.ID != w.id || c.Tag != w.tag || c.Type != w.kind ||
			c.Name != w.name || c.Bind != w.bind || c.Value != w.val ||
			c.Checked != w.checked || !reflect.DeepEqual(c.Options, w.options) {
			t.Fatalf("span %d = %+v [%d:%d], want %+v", i, c, spans[i].Start, spans[i].End, w)
		}
	}

	byID, order := form.Merge(map[string]Control{}, spans)
	form.ApplyBinds(byID, func(c Control) (Control, bool) {
		if c.Bind == "Name" {
			c.Value = "Bob"

			return c, true
		}

		return c, false
	})
	if len(order) != len(want) {
		t.Fatalf("order = %d, want %d", len(order), len(want))
	}

	if got := form.RewriteCached(rewritten, hash, spans, byID, form.Caret{Caret: -1}); got != formSyncGoldenState {
		t.Fatalf("rewrite mismatch:\n%s", got)
	}

	final := draw.EmojiPass(formSyncGoldenState)
	if final != formSyncGoldenFinal {
		t.Fatalf("emoji mismatch:\n%s", final)
	}

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
}

// TestFormScanSingleMatchesScan proves the fused single-pass scan finds exactly
// the spans Scan finds, on control, comment, script, style, and malformed
// edge cases, and that its button flag is exactly the button rewrite's
// identity condition: no flag means the rewrite returns the source unchanged.
func TestFormScanSingleMatchesScan(t *testing.T) {
	docs := []string{
		formSyncGoldenDoc,
		`<!-- <input id="z"> --><script>var s="<input id=no>";</script><style>.a{content:"<input>"}</style>`,
		`<input type=submit value=Go><input id="t" TYPE="TEXT" value="a"><INPUT ID="U" TYPE="CHECKBOX" CHECKED>`,
		`<textarea id="m">x</textarea><select id="s"><option value="1">One</option></select>`,
		`<button id="b">Go</button><input id="x" type="button"><input id="oops"`,
		`<div title="<input">text</div><input id="q" type="reset"><input id="w" type="text" value="v">`,
	}
	for _, doc := range docs {
		spans, buttons := form.ScanSingle(doc)
		if got := form.Scan(doc); !reflect.DeepEqual(spans, got) {
			t.Fatalf("scan mismatch for %q:\nsingle=%+v\nscan  =%+v", doc, spans, got)
		}
		if changed := rewriteButtons(doc) != doc; changed != buttons {
			t.Fatalf("button flag = %v but rewrite changed = %v for %q", buttons, changed, doc)
		}
	}
}

// TestFormSyncSkipsUnchanged proves the repeat-redraw fast paths with
// counters, not timing: a dirty redraw of the same body must not rescan,
// re-probe, re-run the head scan, rebuild, re-run the button rewrite, or
// re-run the emoji probe, and the output must be identical.
func TestFormSyncSkipsUnchanged(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<p id="m">Good 😂 and plain</p>`, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.source

	RequestFocus(p, "ghost")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source != before {
		t.Fatalf("repeat redraw changed output:\n%s", p.source)
	}

	f0, e0, r0 := form.Telemetry(), draw.EmojiStats(), rewriteButtonCalls.Load()
	RequestFocus(p, "ghost")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source != before {
		t.Fatalf("repeat redraw changed output:\n%s", p.source)
	}
	if f1 := form.Telemetry(); f1 != f0 {
		t.Fatalf("repeat redraw rescanned: %+v -> %+v", f0, f1)
	}
	if e1 := draw.EmojiStats(); e1.Probes != e0.Probes || e1.Hits != e0.Hits+1 {
		t.Fatalf("emoji probe not gated: %+v -> %+v", e0, e1)
	}
	if r1 := rewriteButtonCalls.Load(); r1 != r0 {
		t.Fatalf("button rewrite ran on span-free body: %d -> %d", r0, r1)
	}
}

// TestFormSyncSkipsRewriteWithoutSpans proves the span-count gate at the
// rewrite level: a body with no spans, and a body with only plain buttons,
// return the source unchanged without moving the rewrite counter, while a
// body with a text field rebuilds exactly once.
func TestFormSyncSkipsRewriteWithoutSpans(t *testing.T) {
	f0 := form.Telemetry()

	plain := `<html><head><title>t</title></head><body><p>hi</p></body></html>`
	if got := form.RewriteCached(plain, form.HashSource(plain), nil, nil, form.Caret{Caret: -1}); got != plain {
		t.Fatalf("span-free rewrite changed output:\n%s", got)
	}

	buttons := `<button id="a">A</button><button id="b">B</button>`
	spans, _ := form.ScanSingle(buttons)
	if got := form.RewriteCached(buttons, form.HashSource(buttons), spans, nil, form.Caret{Caret: -1}); got != buttons {
		t.Fatalf("button-only rewrite changed output:\n%s", got)
	}

	field := `<input id="t" type="text" value="v">`
	fhash := form.HashSource(field)
	_, fspans, _ := form.ScanCached(field, fhash) // stores the head-cache entry, like the page sync
	first := form.RewriteCached(field, fhash, fspans, nil, form.Caret{Caret: -1})
	if first == field {
		t.Fatal("field rewrite left the source unchanged")
	}
	second := form.RewriteCached(field, fhash, fspans, nil, form.Caret{Caret: -1})
	if second != first {
		t.Fatalf("rewrite not stable:\n%s\n%s", first, second)
	}

	// The two span-free rewrites move nothing; the field body pays one
	// probe, one scan, one head evaluation (reused by the second build),
	// and one build per call.
	if f1 := form.Telemetry(); f1.Probes != f0.Probes+1 || f1.Scans != f0.Scans+1 ||
		f1.Heads != f0.Heads+1 || f1.Rewrites != f0.Rewrites+2 {
		t.Fatalf("rewrite gate miscounted: %+v -> %+v", f0, f1)
	}
}

// TestFormSyncAllocs bounds the allocations of the cached paths: hashing,
// a cached scan, a gated emoji pass, and the no-controls sync must not
// allocate, and a cached full merge stays in single digits.
func TestFormSyncAllocs(t *testing.T) {
	if a := testing.AllocsPerRun(100, func() {
		_ = form.HashSource(formSyncGoldenDoc)
	}); a != 0 {
		t.Fatalf("HashSource allocs = %v, want 0", a)
	}

	body := formSyncGoldenRewritten
	hash := form.HashSource(body)
	_, _, _ = form.ScanCached(body, hash) // warm the entry; the loop below must hit it
	if a := testing.AllocsPerRun(100, func() {
		_, _, _ = form.ScanCached(body, hash)
	}); a != 0 {
		t.Fatalf("cached scan allocs = %v, want 0", a)
	}

	out := draw.EmojiPass(formSyncGoldenFinal)
	if a := testing.AllocsPerRun(100, func() {
		_ = draw.EmojiPass(formSyncGoldenFinal)
	}); a != 0 {
		t.Fatalf("gated emoji allocs = %v, want 0", a)
	}
	if out != formSyncGoldenFinal {
		t.Fatal("emoji gate changed the output")
	}

	ctx := context.Background()
	p, err := New(Config{HTML: `<p id="m">Good 😂 and plain</p>`, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	syncBody := p.lastSyncBody
	markFormDirty(p)
	_ = syncForm(p, syncBody) // warm the scan entry and the emoji gate
	if a := testing.AllocsPerRun(100, func() {
		markFormDirty(p)
		_ = syncForm(p, syncBody)
	}); a != 0 {
		t.Fatalf("skip-path sync allocs = %v, want 0", a)
	}

	q, err := New(Config{HTML: `<input id="t" type="text" value="v">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	qbody := q.lastSyncBody
	markFormDirty(q)
	_ = syncForm(q, qbody)
	// Budget: merge 4 + bind snapshot 1 + rewrite ~13, all else cached.
	// A re-added full source pass would cost a builder plus a copy
	// on top, so the bound catches pass regressions, not alloc noise.
	if a := testing.AllocsPerRun(100, func() {
		markFormDirty(q)
		_ = syncForm(q, qbody)
	}); a > 24 {
		t.Fatalf("cached full sync allocs = %v, want <= 24", a)
	}
}

type formSyncBindData struct {
	Name string
}

// TestFormSyncConcurrentSetData hammers SetData from worker goroutines while
// the page loop redraws, proving the narrowed sync locks stay race-clean.
// No output assertion: a concurrent publish legitimately wins either way.
func TestFormSyncConcurrentSetData(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="t" type="text" data-bind="Name" value="x">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(&formSyncBindData{Name: "zero"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				p.SetData(&formSyncBindData{Name: fmt.Sprintf("g%d-%d", g, i)})
			}
		}(g)
	}
	for i := 0; i < 50; i++ {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()

	p.SetData(&formSyncBindData{Name: "settled"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("t"); got != "settled" {
		t.Fatalf("value = %q, want %q", got, "settled")
	}
}

// TestSyncFormSkipsWithoutControls redraws an input-free page twice: the
// second sync must take the no-controls fast path, which keeps the same
// bookkeeping as the full merge (body recorded, flag cleared, no live
// controls) and still runs the emoji pass.
func TestSyncFormSkipsWithoutControls(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<p id="m">Good 😂 and plain</p>`, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	fs := formState(p)
	if fs == nil {
		t.Fatal("no form state after redraw")
	}

	if len(fs.ByID) != 0 || len(fs.Order) != 0 {
		t.Fatalf("controls = %d/%d, want 0/0", len(fs.ByID), len(fs.Order))
	}

	if p.formDirty {
		t.Fatal("formDirty still set after sync")
	}

	if p.lastSyncBody == "" {
		t.Fatal("lastSyncBody not recorded")
	}

	if strings.Contains(p.source, "data-ownframe-field") {
		t.Fatal("input-free source carries field markup")
	}

	// The emoji pass still runs on the skip path.
	if !strings.Contains(p.source, `src="emoji/1f602"`) {
		t.Fatalf("emoji lost on input-free page: %s", p.source)
	}

	// A focus request for an absent control is consumed, exactly as the
	// full merge does, so it never sticks around for a later template.
	RequestFocus(p, "ghost")

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if p.ff.hasFocusRequest {
		t.Fatal("focus request left pending on input-free page")
	}

	if got := formState(p); len(got.ByID) != 0 || len(got.Order) != 0 {
		t.Fatalf("controls = %d/%d after second sync, want 0/0", len(got.ByID), len(got.Order))
	}
}

// TestSyncFormKeepsControlsWithInput proves pages with inputs still run
// the full merge: the live set is populated and a user value survives a
// redraw that re-executes the template.
func TestSyncFormKeepsControlsWithInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="t" type="text" value="hi">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	fs := formState(p)
	if fs == nil || len(fs.ByID) != 1 || len(fs.Order) != 1 {
		t.Fatalf("live controls not merged: %+v", fs)
	}

	p.SetFormValue("t", "typed")

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("t"); got != "typed" {
		t.Fatalf("value = %q, want %q", got, "typed")
	}

	if !strings.Contains(p.source, "data-ownframe-field") {
		t.Fatal("input page lost its field markup")
	}
}

func TestSelectWordAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="hello world"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "hello world")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectWordAt(ctx, left+width*0.1, base-4); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	start, end := st.Bounds(11)
	if start != 0 || end != 5 {
		t.Fatalf("first word [%d,%d)", start, end)
	}

	if err := p.SelectWordAt(ctx, left+width*0.6, base-4); err != nil {
		t.Fatal(err)
	}

	st = formState(p)
	start, end = st.Bounds(11)
	if start != 6 || end != 11 {
		t.Fatalf("second word [%d,%d)", start, end)
	}
}

func TestSelectLineAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<textarea id="t" style="display:block;width:220px;height:60px">one
two</textarea>`)
	bftDraw(t, p, nil)

	op := runOp(p, "two")
	if op == nil {
		t.Fatal("no second line run")
	}

	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectLineAt(ctx, left+width/2, base-4); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	if start, end := st.Bounds(7); start != 4 || end != 7 {
		t.Fatalf("line [%d,%d)", start, end)
	}
}

func TestRangeClampsAfterRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	st := formState(p)
	st.Anchor, st.Caret = 2, 5
	c := st.ByID["t"]
	c.Value = "ab"
	st.ByID["t"] = c

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	st = formState(p)
	if st.Caret != 2 || st.Anchor != 2 {
		t.Fatalf("caret %d anchor %d after clamp", st.Caret, st.Anchor)
	}

	if start, end := st.Bounds(2); start != 2 || end != 2 {
		t.Fatalf("range [%d,%d) after clamp", start, end)
	}
}

// runOp returns the text operation that paints want, or nil.
func runOp(p *Page, want string) *DisplayOp {
	for i := range p.draw.Display.Ops {
		op := &p.draw.Display.Ops[i]
		if op.Kind == DisplayOpText && op.Text == want {
			return op
		}
	}

	return nil
}

// runSpan returns the run's left edge and width in CSS pixels.
func runSpan(p *Page, op *DisplayOp) (float64, float64) {
	pp := p.draw.Display.PointsPerPixel

	return op.X / pp, op.W / pp
}

// runBase returns the run's baseline in CSS pixels.
func runBase(p *Page, op *DisplayOp) float64 {
	return op.Y / p.draw.Display.PointsPerPixel
}

func TestRewriteButtonsSubmit(t *testing.T) {
	t.Parallel()

	src := `<input id="sub" class="go" type="submit" value="Send" data-action="send">`
	got := rewriteButtons(src)
	if !strings.Contains(got, `<button id="sub" class="go" data-action="send">Send</button>`) {
		t.Fatalf("%s", got)
	}
	if strings.Contains(got, "type=") || strings.Contains(got, "value=") {
		t.Fatalf("%s", got)
	}
}

func TestRewriteButtonsDefaults(t *testing.T) {
	t.Parallel()

	if got := rewriteButtons(`<input type="reset">`); !strings.Contains(got, ">Reset</button>") {
		t.Fatalf("%s", got)
	}
	if rewriteButtons("plain") != "plain" {
		t.Fatal("plain text changed")
	}
}

func TestButtonLeftToEngine(t *testing.T) {
	t.Parallel()

	src := `<button id="go">Sign in</button>`
	if got := rewriteControls(src, nil, nil, "", false); got != src {
		t.Fatalf("%s", got)
	}
}

func TestButtonGetsPaintBox(t *testing.T) {
	t.Parallel()

	src := rewriteControls(`<button id="go" data-action="go">Sign in</button>`, nil, nil, "", false)
	_, boxes, err := render.Paint(context.Background(), src, 400, 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range boxes {
		if b.ID != "go" {
			continue
		}
		if b.W <= 0 || b.H <= 0 {
			t.Fatalf("button box %+v", b)
		}

		return
	}
	t.Fatal("no button box")
}

func TestUnusedStatePseudosSkipRedraw(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`#e:focus-visible{outline:2px solid red}` +
		`</style></head><body><div id="e"></div></body></html>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.statePseudos.Hover || p.statePseudos.Active {
		t.Fatalf("unused state reported as used: %+v", p.statePseudos)
	}

	if !p.statePseudos.Focus {
		t.Fatal("focus-visible was not reported")
	}

	before := p.Stats().Redraws
	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := p.Press(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := p.Release(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Redraws; got != before {
		t.Fatalf("redraws = %d after hover/press/release, want %d", got, before)
	}
	if err := p.Hover(ctx, 300, 150); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Redraws; got != before {
		t.Fatalf("redraws = %d after unhover, want %d", got, before)
	}
}

func TestHoverStateStillRedrawsWhenUsed(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`#e:hover{width:120px}` +
		`</style></head><body><div id="e"></div></body></html>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := p.Stats().Redraws
	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Redraws; got <= before {
		t.Fatalf("redraws = %d after used hover, want more than %d", got, before)
	}
}
