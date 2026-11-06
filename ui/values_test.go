package ui

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"strconv"
	"strings"
	"testing"
)

// TestScalarValueFormattingParity pins the exact CSS text of every
// single-number constructor and of the composite values that embed one. The
// element stores lengths as numbers now, so a formatting drift would show up
// only at render time.
func TestScalarValueFormattingParity(t *testing.T) {
	// Negating a runtime zero keeps the sign bit, which Px must preserve.
	var zero float64
	negativeZero := -zero
	cases := []struct {
		name string
		root *Element
		want string
	}{
		{"zero length", Div().Width(Px(0)), `<div style="width:0px;"></div>`},
		{"negative zero length", Div().Width(Px(negativeZero)), `<div style="width:-0px;"></div>`},
		{"fractional length", Div().Width(Px(0.1)), `<div style="width:0.1px;"></div>`},
		{"large length", Div().Width(Px(1e6)), `<div style="width:1000000px;"></div>`},
		{"percentage", Div().Width(Percent(12.5)), `<div style="width:12.5%;"></div>`},
		{"viewport height", Div().Height(Vh(100)), `<div style="height:100vh;"></div>`},
		{"rem", Div().Width(Rem(1.5)), `<div style="width:1.5rem;"></div>`},
		{"unitless", Div().Opacity(Number(0.5)), `<div style="opacity:0.5;"></div>`},
		{"edges", Div().Padding(Edges(Px(1), Px(2), Px(3), Px(4))), `<div style="padding:1px 2px 3px 4px;"></div>`},
		{"variable", Div().Color(Var("--accent", "#fff")), `<div style="color:var(--accent, #fff);"></div>`},
		{"variable without fallback", Div().Color(Var("--accent")), `<div style="color:var(--accent);"></div>`},
		{"solid border", Div().Border(SolidBorder(Px(1), Hex("#abcdef"))), `<div style="border:1px solid #abcdef;"></div>`},
		{"translate", Div().Transform(TranslateY(Px(-20))), `<div style="transform:translateY(-20px);"></div>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Render(tc.root)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("Render = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestScalarAppendMatchesFormatFloat checks the numeric fast path against
// the formatting the old string-based constructor used, including negative
// zero and values whose fixed-point spelling is long.
func TestScalarAppendMatchesFormatFloat(t *testing.T) {
	const (
		tenth        = 0.1
		third        = 1.0 / 3.0
		million      = 1e6
		millionth    = 1e-6
		longFraction = 123456789.125
	)
	var zero float64
	values := []float64{
		zero, -zero, tenth, third, million, millionth,
		longFraction, math.MaxFloat64, math.SmallestNonzeroFloat64,
	}
	units := []unitKind{unitNone, unitPx, unitPercent, unitVh, unitRem}
	for _, n := range values {
		for _, unit := range units {
			want := strconv.FormatFloat(n, 'f', -1, 64) + unit.suffix()
			if got := string(appendScalar(nil, n, unit)); got != want {
				t.Errorf("appendScalar(%v, %v) = %q, want %q", n, unit, got, want)
			}
			if got := formatScalar(n, unit); got != want {
				t.Errorf("formatScalar(%v, %v) = %q, want %q", n, unit, got, want)
			}
		}
		want := strconv.FormatFloat(n, 'f', -1, 64) + "px"
		if got, err := (Length{valueData: scalar(n, unitPx)}).cssText(); err != nil || got != want {
			t.Errorf("Length cssText(%v) = %q, %v; want %q", n, got, err, want)
		}
	}
}

// TestStyleEntriesOverflowInlineCapacity exercises the spill from the inline
// declarations into the element's overflow slice and the sorted output both
// paths share.
func TestStyleEntriesOverflowInlineCapacity(t *testing.T) {
	e := Div().Width(Px(1)).Height(Px(2)).Padding(Px(3)).Margin(Px(4)).
		Gap(Px(5)).FontSize(Px(6)).Color(Hex("#fff"))
	if got, want := e.styles.len(), inlineStyleEntries+3; got != want {
		t.Fatalf("styles len = %d, want %d", got, want)
	}
	e.Width(Px(9)).FontSize(Px(11))
	if got, want := e.styles.len(), 7; got != want {
		t.Fatalf("styles len after replace = %d, want %d", got, want)
	}
	got, err := Render(e)
	if err != nil {
		t.Fatal(err)
	}
	const want = `<div style="color:#fff;font-size:11px;gap:5px;height:2px;margin:4px;padding:3px;width:9px;"></div>`
	if got != want {
		t.Fatalf("Render = %q, want %q", got, want)
	}
}

// TestStyleSetEqualityIgnoresOrder checks the patch comparison: declarations
// set in a different order still match, a changed overflow entry does not, and
// a numeric value spelled like raw text still matches.
func TestStyleSetEqualityIgnoresOrder(t *testing.T) {
	build := func(fontSize float64) *Element {
		return Div().Width(Px(1)).Height(Px(2)).Padding(Px(3)).Margin(Px(4)).
			Gap(Px(5)).FontSize(Px(fontSize)).Color(Hex("#fff"))
	}
	first, second := build(10), build(10)
	if !first.styles.equal(&second.styles) {
		t.Fatal("same declarations in insertion order compare unequal")
	}
	reordered := Div().Color(Hex("#fff")).FontSize(Px(10)).Gap(Px(5)).
		Margin(Px(4)).Padding(Px(3)).Height(Px(2)).Width(Px(1))
	if !first.styles.equal(&reordered.styles) {
		t.Fatal("same declarations in a different order compare unequal")
	}
	if first.styles.equal(&build(12).styles) {
		t.Fatal("a changed overflow declaration compares equal")
	}
	var empty Element
	if first.styles.equal(&empty.styles) {
		t.Fatal("declarations compare equal to an empty set")
	}
	numeric := Div().Width(Px(10))
	raw := Div().Width(RawValue("10px"))
	if !numeric.styles.equal(&raw.styles) {
		t.Fatal("Px(10) and RawValue(\"10px\") compare unequal")
	}
	var zero float64
	negativeZero := Div().Width(Px(-zero))
	zeroWidth := Div().Width(Px(0))
	if negativeZero.styles.equal(&zeroWidth.styles) {
		t.Fatal("Px(-0) and Px(0) compare equal")
	}
}

func TestTypedStylesAndCustomProperties(t *testing.T) {
	root := Div().Display(DisplayFlex).FlexDirection(FlexDirectionColumn).
		JustifyContent(JustifyContentCenter).Padding(Edges(Px(8), Px(16))).
		Gap(Rem(1)).Width(Percent(100)).Opacity(Number(0.5)).
		CustomProperty("--Accent", "#20242b").BackgroundColor(Var("--Accent", "#fff")).
		Color(Hex("#abc")).Border(RawValue("1px solid #ddd"))
	text, err := Render(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"display:flex;", "flex-direction:column;", "justify-content:center;",
		"padding:8px 16px;", "gap:1rem;", "width:100%;", "opacity:0.5;",
		"--Accent:#20242b;", "background-color:var(--Accent, #fff);", "color:#abc;",
		"border:1px solid #ddd;",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %s", want, text)
		}
	}
}

func TestRejectInvalidRuntimeValues(t *testing.T) {
	var length Length
	var color Color
	var display displayKeyword
	var pointer *Length
	for name, root := range map[string]*Element{
		"zero length":         Div().Width(length),
		"zero color":          Div().Color(color),
		"zero keyword":        Div().Display(display),
		"nil pointer":         Div().Width(pointer),
		"nil value":           Div().Width(nil),
		"nonfinite":           Div().Width(Px(math.Inf(1))),
		"nan":                 Div().Opacity(Number(math.NaN())),
		"bad hex":             Div().Color(Hex("#ggg")),
		"empty edges":         Div().Padding(Edges()),
		"bad edge":            Div().Padding(Edges(length)),
		"too many edges":      Div().Padding(Edges(Px(1), Px(2), Px(3), Px(4), Px(5))),
		"unknown property":    Div().CSSRaw("invented", "1px"),
		"custom via raw":      Div().CSSRaw("--accent", "red"),
		"bad variable name":   Div().CustomProperty("accent", "red"),
		"empty variable name": Div().CustomProperty("--", "red"),
		"variable injection":  Div().CustomProperty("--a;color", "red"),
		"bad reference":       Div().Color(Var("color")),
		"bad fallback":        Div().Color(Var("--accent", "red;color:blue")),
		"too many fallbacks":  Div().Color(Var("--accent", "red", "blue")),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Render(root); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestDeclarationBoundaries(t *testing.T) {
	for _, text := range []string{
		"", "red;color:blue", "red}", "rgb(1,2,3", "'open", "red/*",
		"red\x00", "red\\", "red\\\n", "([)]",
	} {
		if _, err := Render(Div().Color(RawValue(text))); err == nil {
			t.Errorf("accepted invalid value %q", text)
		}
		if _, err := Render(Div().CustomProperty("--value", text)); err == nil {
			t.Errorf("accepted invalid custom value %q", text)
		}
		if _, err := Render(Div().CSSRaw("color", text)); err == nil {
			t.Errorf("accepted invalid raw declaration %q", text)
		}
	}
	for _, text := range []string{
		`url("data:image/png;base64,abc")`, `"a;b{}"`, `calc(100% - 2px)`,
		`var(--a, rgb(1, 2, 3))`, `url(data:image/png;base64,abc)`,
	} {
		if _, err := Render(Div().BackgroundImage(RawValue(text))); err != nil {
			t.Errorf("rejected single value %q: %v", text, err)
		}
	}
}

// Check consumers as a separate package so private marker methods cannot be
// implemented by callers. go/types checks the real builder and generated API.
func TestPublicStyleAPITypeChecking(t *testing.T) {
	fset := token.NewFileSet()
	files := []*ast.File{}
	for _, path := range []string{"builder.go", "values.go", "properties_generated.go"} {
		file, err := parser.ParseFile(
			fset,
			path,
			nil,
			0,
		)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}
	config := types.Config{Importer: importer.Default()}
	pkg, err := config.Check("ownframe-ui", fset, files, nil)
	if err != nil {
		t.Fatal(err)
	}
	config.Importer = styleImporter{pkg: pkg, fallback: importer.Default()}
	for _, test := range []struct{ name, code, want string }{
		{name: "typed theme color", code: `ui.Div().Color(ui.ColorVar("--text", ui.Hex("#fff")))`},
		{name: "typed border", code: `ui.Div().Border(ui.SolidBorder(ui.Px(1), ui.Hex("#fff")))`},
		{name: "color cannot be length", code: `ui.Div().Width(ui.ColorVar("--text", ui.Hex("#fff")))`, want: "does not implement"},
		{name: "length cannot be color", code: `ui.Div().Color(ui.LengthVar("--gap", ui.Px(8)))`, want: "does not implement"},
		{name: "typed", code: `ui.Div().Width(ui.Px(10)).Display(ui.DisplayFlex).Color(ui.Hex("#fff"))`},
		{name: "raw", code: `ui.Div().Width(ui.RawValue("calc(100% - 2px)"))`},
		{name: "variable", code: `ui.Div().Color(ui.Var("--accent"))`},
		{name: "unknown property", code: `ui.Div().Invented(ui.Px(10))`, want: "Invented undefined"},
		{name: "string value", code: `ui.Div().Display("flex")`, want: "does not implement"},
		{name: "wrong keyword", code: `ui.Div().Display(ui.PositionAbsolute)`, want: "does not implement"},
		{name: "wrong family", code: `ui.Div().Color(ui.Px(10))`, want: "does not implement"},
		{name: "percent border", code: `ui.Div().BorderWidth(ui.Percent(10))`, want: "does not implement"},
		{name: "legacy CSS", code: `ui.Div().CSS("width", "10px")`, want: "CSS undefined"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "package consumer\nimport ui \"ownframe-ui\"\nfunc demo() { _ = " + test.code + " }"
			file, err := parser.ParseFile(
				fset,
				"consumer.go",
				source,
				0,
			)
			if err != nil {
				t.Fatal(err)
			}
			_, err = config.Check("consumer", fset, []*ast.File{file}, nil)
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v; want compiler error containing %q", err, test.want)
			}
		})
	}
}

type styleImporter struct {
	pkg      *types.Package
	fallback types.Importer
}

func (i styleImporter) Import(path string) (*types.Package, error) {
	if path == i.pkg.Path() {
		return i.pkg, nil
	}
	return i.fallback.Import(path)
}

func TestTypedThemeAndLayoutValues(t *testing.T) {
	root := Div().Height(Vh(100)).Flex(Number(1)).FontFamily(FontFamilySansSerif).
		Color(ColorVar("--text", Hex("#123456"))).
		Border(SolidBorder(Px(1), ColorVar("--border", Hex("#abcdef")))).
		Padding(LengthVar("--spacing", Px(12))).Transform(TranslateY(Px(-20)))
	source, err := Render(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"height:100vh;", "flex:1;", "font-family:sans-serif;", "color:var(--text, #123456);", "translateY(-20px)"} {
		if !strings.Contains(source, want) {
			t.Fatalf("missing %q in %s", want, source)
		}
	}
	for name, root := range map[string]*Element{
		"bad color fallback":  Div().Color(ColorVar("--color", Hex("no"))),
		"bad color name":      Div().Color(ColorVar("color", Hex("#fff"))),
		"bad length fallback": Div().Padding(LengthVar("--spacing", Length{})),
		"bad border":          Div().Border(SolidBorder(Length{}, Hex("#fff"))),
		"bad transform":       Div().Transform(TranslateY(Px(math.NaN()))),
	} {
		if _, err := Render(root); err == nil {
			t.Errorf("accepted %s", name)
		}
	}
}
