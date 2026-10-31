package ui

import (
	"bufio"
	"os"
	"slices"
	"strings"
	"testing"
)

const expectedImplementedPropertyCount = 555

func TestAllImplementedPropertiesMatchCompatibilityMatrix(t *testing.T) {
	f, err := os.Open("../documentation/compatibility-matrix.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	seen := make(map[string]bool)
	inImplemented := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "### 2.1 Implemented") {
			inImplemented = true
			continue
		}
		if inImplemented && strings.HasPrefix(line, "### ") {
			break
		}
		if !inImplemented || !strings.HasPrefix(line, "| \x60") {
			continue
		}
		rest := strings.TrimPrefix(line, "| \x60")
		end := strings.Index(rest, "\x60 |")
		if end < 0 {
			continue
		}
		name := rest[:end]
		if seen[name] {
			t.Fatalf("duplicate compatibility matrix property %q", name)
		}
		seen[name] = true
		if !SupportsProperty(name) {
			t.Errorf("implemented CSS property %q is missing from builder", name)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(seen) != expectedImplementedPropertyCount {
		t.Fatalf("compatibility matrix: got %d implemented properties; want 555", len(seen))
	}
	names := SupportedProperties()
	if len(names) != expectedImplementedPropertyCount {
		t.Fatalf("builder: got %d implemented properties; want 555", len(names))
	}
	if !slices.IsSorted(names) {
		t.Fatal("SupportedProperties must be sorted")
	}
	for _, name := range names {
		if !seen[name] {
			t.Errorf("builder includes %q, which is not in the implemented matrix", name)
		}
	}
}

func TestBuilderCanExpressEveryImplementedProperty(t *testing.T) {
	root := Div()
	for _, property := range SupportedProperties() {
		root.CSS(property, "inherit")
	}
	html, err := Render(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, property := range SupportedProperties() {
		if !strings.Contains(html, property+":inherit;") {
			t.Errorf("missing serialized property %q", property)
		}
	}
}

func TestStyledCardAndEscapedMarkup(t *testing.T) {
	root := Div().
		BackgroundColor("#20242b").
		Padding("16px").
		BorderRadius("8px").
		Attr("title", "A & \"B\" {{.Private}}").
		Children(Text("<OwnFrame> {{.Secret}} & Go"))
	markup, err := Render(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"background-color:#20242b;",
		"border-radius:8px;",
		"padding:16px;",
		"A &amp; &#34;B&#34; &#123;&#123;.Private}}",
		"&lt;OwnFrame&gt; &#123;&#123;.Secret}} &amp; Go",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup did not contain %q: %s", want, markup)
		}
	}
	if strings.Contains(markup, "{{") {
		t.Fatalf("raw Go template action reached HTML source: %s", markup)
	}
}

func TestFlexToolbarAndList(t *testing.T) {
	toolbar := Div().Display("flex").FlexDirection("row").
		AlignItems("center").Gap("8px").
		Children(Button("Save"), Button("Cancel"))
	markup, err := Render(toolbar)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"display:flex;",
		"flex-direction:row;",
		"align-items:center;",
		"gap:8px;",
		"<button>Save</button><button>Cancel</button>",
	} {
		if !strings.Contains(markup, want) {
			t.Fatalf("missing %q in %s", want, markup)
		}
	}
	files := Ul().Children(Li().Text("main.go"), Li().Text("go.mod"))
	fragment, err := Render(files)
	if err != nil {
		t.Fatal(err)
	}
	if fragment != "<ul><li>main.go</li><li>go.mod</li></ul>" {
		t.Fatalf("unexpected list: %s", fragment)
	}
}

func TestRejectUnsupportedStylesAndUnsafeAttrs(t *testing.T) {
	for _, root := range []*Element{
		Div().CSS("does-not-exist", "1px"),
		Div().Attr("style", "color:red"),
		Div().Attr("onclick", "javascript:alert(1)"),
		Div().Attr("bad name", "value"),
		Tag("script").Text("alert(1)"),
		Input().Children(Span()),
		Input().Text("not allowed"),
		Div().CSS("color", ""),
	} {
		if _, err := Render(root); err == nil {
			t.Errorf("expected validation failure for %#v", root)
		}
	}
}

func TestRejectCycles(t *testing.T) {
	root := Div()
	root.Children(root)
	if _, err := Render(root); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestRepeatedChildAllowed(t *testing.T) {
	child := Text("repeat")
	got, err := Render(Div().Children(child, child))
	if err != nil {
		t.Fatal(err)
	}
	if got != "<div><span>repeat</span><span>repeat</span></div>" {
		t.Fatalf("unexpected markup: %s", got)
	}
}
