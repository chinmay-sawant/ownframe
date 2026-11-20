package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/page/source"
	"net/url"
	"strings"
	"testing"
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

	if !hasBox(p, id) {
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
			case tc.wantDeep && !errors.Is(err, source.ErrDOMTooDeep):
				t.Fatalf("%q: got %v, want depth error", tc.doc, err)
			case tc.wantMany && !errors.Is(err, source.ErrTooManyNodes):
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
	if err := source.CheckDOM(deep, source.MaxDOMDepth, source.MaxDOMNodes); !errors.Is(err, source.ErrDOMTooDeep) {
		t.Fatalf("deep: got %v, want depth error", err)
	}

	wide := strings.Repeat("<i>x</i>", 30)
	if err := source.CheckDOM(wide, 10, 10); !errors.Is(err, source.ErrTooManyNodes) {
		t.Fatalf("wide: got %v, want nodes error", err)
	}

	upperDeep := strings.Repeat("<DIV>", 300)
	if err := source.CheckDOM(upperDeep, source.MaxDOMDepth, source.MaxDOMNodes); !errors.Is(err, source.ErrDOMTooDeep) {
		t.Fatalf("upper deep: got %v, want depth error", err)
	}

	if n := testing.AllocsPerRun(20, func() { _ = source.CheckDOM(doc, 4, 500) }); n != 0 {
		t.Fatalf("CheckDOM allocs = %v, want 0", n)
	}
}

// TestClassifyLinkSharedCorpus drives one allow-plus-deny table through both
// link entries: parseOutboundLink and DispatchDeepLink share classifyLink,
// so allow, javascript deny, empty-host http, and relative fallthrough must
// agree on both.
func TestClassifyLinkSharedCorpus(t *testing.T) {
	deep := func(target string) (*url.URL, error) {
		p, err := New(Config{HTML: `<p>x</p>`, Width: 10, Height: 10,
			DeepLinkSchemes: []string{"https", "mailto", "myapp"}})
		if err != nil {
			t.Fatal(err)
		}
		var got *url.URL
		p.Handle(Handlers{DeepLink: func(_ context.Context, link *url.URL) error {
			got = link
			return nil
		}})
		err = p.DispatchDeepLink(context.Background(), target)
		return got, err
	}
	for _, tc := range []struct {
		target   string
		outLink  bool
		outErr   error
		deepLink bool
		deepErr  error
	}{
		{"https://example.com/x", true, nil, true, nil},
		{"mailto:a@b.c", true, nil, true, nil},
		{"myapp://open/chat", false, ErrUnsupportedLinkScheme, true, nil},
		{"javascript:alert(1)", false, ErrUnsupportedLinkScheme, false, ErrUnsupportedLinkScheme},
		{"http://", false, ErrInvalidLink, false, ErrInvalidLink},
		{"settings", false, nil, false, ErrInvalidLink},
	} {
		link, err := parseOutboundLink(tc.target)
		if tc.outErr != nil {
			if !errors.Is(err, tc.outErr) || link != nil {
				t.Errorf("outbound(%q) = %v, %v, want %v", tc.target, link, err, tc.outErr)
			}
		} else if err != nil || (link == nil) != !tc.outLink {
			t.Errorf("outbound(%q) = %v, %v, want link %v", tc.target, link, err, tc.outLink)
		}
		got, err := deep(tc.target)
		if tc.deepErr != nil {
			if !errors.Is(err, tc.deepErr) || got != nil {
				t.Errorf("deep(%q) = %v, %v, want %v", tc.target, got, err, tc.deepErr)
			}
		} else if err != nil || (got == nil) != !tc.deepLink {
			t.Errorf("deep(%q) = %v, %v, want link %v", tc.target, got, err, tc.deepLink)
		}
	}
}

func TestDispatchDeepLink(t *testing.T) {
	p := deepLinkPage(t, []string{"myapp"})
	var got *url.URL
	p.Handle(Handlers{DeepLink: func(_ context.Context, link *url.URL) error {
		got = link
		return nil
	}})
	if err := p.DispatchDeepLink(context.Background(), "MyApp://open/chat?id=3"); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Scheme != "myapp" || got.Host != "open" || got.Path != "/chat" {
		t.Fatalf("link = %#v", got)
	}
}

func TestDispatchDeepLinkRejectsBadTargets(t *testing.T) {
	p := deepLinkPage(t, []string{"myapp"})
	p.Handle(Handlers{DeepLink: func(context.Context, *url.URL) error { return nil }})
	for _, test := range []struct {
		target string
		want   error
	}{
		{"other://open", ErrUnsupportedLinkScheme},
		{"javascript:alert(1)", ErrUnsupportedLinkScheme},
		{"settings", ErrInvalidLink},
		{"myapp://", ErrInvalidLink},
	} {
		if err := p.DispatchDeepLink(context.Background(), test.target); !errors.Is(err, test.want) {
			t.Errorf("DispatchDeepLink(%q) = %v, want %v", test.target, err, test.want)
		}
	}
}

func TestDeepLinkConfigAndMissingHandler(t *testing.T) {
	if _, err := New(Config{HTML: `<p>x</p>`, Width: 100, Height: 100, DeepLinkSchemes: []string{"javascript"}}); !errors.Is(err, ErrBadLinkScheme) {
		t.Fatalf("bad scheme error = %v", err)
	}
	p := deepLinkPage(t, []string{"myapp"})
	if err := p.DispatchDeepLink(context.Background(), "myapp://open"); !errors.Is(err, ErrNoDeepLinkHandler) {
		t.Fatalf("missing handler error = %v", err)
	}
}

func deepLinkPage(t *testing.T, schemes []string) *Page {
	t.Helper()
	p, err := New(Config{HTML: `<p>Page</p>`, Width: 100, Height: 100, DeepLinkSchemes: schemes})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestValidateOpenTarget(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"https://example.test/p", "http://a.b/", "mailto:a@b.c"} {
		if _, err := ValidateOpenTarget(ok); err != nil {
			t.Fatalf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"javascript:alert(1)", "file:///etc/passwd", "data:text/html,hi", "  https://a.b/ ", "relative/path"} {
		if _, err := ValidateOpenTarget(bad); err == nil {
			t.Fatalf("%s allowed", strings.TrimSpace(bad))
		}
	}
}

func TestClickHandsValidatedOutboundLinkToApp(t *testing.T) {
	p, err := New(Config{HTML: `<a href="https://example.test/path">Open</a>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	var got *url.URL
	p.Handle(Handlers{OpenLink: func(_ context.Context, link *url.URL) error {
		got = link
		return nil
	}})
	for i := range p.draw.Display.Ops {
		op := &p.draw.Display.Ops[i]
		if op.Kind != DisplayOpLinkURI {
			continue
		}
		x := (op.X + op.W/2) / p.draw.Display.PointsPerPixel
		y := (op.Y + op.H/2) / p.draw.Display.PointsPerPixel
		if err = p.Click(context.Background(), x, y); err != nil {
			t.Fatal(err)
		}
		break
	}
	if got == nil || got.Host != "example.test" || got.Path != "/path" {
		t.Fatalf("opened link = %#v", got)
	}
}
