package ui

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"math"
	"strings"
	"testing"
)

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
