package page_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"github.com/chinmay-sawant/ownframe/ui"
)

// seamTree builds a deterministic large builder tree: many nested sections,
// text with escapable bytes and literal template braces, inputs, and click
// callbacks. Two calls with the same seed build equal but distinct trees.
func seamTree(seed, sections, rows int) *ui.Element {
	root := ui.Div().ID("seam-root")
	for s := 0; s < sections; s++ {
		sec := ui.Div().ID(fmt.Sprintf("sec-%d", s)).Class("sec")
		for r := 0; r < rows; r++ {
			id := fmt.Sprintf("cell-%d-%d-%d", seed, s, r)
			sec.Children(ui.Div().ID(id).Class("cell").Children(
				ui.Text(fmt.Sprintf("value <&> %d-%d {{literal}}", s, r)),
				ui.Button(fmt.Sprintf("Go {x} %d", r)).OnClick(func(context.Context) error { return nil }),
			))
		}
		if s%2 == 0 {
			sec.Children(ui.Input().ID(fmt.Sprintf("field-%d-%d", seed, s)).Attr("value", `edit "me" & you`))
		}
		root.Children(sec)
	}
	return root
}

// TestNavSeamBuilderOutputByteIdentical proves the direct-to-Document render
// emits byte-identical output to the fragment plus Document assembly on a
// generated large document, through both Mount and Update.
func TestNavSeamBuilderOutputByteIdentical(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const title = `Seam <Title> & "quotes" {{kept-literal}}`

	frag, err := ui.Render(seamTree(1, 40, 70))
	if err != nil {
		t.Fatal(err)
	}
	want := ui.Document(title, frag)

	view, err := ui.Mount(ownframe.Config{Title: title, Width: 800, Height: 600}, seamTree(1, 40, 70))
	if err != nil {
		t.Fatal(err)
	}
	if got := view.Page.HTML(); got != want {
		t.Fatalf("Mount output differs: got %d bytes, want %d bytes", len(got), len(want))
	}

	frag2, err := ui.Render(seamTree(2, 40, 70))
	if err != nil {
		t.Fatal(err)
	}
	want2 := ui.Document(title, frag2)
	if err := view.Update(ctx, seamTree(2, 40, 70)); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.HTML(); got != want2 {
		t.Fatalf("Update output differs: got %d bytes, want %d bytes", len(got), len(want2))
	}

	// Updates replace the current document without navigating.
	if err := view.Page.Back(ctx); !errors.Is(err, ownframe.ErrNoHistory) {
		t.Fatalf("Back after Update = %v, want ErrNoHistory", err)
	}
}

// TestNavSeamIdenticalSharesTemplate proves no-change navigations reuse the
// parsed template and the render cache, via parse and cascade counters: no
// timing. Load, ReplaceHTML, Back, and Forward to an identical source move
// no counters; distinct sources parse exactly once.
func TestNavSeamIdenticalSharesTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	third := `<div id="c">C</div>`
	p := navPage(t, first)
	parses, cascades := p.Stats().Parses, p.Stats().Cascades

	same := func(op string) {
		t.Helper()
		if got := p.Stats(); got.Parses != parses || got.Cascades != cascades {
			t.Fatalf("%s moved counters: parses %d->%d cascades %d->%d",
				op, parses, got.Parses, cascades, got.Cascades)
		}
	}

	if err := p.Load(ctx, first); err != nil {
		t.Fatal(err)
	}
	same("identical Load")

	if err := page.ReplaceHTML(ctx, p, first); err != nil {
		t.Fatal(err)
	}
	same("identical ReplaceHTML")

	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}
	parses, cascades = p.Stats().Parses, p.Stats().Cascades

	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}
	same("duplicate Load")

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != second {
		t.Fatalf("Back HTML = %q, want %q", p.HTML(), second)
	}
	same("identical Back")

	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != second {
		t.Fatalf("Forward HTML = %q, want %q", p.HTML(), second)
	}
	same("identical Forward")

	// Past the duplicate entries the source differs, so Back reparses once.
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	same("identical Back over duplicate")
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != first {
		t.Fatalf("Back HTML = %q, want %q", p.HTML(), first)
	}
	if got := p.Stats(); got.Parses != parses+1 || got.Cascades != cascades+1 {
		t.Fatalf("distinct Back parsed %d->%d cascaded %d->%d, want +1 and +1",
			parses, got.Parses, cascades, got.Cascades)
	}
	parses, cascades = p.Stats().Parses, p.Stats().Cascades

	if err := page.ReplaceHTML(ctx, p, third); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != third {
		t.Fatalf("ReplaceHTML HTML = %q, want %q", p.HTML(), third)
	}
	if got := p.Stats(); got.Parses != parses+1 || got.Cascades != cascades+1 {
		t.Fatalf("distinct ReplaceHTML parsed %d->%d cascaded %d->%d, want +1 and +1",
			parses, got.Parses, cascades, got.Cascades)
	}

	if err := page.ReplaceHTML(ctx, p, third); err != nil {
		t.Fatal(err)
	}
	parses, cascades = parses+1, cascades+1
	same("identical ReplaceHTML after retarget")
}

// TestHistorySeamStoredHashes proves each history entry carries its own Scan
// verdicts: the stored hash and delimiter flag follow Push, GoTo, and
// Retarget, and the fused scan agrees with strings.Contains.
func TestHistorySeamStoredHashes(t *testing.T) {
	t.Parallel()

	withDelim := `<div>{{.X}}</div>`
	plain := `<div id="a">A</div>`

	for _, s := range []string{"", "{{", "{{{", "a{{b}}c", "{x}", plain, withDelim} {
		_, delim := nav.Scan(s)
		if delim != strings.Contains(s, "{{") {
			t.Fatalf("Scan(%q) delim = %v, want %v", s, delim, !delim)
		}
		if nav.Hash(s) == 0 && s != "" {
			t.Fatalf("Hash(%q) = 0", s)
		}
	}

	h := nav.New(withDelim)
	if h.Current() != withDelim || h.CurrentHash() != nav.Hash(withDelim) || !h.CurrentDelim() {
		t.Fatal("New did not store the Scan verdicts")
	}

	h.Push(plain)
	if h.Current() != plain || h.CurrentHash() != nav.Hash(plain) || h.CurrentDelim() {
		t.Fatal("Push did not store the Scan verdicts")
	}

	h.Push(plain)
	h.GoTo(0)
	if h.Current() != withDelim || h.CurrentHash() != nav.Hash(withDelim) || !h.CurrentDelim() {
		t.Fatal("GoTo did not restore the entry verdicts")
	}
	if h.HashAt(1) != nav.Hash(plain) || h.DelimAt(1) {
		t.Fatal("HashAt/DelimAt mismatch on pushed entry")
	}
	if h.HashAt(2) != nav.Hash(plain) || h.DelimAt(2) {
		t.Fatal("HashAt/DelimAt mismatch on duplicate entry")
	}

	h.GoTo(2)
	replacement := `<span>{{.Y}}</span>`
	h.Retarget(replacement)
	if h.Current() != replacement || h.CurrentHash() != nav.Hash(replacement) || !h.CurrentDelim() {
		t.Fatal("Retarget did not store the fresh Scan verdicts")
	}
	if !h.CanBack() || h.CanForward() {
		t.Fatal("Retarget moved the history index")
	}
}

// TestHistorySeamCorrectness proves Load, Back, and Forward keep the exact
// history semantics: entries, forward drops, end errors, and the DOM guard
// for documents outside the builder envelope.
func TestHistorySeamCorrectness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	third := `<div id="c">C</div>`
	fourth := `<div id="d">D</div>`
	p := navPage(t, first)

	if err := p.Load(ctx, second); err != nil {
		t.Fatal(err)
	}
	if err := p.Load(ctx, third); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{second, first} {
		if err := p.Back(ctx); err != nil {
			t.Fatal(err)
		}
		if p.HTML() != want {
			t.Fatalf("Back HTML = %q, want %q", p.HTML(), want)
		}
	}
	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != second {
		t.Fatalf("Forward HTML = %q, want %q", p.HTML(), second)
	}
	if err := p.Load(ctx, fourth); err != nil {
		t.Fatal(err)
	}
	if err := p.Forward(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("Forward after replacement = %v, want ErrNoHistory", err)
	}
	if p.HTML() != fourth {
		t.Fatalf("HTML after replacement = %q, want %q", p.HTML(), fourth)
	}
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != second {
		t.Fatalf("Back HTML = %q, want %q", p.HTML(), second)
	}
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != first {
		t.Fatalf("Back HTML = %q, want %q", p.HTML(), first)
	}
	if err := p.Back(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("Back at start = %v, want ErrNoHistory", err)
	}

	// The guard skip is envelope-scoped: a deep hand-written document still
	// fails and leaves history untouched.
	deep := strings.Repeat("<div>", page.MaxDOMDepth+1) + "x" + strings.Repeat("</div>", page.MaxDOMDepth+1)
	if err := p.Load(ctx, deep); !errors.Is(err, page.ErrDOMTooDeep) {
		t.Fatalf("deep Load = %v, want ErrDOMTooDeep", err)
	}
	if p.HTML() != first {
		t.Fatalf("rejected Load moved history to %q", p.HTML())
	}
}
