package ownframe_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestBuilderEnvelopeDoesNotSelectStaticHTML(t *testing.T) {
	t.Parallel()

	p, err := ownframe.New(ownframe.Config{
		HTML:   ui.Document("Title", `<p id="output">{{.Name}}</p>`),
		Width:  320,
		Height: 120,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(struct{ Name string }{Name: "Ada"})
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "Ada")
}

func TestStaticHTMLHistoryAndReplacement(t *testing.T) {
	t.Parallel()

	const source = `<p>{{.Name}}</p>`
	ctx := t.Context()
	p, err := ownframe.New(ownframe.Config{HTML: source, StaticHTML: true, Width: 320, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(struct{ Name string }{Name: "Ada"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "{{.Name}}")

	if err := p.Load(ctx, source); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "Ada")

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "{{.Name}}")
	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "Ada")

	if err := p.Load(ctx, source, ownframe.LoadOptions{StaticHTML: true, ReplaceCurrent: true}); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "{{.Name}}")

	if err := p.Load(ctx, `<p>{{.Missing}}</p>`, ownframe.LoadOptions{ReplaceCurrent: true}); err == nil {
		t.Fatal("replacing with a template that fails execution succeeded")
	}
	if p.HTML() != source {
		t.Fatalf("failed replacement changed HTML to %q", p.HTML())
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	assertRenderedText(t, p, "{{.Name}}")
}

func TestStaticHTMLFileReloadKeepsConfiguredMode(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "page.html")
	write := func(html string) {
		t.Helper()

		if err := os.WriteFile(path, []byte(html), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(`<p>{{.Name}}</p>`)
	ctx := t.Context()
	p, err := ownframe.New(ownframe.Config{File: path, StaticHTML: true, Width: 320, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Load(ctx, `<p>Another page</p>`); err != nil {
		t.Fatal(err)
	}
	write(`<p>{{unclosed</p>`)
	if applied, err := p.Window().PollReload(ctx); err != nil || !applied {
		t.Fatalf("reload = %v, %v", applied, err)
	}
	assertRenderedText(t, p, "{{unclosed")

	write(`   `)
	if applied, err := p.Window().PollReload(ctx); err == nil || applied {
		t.Fatalf("blank reload = %v, %v, want rejection", applied, err)
	}
	assertRenderedText(t, p, "{{unclosed")
}

func TestTemplateFileReloadStillAcceptsBlank(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "page.html")
	if err := os.WriteFile(path, []byte("<p>Original</p>"), 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := ownframe.New(ownframe.Config{File: path, Width: 320, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if applied, err := p.Window().PollReload(ctx); err != nil || !applied {
		t.Fatalf("blank template reload = %v, %v", applied, err)
	}
	if p.HTML() != "" {
		t.Fatalf("blank template reload kept %q", p.HTML())
	}
	if err := p.Load(ctx, ""); !errors.Is(err, ownframe.ErrEmptyHTML) {
		t.Fatalf("Load of blank HTML = %v, want ErrEmptyHTML", err)
	}
}

func TestAdoptTreeRequiresExplicitStaticHTML(t *testing.T) {
	t.Parallel()

	const source = `<p>Previous document</p>`
	p, err := ownframe.New(ownframe.Config{HTML: source, Width: 320, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Load(t.Context(), `<p>Different document</p>`, ownframe.LoadOptions{AdoptTree: true}); err == nil {
		t.Fatal("AdoptTree accepted a source without StaticHTML")
	}
	if p.HTML() != source {
		t.Fatalf("refused adoption changed source to %q", p.HTML())
	}
}

func assertRenderedText(t *testing.T, p *ownframe.Page, want string) {
	t.Helper()

	list := p.Probe().Display()
	if list == nil {
		t.Fatal("no display list")
	}
	for _, op := range list.Ops {
		if op.Kind == ownframe.DisplayOpText && op.Text == want {
			return
		}
	}
	t.Fatalf("display did not paint %q", want)
}
