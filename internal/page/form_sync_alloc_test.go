package page

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

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
	p.markFormDirty()
	_ = p.syncForm(syncBody) // warm the scan entry and the emoji gate
	if a := testing.AllocsPerRun(100, func() {
		p.markFormDirty()
		_ = p.syncForm(syncBody)
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
	q.markFormDirty()
	_ = q.syncForm(qbody)
	// Budget: merge 4 + bind snapshot 1 + rewrite ~13, all else cached.
	// A re-added full source pass would cost a builder plus a copy
	// on top, so the bound catches pass regressions, not alloc noise.
	if a := testing.AllocsPerRun(100, func() {
		q.markFormDirty()
		_ = q.syncForm(qbody)
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
