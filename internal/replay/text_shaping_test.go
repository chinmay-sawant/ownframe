package replay

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// shapedOp returns the first text op of a one-line page with style applied.
func shapedOp(t *testing.T, style, content string) *layout.DisplayOp {
	t.Helper()
	d, err := render.DisplayList(context.Background(),
		`<div style="font-size:24px;`+style+`">`+content+`</div>`, 400, 80)
	if err != nil {
		t.Fatal(err)
	}
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpText && d.Ops[i].Text != "" {
			return &d.Ops[i]
		}
	}
	t.Fatal("the page painted no text")

	return nil
}

// TestPagesWithShapingOptionsRedraw is the page-level repro: each declaration
// used to fail Redraw with "bitmap: unsupported text shaping options".
func TestPagesWithShapingOptionsRedraw(t *testing.T) {
	t.Parallel()
	for _, style := range []string{
		"font-variant:small-caps",
		"font-variant-numeric:tabular-nums",
		"font-variant-ligatures:none",
		"font-variant-position:super",
		"font-kerning:none",
		"text-autospace:normal",
		"font-variant:small-caps;text-autospace:normal;letter-spacing:1px",
	} {
		p, err := page.New(page.Config{HTML: `<div style="` + style + `">Hello 123 日本語abc</div>`, Width: 300, Height: 100})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Redraw(context.Background()); err != nil {
			t.Fatalf("%s: Redraw: %v", style, err)
		}
		if p.Display() == nil || p.Image() != nil {
			t.Fatalf("%s: the page should replay, not fall back to a bitmap", style)
		}
	}
}

// TestFontFeaturesReachTheFace pins the wiring: kerning off must measure wider
// than the default on the same run, equal faces share one cached instance, and
// faces that differ only in features never share one.
func TestFontFeaturesReachTheFace(t *testing.T) {
	t.Parallel()
	var r Renderer
	plain := shapedOp(t, "", "AVAVAVAV")
	unkerned := shapedOp(t, "font-kerning:none", "AVAVAVAV")
	if unkerned.FontFeatures() == "" {
		t.Fatal("font-kerning:none carried no feature list")
	}
	pf, uf := r.textFace(plain), r.textFace(unkerned)
	if pf == uf {
		t.Fatal("faces with different features share one instance")
	}
	if r.textFace(unkerned) != uf {
		t.Fatal("an equal op did not reuse its cached face")
	}
	pw, _ := text.Measure(plain.Text, pf, 0)
	uw, _ := text.Measure(unkerned.Text, uf, 0)
	if uw <= pw {
		t.Fatalf("kerning off measured %.2f, want wider than the default %.2f", uw, pw)
	}
}

func glyphIDs(content string, face *text.GoTextFace) []uint32 {
	glyphs := text.AppendLazyGlyphs(nil, content, face, nil)
	ids := make([]uint32, len(glyphs))
	for i, g := range glyphs {
		ids[i] = g.GID
	}
	return ids
}

// TestFontLanguageOverrideReachesTheFace verifies that OpenType language tags
// from CSS font-language-override (such as "SRB") are mapped to BCP 47 before
// reaching Ebiten's GoTextFace so HarfBuzz applies the font's locl table.
func TestFontLanguageOverrideReachesTheFace(t *testing.T) {
	t.Parallel()
	var r Renderer
	plain := shapedOp(t, "font-family:'DejaVu Sans'", "б")
	srb := shapedOp(t, "font-family:'DejaVu Sans';font-language-override:'SRB'", "б")
	if srb.TextLanguage() != "SRB" {
		t.Fatalf("TextLanguage = %q, want SRB", srb.TextLanguage())
	}
	pf, sf := r.textFace(plain), r.textFace(srb)
	if sf.Language.String() != "sr" {
		t.Fatalf("GoTextFace.Language = %q, want sr", sf.Language.String())
	}
	pg, sg := glyphIDs(plain.Text, pf), glyphIDs(srb.Text, sf)
	if reflect.DeepEqual(pg, sg) {
		t.Fatalf("SRB locl did not change glyph IDs for DejaVu Sans 'б': plain=%v srb=%v", pg, sg)
	}
}

// TestFontLigaturesAndAlternatesReachTheFace verifies that DejaVu Sans's GSUB
// liga and salt tables visibly change shaped glyph output through replay.
func TestFontLigaturesAndAlternatesReachTheFace(t *testing.T) {
	t.Parallel()
	var r Renderer

	ligPlain := shapedOp(t, "font-family:'DejaVu Sans'", "fi fl")
	ligOffFeat := shapedOp(t, "font-family:'DejaVu Sans';font-feature-settings:'liga' 0", "fi fl")
	ligOffVar := shapedOp(t, "font-family:'DejaVu Sans';font-variant-ligatures:none", "fi fl")

	pg := glyphIDs(ligPlain.Text, r.textFace(ligPlain))
	fg := glyphIDs(ligOffFeat.Text, r.textFace(ligOffFeat))
	vg := glyphIDs(ligOffVar.Text, r.textFace(ligOffVar))
	if len(pg) >= len(fg) {
		t.Fatalf("font-feature-settings:\"liga\" 0 did not break ligatures: plain=%v off=%v", pg, fg)
	}
	if len(pg) >= len(vg) {
		t.Fatalf("font-variant-ligatures:none did not break ligatures: plain=%v off=%v", pg, vg)
	}

	altPlain := shapedOp(t, "font-family:'DejaVu Sans'", "I a l")
	altSalt := shapedOp(t, "font-family:'DejaVu Sans';font-variant-alternates:stylistic(1)", "I a l")
	apg := glyphIDs(altPlain.Text, r.textFace(altPlain))
	asg := glyphIDs(altSalt.Text, r.textFace(altSalt))
	if reflect.DeepEqual(apg, asg) {
		t.Fatalf("font-variant-alternates:stylistic(1) did not change glyph IDs: plain=%v salt=%v", apg, asg)
	}
}

// TestRasterKeyCarriesShapingOptions keeps one baked run from being reused for
// a run that differs only in features or autospace gap.
func TestRasterKeyCarriesShapingOptions(t *testing.T) {
	t.Parallel()
	a := textRasterKey{content: "x"}
	b, c := a, a
	b.features = `"smcp" 1`
	c.gap = 2
	if a == b || a == c || b == c {
		t.Fatal("raster keys ignore features or the autospace gap")
	}
}

func TestWalkRun(t *testing.T) {
	t.Parallel()
	var r Renderer
	op := shapedOp(t, "", "x")
	face := r.textFace(op)
	type piece struct {
		s   string
		off float64
	}
	collect := func(content string, spacing, gap float64) []piece {
		var out []piece
		walkRun(face, content, spacing, gap, func(s string, off float64) { out = append(out, piece{s, off}) })

		return out
	}

	if got := collect("日本語abc", 0, 0); !reflect.DeepEqual(got, []piece{{"日本語abc", 0}}) {
		t.Fatalf("no spacing = %v, want one piece", got)
	}

	const gap = 3.0
	got := collect("日本語abc日", 0, gap)
	if len(got) != 3 || got[0].s != "日本語" || got[1].s != "abc" || got[2].s != "日" {
		t.Fatalf("autospace split = %v, want 日本語 | abc | 日", got)
	}
	w0, _ := text.Measure("日本語", face, 0)
	w1, _ := text.Measure("abc", face, 0)
	if got[0].off != 0 || math.Abs(got[1].off-(w0+gap)) > 1e-9 || math.Abs(got[2].off-(w0+gap+w1+gap)) > 1e-9 {
		t.Fatalf("autospace offsets = %v, want 0, %.2f, %.2f", got, w0+gap, w0+gap+w1+gap)
	}

	// With letter spacing every glyph is its own piece, and the gap lands once
	// before the glyph that opens a boundary.
	spaced := collect("語ab", 1, gap)
	if len(spaced) != 3 {
		t.Fatalf("letter-spaced = %v, want one piece per glyph", spaced)
	}
	wg, _ := text.Measure("語", face, 0)
	if want := wg + 1 + gap; math.Abs(spaced[1].off-want) > 1e-9 {
		t.Fatalf("glyph after the boundary at %.2f, want %.2f", spaced[1].off, want)
	}
	wa, _ := text.Measure("a", face, 0)
	if want := wg + 1 + gap + wa + 1; math.Abs(spaced[2].off-want) > 1e-9 {
		t.Fatalf("second glyph at %.2f, want %.2f (no gap inside Latin text)", spaced[2].off, want)
	}
}

func TestBlinklessOTFontSynthesisNumericAndRuby(t *testing.T) {
	t.Parallel()
	var r Renderer

	// 1. font-synthesis-weight: none disables fake bold on regular-only Blinkless OT.
	boldAuto := shapedOp(t, "font-family:'Blinkless OT';font-weight:700", "Bold text")
	boldNone := shapedOp(t, "font-family:'Blinkless OT';font-weight:700;font-synthesis-weight:none", "Bold text")
	if !layout.DisplayFakeBold(boldAuto) {
		t.Fatal("expected FakeBold=true on regular-only Blinkless OT with font-weight:700")
	}
	if layout.DisplayFakeBold(boldNone) {
		t.Fatal("expected FakeBold=false with font-synthesis-weight:none")
	}

	// 2. font-variant-numeric: tabular-nums widens narrow '1' to tabular width.
	numProp := shapedOp(t, "font-family:'Blinkless OT'", "1111")
	numTab := shapedOp(t, "font-family:'Blinkless OT';font-variant-numeric:tabular-nums", "1111")
	wProp, _ := text.Measure(numProp.Text, r.textFace(numProp), 0)
	wTab, _ := text.Measure(numTab.Text, r.textFace(numTab), 0)
	if wTab <= wProp*1.5 {
		t.Fatalf("tabular-nums width %.2f should be much wider than proportional %.2f", wTab, wProp)
	}

	// 3. font-variant-east-asian: ruby substitutes compact .ruby glyphs.
	eaPlain := shapedOp(t, "font-family:'Blinkless OT'", "あいう")
	eaRuby := shapedOp(t, "font-family:'Blinkless OT';font-variant-east-asian:ruby", "あいう")
	gp := glyphIDs(eaPlain.Text, r.textFace(eaPlain))
	gr := glyphIDs(eaRuby.Text, r.textFace(eaRuby))
	if reflect.DeepEqual(gp, gr) {
		t.Fatalf("ruby feature did not change glyph IDs: plain=%v ruby=%v", gp, gr)
	}
}
