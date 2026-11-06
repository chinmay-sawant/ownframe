package page

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/source"
)

func navCacheNewPage(t *testing.T, html string) *Page {
	t.Helper()

	p, err := New(Config{HTML: html, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func navCacheMustHaveBox(t *testing.T, p *Page, id string) {
	t.Helper()

	if !p.hasBox(id) {
		t.Fatalf("missing box %q", id)
	}
}

// TestNavIdenticalLoadSharesTemplate proves an identical Load reuses the
// parsed template and keeps the render cache: pointer equality plus
// unchanged parse counters, no timing.
func TestNavIdenticalLoadSharesTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	p := navCacheNewPage(t, first)

	beforeTpl := p.tpl
	beforeCache := p.cache
	beforeParses := p.stats.parses
	beforeCascades := p.stats.cascades

	if err := p.Load(ctx, first); err != nil {
		t.Fatal(err)
	}

	if p.tpl != beforeTpl {
		t.Fatal("identical Load reparsed the template")
	}

	if p.cache != beforeCache {
		t.Fatal("identical Load dropped the render cache")
	}

	if p.stats.parses != beforeParses || p.stats.cascades != beforeCascades {
		t.Fatalf("parses = %d cascades = %d, want %d and %d",
			p.stats.parses, p.stats.cascades, beforeParses, beforeCascades)
	}

	if p.HTML() != first {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "a")

	second := `<div id="b">B</div>`
	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}

	if p.tpl == beforeTpl {
		t.Fatal("different Load kept the old template")
	}

	if p.stats.parses == beforeParses {
		t.Fatal("different Load did not restyle")
	}

	if p.HTML() != second {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "b")
}

// TestNavIdenticalBackForwardSharesTemplate proves Back and Forward to an
// identical source reuse the parsed template without restyling.
func TestNavIdenticalBackForwardSharesTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	p := navCacheNewPage(t, first)

	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}

	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}

	beforeTpl := p.tpl
	beforeCache := p.cache
	beforeParses := p.stats.parses

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if p.tpl != beforeTpl {
		t.Fatal("identical Back reparsed the template")
	}

	if p.cache != beforeCache {
		t.Fatal("identical Back dropped the render cache")
	}

	if p.stats.parses != beforeParses {
		t.Fatalf("parses = %d, want %d", p.stats.parses, beforeParses)
	}

	if p.HTML() != second {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "b")

	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}

	if p.tpl != beforeTpl {
		t.Fatal("identical Forward reparsed the template")
	}

	if p.HTML() != second {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "b")
}

// TestNavBackForwardCorrectness proves distinct Back and Forward entries
// still draw the right document.
func TestNavBackForwardCorrectness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	p := navCacheNewPage(t, first)

	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}

	if p.HTML() != second {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "b")

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if p.HTML() != first {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "a")

	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}

	if p.HTML() != second {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "b")
}

// TestHistoryRetargetIdenticalKeepsTemplate proves ReplaceHTML with the
// same source reuses the template and the cache, while a new source
// reparses.
func TestHistoryRetargetIdenticalKeepsTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	p := navCacheNewPage(t, first)

	beforeTpl := p.tpl
	beforeCache := p.cache
	beforeParses := p.stats.parses

	if err := ReplaceHTML(ctx, p, first); err != nil {
		t.Fatal(err)
	}

	if p.tpl != beforeTpl {
		t.Fatal("identical Retarget reparsed the template")
	}

	if p.cache != beforeCache {
		t.Fatal("identical Retarget dropped the render cache")
	}

	if p.stats.parses != beforeParses {
		t.Fatalf("parses = %d, want %d", p.stats.parses, beforeParses)
	}

	navCacheMustHaveBox(t, p, "a")

	second := `<div id="b">B</div>`
	if err := ReplaceHTML(ctx, p, second); err != nil {
		t.Fatal(err)
	}

	if p.tpl == beforeTpl {
		t.Fatal("different Retarget kept the old template")
	}

	if p.HTML() != second {
		t.Fatalf("HTML = %q", p.HTML())
	}

	navCacheMustHaveBox(t, p, "b")
}

// TestCheckDOMFoldVerdicts proves the allocation-free guard returns the
// same verdicts as the lowercasing parser, including adversarial tag
// cases: uppercase voids, quoted angle brackets, comments, self-closes,
// unfinished tags, and closes.
func TestCheckDOMFoldVerdicts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		doc      string
		depth    int
		nodes    int
		wantDeep bool
		wantMany bool
	}{
		{name: "flat", doc: `<p>x</p>`, depth: 4, nodes: 4},
		{name: "nested over", doc: `<div><div>`, depth: 1, nodes: 10, wantDeep: true},
		{name: "nodes over", doc: `<i></i><i></i>`, depth: 10, nodes: 1, wantMany: true},
		{name: "void lower", doc: `<br><br><div></div>`, depth: 1, nodes: 10},
		{name: "void upper", doc: `<BR><BR><DIV></DIV>`, depth: 1, nodes: 10},
		{name: "void mixed self", doc: `<Br/><Img src="a>b">`, depth: 1, nodes: 10},
		{name: "void depth zero", doc: `<INPUT><BR><WBR>`, depth: 0, nodes: 10},
		{name: "self upper", doc: `<DIV/>`, depth: 1, nodes: 10},
		{name: "comment hides", doc: `<!-- <div><div> --><p></p>`, depth: 1, nodes: 10},
		{name: "comment upper", doc: `<!-- <DIV> --><BR>`, depth: 1, nodes: 10},
		{name: "decl", doc: `<!DOCTYPE html><p></p>`, depth: 1, nodes: 10},
		{name: "close upper", doc: `<DIV></DIV>`, depth: 1, nodes: 10},
		{name: "close no depth", doc: `</div>`, depth: 1, nodes: 10},
		{name: "empty tag", doc: `<>`, depth: 1, nodes: 10},
		{name: "attr quote angle", doc: `<div title="a>b"><p></p></div>`, depth: 4, nodes: 10},
		{name: "single quote angle", doc: `<div title='a>b'><p></p></div>`, depth: 4, nodes: 10},
		{name: "unterminated quote", doc: `<div title="oops><p></p>`, depth: 4, nodes: 10},
		{name: "missing close", doc: `<div`, depth: 1, nodes: 10},
		{name: "missing close over", doc: `<div`, depth: 0, nodes: 10, wantDeep: true},
		{name: "void with slash", doc: `<input/>`, depth: 0, nodes: 10},
		{name: "nonvoid with slash", doc: `<div/>`, depth: 1, nodes: 10},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := source.CheckDOM(tc.doc, tc.depth, tc.nodes)
			switch {
			case tc.wantDeep && err != source.ErrDOMTooDeep:
				t.Fatalf("%q: got %v, want depth error", tc.doc, err)
			case tc.wantMany && err != source.ErrTooManyNodes:
				t.Fatalf("%q: got %v, want nodes error", tc.doc, err)
			case !tc.wantDeep && !tc.wantMany && err != nil:
				t.Fatalf("%q: got %v, want nil", tc.doc, err)
			}
		})
	}
}

// TestCheckDOMGeneratedDoc proves verdicts on a generated document with
// mixed nesting, voids, comments, uppercase names, and quoted attributes.
func TestCheckDOMGeneratedDoc(t *testing.T) {
	// No t.Parallel: AllocsPerRun cannot run inside a parallel test.

	var b strings.Builder
	b.WriteString(`<!-- generated -->`)
	b.WriteString(`<!DOCTYPE html>`)
	for i := 0; i < 50; i++ {
		b.WriteString(`<DIV id="n">`)
		b.WriteString(`<BR>`)
		b.WriteString(`<img src="a>b">`)
		b.WriteString(`<span title='x>y'>ok</span>`)
		b.WriteString(`</DIV>`)
	}
	doc := b.String()

	if err := source.CheckDOM(doc, 4, 500); err != nil {
		t.Fatalf("generated doc: %v", err)
	}

	deep := strings.Repeat("<div>", 300)
	if err := source.CheckDOM(deep, source.MaxDOMDepth, source.MaxDOMNodes); err != source.ErrDOMTooDeep {
		t.Fatalf("deep: got %v, want depth error", err)
	}

	wide := strings.Repeat("<i>x</i>", 30)
	if err := source.CheckDOM(wide, 10, 10); err != source.ErrTooManyNodes {
		t.Fatalf("wide: got %v, want nodes error", err)
	}

	upperDeep := strings.Repeat("<DIV>", 300)
	if err := source.CheckDOM(upperDeep, source.MaxDOMDepth, source.MaxDOMNodes); err != source.ErrDOMTooDeep {
		t.Fatalf("upper deep: got %v, want depth error", err)
	}

	if n := testing.AllocsPerRun(20, func() { _ = source.CheckDOM(doc, 4, 500) }); n != 0 {
		t.Fatalf("CheckDOM allocs = %v, want 0", n)
	}
}
