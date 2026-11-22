package page_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/ui"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	if err := page.ReplaceHTML(ctx, p, first, false); err != nil {
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

	if err := page.ReplaceHTML(ctx, p, third, false); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != third {
		t.Fatalf("ReplaceHTML HTML = %q, want %q", p.HTML(), third)
	}
	if got := p.Stats(); got.Parses != parses+1 || got.Cascades != cascades+1 {
		t.Fatalf("distinct ReplaceHTML parsed %d->%d cascaded %d->%d, want +1 and +1",
			parses, got.Parses, cascades, got.Cascades)
	}

	if err := page.ReplaceHTML(ctx, p, third, false); err != nil {
		t.Fatal(err)
	}
	parses, cascades = parses+1, cascades+1
	same("identical ReplaceHTML after retarget")
}

// TestHistorySeamCorrectness proves Load, Back, and Forward keep the exact
// history semantics: entries, forward drops, end errors, and the DOM guard
// for new documents.
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

// TestLoadRouteRejectOversizedSource checks that Load and Route enforce
// MaxSourceSize before parsing trusted template source.
func TestLoadRouteRejectOversizedSource(t *testing.T) {
	t.Parallel()

	screen := navPage(t, `<div id="a">A</div>`)
	tooBig := strings.Repeat("x", page.MaxSourceSize+1)

	if err := screen.Load(context.Background(), tooBig); !errors.Is(err, page.ErrSourceTooLarge) {
		t.Fatalf("Load err = %v, want %v", err, page.ErrSourceTooLarge)
	}
	if err := screen.Route("huge", tooBig); !errors.Is(err, page.ErrSourceTooLarge) {
		t.Fatalf("Route err = %v, want %v", err, page.ErrSourceTooLarge)
	}
}

func navPage(t *testing.T, html string) *page.Page {
	t.Helper()

	screen, err := page.New(page.Config{
		HTML:   html,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return screen
}

func navAt(t *testing.T, screen *page.Page, html, id string) {
	t.Helper()

	if screen.HTML() != html {
		t.Fatalf("html = %q", screen.HTML())
	}

	boxCenter(t, screen, id)
}

func TestLoadBackForward(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	screen := navPage(t, first)
	navAt(t, screen, first, "a")

	if err := screen.Load(ctx, second); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, second, "b")
	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, first, "a")
	if err := screen.Forward(ctx); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, second, "b")
	if err := screen.Forward(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("forward = %v", err)
	}

	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if err := screen.Back(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("back = %v", err)
	}
}

func TestRouteClickLoadsHTML(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	next := `<div id="b">B</div>`
	screen := navPage(t, `<div id="a" data-action="go">A</div>`)
	if err := screen.Route("go", `<div id="c">C</div>`); err != nil {
		t.Fatal(err)
	}
	if err := screen.Route("go", next); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, screen, "a")
	if err := screen.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	navAt(t, screen, next, "b")
}

func TestLoadRejectsBlankHTML(t *testing.T) {
	t.Parallel()

	screen := navPage(t, `<div id="a">A</div>`)
	err := screen.Load(context.Background(), " ")
	if !errors.Is(err, page.ErrEmptyHTML) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadAfterBackReplacesForwardHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	first := `<div id="a">A</div>`
	second := `<div id="b">B</div>`
	third := `<div id="c">C</div>`
	screen := navPage(t, first)
	if err := screen.Load(ctx, second); err != nil {
		t.Fatal(err)
	}
	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if err := screen.Load(ctx, third); err != nil {
		t.Fatal(err)
	}
	if err := screen.Back(ctx); err != nil {
		t.Fatal(err)
	}
	navAt(t, screen, first, "a")
	if err := screen.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	navAt(t, screen, third, "c")
	if err := screen.Forward(ctx); !errors.Is(err, page.ErrNoHistory) {
		t.Fatalf("forward after replacement = %v", err)
	}
}

func TestNewRejectsBothSources(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{
		HTML:   "<p>x</p>",
		File:   writeSource(t, "index.html", "<p>y</p>"),
		Width:  320,
		Height: 200,
	})
	if !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("err = %v", err)
	}
}

func TestNewRejectsMissingFile(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{
		File:   filepath.Join(t.TempDir(), "gone.html"),
		Width:  320,
		Height: 200,
	})
	if !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("err = %v", err)
	}
}

func TestNewBlankFileIsEmptyHTML(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{
		File:   writeSource(t, "blank.html", " \n"),
		Width:  320,
		Height: 200,
	})
	if !errors.Is(err, page.ErrEmptyHTML) {
		t.Fatalf("err = %v", err)
	}
}

func TestOversizeWrapsBothSentinels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	big := strings.Repeat("x", page.MaxSourceSize+1)
	if _, err := page.New(page.Config{HTML: big, Width: 320, Height: 200}); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("New err = %v", err)
	}
	s := navPage(t, `<div id="a">A</div>`)
	if err := s.Load(ctx, big); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Load err = %v", err)
	}
	if err := s.Route("huge", big); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Route err = %v", err)
	}
	if err := s.SetTheme(big); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Theme err = %v", err)
	}
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)
	rewriteSource(t, path, big)
	if _, err := p.PollReload(ctx); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("poll err = %v", err)
	}
}

// writeSource writes body into a fresh temp file and returns its path.
func writeSource(t *testing.T, name, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func newFilePage(t *testing.T, cfg page.Config) *page.Page {
	t.Helper()

	p, err := page.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TestNewReadsFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	html := `<div id="a">from disk</div>`
	fromFile := newFilePage(t, page.Config{
		File:   writeSource(t, "index.html", html),
		Width:  320,
		Height: 200,
	})
	fromString := newFilePage(t, page.Config{HTML: html, Width: 320, Height: 200})

	if fromFile.HTML() != html {
		t.Fatalf("HTML = %q", fromFile.HTML())
	}

	if err := fromFile.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := fromString.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	boxCenter(t, fromFile, "a")

	fileImg, err := png.Decode(bytes.NewReader(fromFile.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	stringImg, err := png.Decode(bytes.NewReader(fromString.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	if !samePicture(fileImg, stringImg) {
		t.Fatal("file page and string page painted differently")
	}
}

func TestNewReadsThemeFile(t *testing.T) {
	t.Parallel()

	p := newFilePage(t, page.Config{
		HTML:      themePage,
		ThemeFile: writeSource(t, "theme.css", `#swatch{background:#ff0000}`),
		Width:     320,
		Height:    200,
	})

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if r, g, b := swatchRGB(t, p); r < 1 || g > 0 || b > 0 {
		t.Fatalf("swatch = %v,%v,%v, want red", r, g, b)
	}
}
