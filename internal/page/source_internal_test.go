package page

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSelectControls(t *testing.T) {
	t.Parallel()

	p, err := New(Config{
		HTML:   `<select id="c"><option value="a">A</option><option value="b" selected>B</option></select>`,
		Width:  120,
		Height: 80,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := p.SelectControls()
	opts, ok := got["c"]
	if !ok || len(opts) != 2 {
		t.Fatalf("controls = %v", got)
	}
	if opts[0].Value != "a" || opts[0].Label != "A" || opts[0].Selected {
		t.Fatalf("opts = %+v", opts)
	}
	if opts[1].Value != "b" || !opts[1].Selected {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestPlaceCaretInsertsAtOffset(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<input id="t" type="text">`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.SetFormValue("t", "ab")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "t"); err != nil {
		t.Fatal(err)
	}
	// A caret for another field is ignored.
	p.PlaceCaret("missing", 0)
	p.PlaceCaret("t", 0)
	if err := p.Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("t"); got != "Xab" {
		t.Fatalf("value = %q", got)
	}
	// Out-of-range offsets clamp to the end.
	p.PlaceCaret("t", 999)
	if err := p.Type(ctx, "Y"); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("t"); got != "XabY" {
		t.Fatalf("value = %q", got)
	}
}

func TestSelectControlsEmpty(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<p>hi</p>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.SelectControls(); len(got) != 0 {
		t.Fatalf("controls = %v", got)
	}
	if p.ExecutedSource() == "" || p.ThemeSource() != "" {
		t.Fatalf("source empty or theme set")
	}
}

func TestAssetResolverSandboxes(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{"img/a.png": {Data: []byte("png")}}
	get := AssetResolver(fsys)
	if data, err := get("img/a.png"); err != nil || string(data) != "png" {
		t.Fatalf("valid: %v %q", err, data)
	}
	for _, bad := range []string{"", "../x", "a/../../x", "/abs", "a\\b", "C:x"} {
		if _, err := get(bad); err == nil {
			t.Fatalf("%q allowed", bad)
		}
	}
}

// staticBuilderDocument wraps a fragment in the ui.Document envelope. The
// page package cannot import ui (ui imports the root package, which imports
// page), so the envelope pieces are spelled out here. The builder escapes
// every brace it emits, so the result provably holds no template action.
func staticBuilderDocument(body string) string {
	return builderDocHead + "Static builder" + builderDocMid + body + builderDocTail
}

func staticSourcePage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{HTML: `<div id="start">start</div>`, Width: 640, Height: 480})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

// TestBuilderDocumentSkipsTemplateParse proves a builder document loaded
// through Load keeps no parsed template and redraws from its source while
// form sync still reaches bound controls.
func TestBuilderDocumentSkipsTemplateParse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := staticSourcePage(t)

	if p.tpl == nil || p.staticSource {
		t.Fatal("hand-written page did not keep its parsed template")
	}

	source := staticBuilderDocument(
		`<div id="root"><input id="email" type="text" data-bind="Email"><div id="label">Hello</div></div>`,
	)
	if err := p.Load(ctx, source); err != nil {
		t.Fatal(err)
	}

	if !p.staticSource {
		t.Fatal("builder document did not set staticSource")
	}
	if p.tpl != nil {
		t.Fatal("builder document parsed a template")
	}
	if p.tplProbeTpl != nil || p.tplProbeHas {
		t.Fatal("builder document kept a template probe")
	}
	if p.HTML() != source {
		t.Fatalf("HTML = %d bytes, want %d", len(p.HTML()), len(source))
	}
	if !p.hasBox("label") {
		t.Fatal("builder document did not lay out")
	}

	// Two clean redraws: a nil template that executed would panic, so the
	// successful frames prove the source was returned directly.
	for i := 0; i < 2; i++ {
		if err := p.Redraw(ctx); err != nil {
			t.Fatalf("redraw %d: %v", i, err)
		}
		if p.tpl != nil || !p.staticSource {
			t.Fatalf("redraw %d changed the static source state", i)
		}
	}

	// An identical Load reuses the static source with no parse.
	if err := p.Load(ctx, source); err != nil {
		t.Fatal(err)
	}
	if p.tpl != nil || !p.staticSource || p.tplProbeHas {
		t.Fatal("identical builder Load parsed a template")
	}

	// SetData still merges into the bound control through form sync, with
	// no template action needed to carry the value.
	p.SetData(&struct{ Email string }{Email: "ada@example.com"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("email"); got != "ada@example.com" {
		t.Fatalf("FormValue = %q", got)
	}
	if !strings.Contains(p.source, "ada@example.com") {
		t.Fatal("form sync did not rewrite the bound control")
	}
	if p.tpl != nil {
		t.Fatal("SetData redraw parsed a template")
	}
}

// TestNewBuilderDocumentSkipsTemplateParse proves the bootstrap path also
// recognises the builder envelope, so a cold page never parses a template it
// provably cannot execute.
func TestNewBuilderDocumentSkipsTemplateParse(t *testing.T) {
	t.Parallel()

	source := staticBuilderDocument(`<div id="root">cold</div>`)
	p, err := New(Config{HTML: source, Width: 640, Height: 480})
	if err != nil {
		t.Fatal(err)
	}
	if p.tpl != nil || !p.staticSource {
		t.Fatal("New parsed the builder document as a template")
	}
	if p.tplProbeTpl != nil || p.tplProbeHas {
		t.Fatal("New kept a template probe for the builder document")
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.hasBox("root") {
		t.Fatal("New builder document did not lay out")
	}
	if p.HTML() != source {
		t.Fatal("New changed the builder source")
	}
}

// TestBuilderDocumentSwitchesToTemplate proves loading a template after a
// builder document parses again and executes, history moves restore both
// modes, and ReplaceHTML in either direction follows the source.
func TestBuilderDocumentSwitchesToTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := staticSourcePage(t)
	builder := staticBuilderDocument(`<div id="builder">builder</div>`)

	if err := p.Load(ctx, builder); err != nil {
		t.Fatal(err)
	}
	if p.tpl != nil || !p.staticSource {
		t.Fatal("builder document did not stay static")
	}

	view := struct{ Name string }{Name: "Ada"}
	p.SetData(view)
	templateSource := `<p id="out">{{.Name}}</p>`
	if err := p.Load(ctx, templateSource); err != nil {
		t.Fatal(err)
	}

	if p.tpl == nil || p.staticSource {
		t.Fatal("template after a builder document did not parse")
	}
	if !p.tplProbeHas || p.tplProbeTpl != p.tpl {
		t.Fatal("template after a builder document did not enable execution")
	}
	if got := bftBox(t, p, "out"); got != "Ada" {
		t.Fatalf("template output = %q", got)
	}

	// A later data change executes the template again.
	view.Name = "Grace"
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := bftBox(t, p, "out"); got != "Grace" {
		t.Fatalf("template output after SetData = %q", got)
	}

	// Back to the builder entry: static again.
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if p.tpl != nil || !p.staticSource || !p.hasBox("builder") {
		t.Fatal("Back to a builder document did not restore the static source")
	}

	// Forward to the template entry: execution returns.
	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	if p.tpl == nil || p.staticSource {
		t.Fatal("Forward to a template did not re-enable execution")
	}
	if got := bftBox(t, p, "out"); got != "Grace" {
		t.Fatalf("template output after Forward = %q", got)
	}

	// ReplaceHTML in both directions follows the source type.
	if err := ReplaceHTML(ctx, p, builder); err != nil {
		t.Fatal(err)
	}
	if p.tpl != nil || !p.staticSource {
		t.Fatal("ReplaceHTML with a builder document parsed a template")
	}

	replaced := `<p id="out">{{.Name}}!</p>`
	if err := ReplaceHTML(ctx, p, replaced); err != nil {
		t.Fatal(err)
	}
	if p.tpl == nil || p.staticSource {
		t.Fatal("ReplaceHTML with a template did not parse")
	}
	if got := bftBox(t, p, "out"); got != "Grace!" {
		t.Fatalf("replaced template output = %q", got)
	}
}

func TestSafeFuncsReturnStrings(t *testing.T) {
	t.Parallel()
	strType := reflect.TypeOf("")
	for name, fn := range SafeFuncs() {
		if got := reflect.TypeOf(fn).Out(0); got != strType {
			t.Fatalf("%s returns %s", name, got)
		}
	}
}
