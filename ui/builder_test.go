package ui

import (
	"bufio"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestAllImplementedPropertiesMatchCompatibilityMatrix(t *testing.T) {
	f, err := os.Open("../documentation/compatibility-matrix.md")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			t.Errorf("close compatibility matrix: %v", cerr)
		}
	}()

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
	if len(seen) == 0 {
		t.Fatal("no implemented properties read from matrix")
	}
	names := SupportedProperties()
	if len(names) != len(seen) {
		t.Fatalf("builder: got %d implemented properties; want %d", len(names), len(seen))
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
		method := reflect.ValueOf(root).MethodByName(propertyMethodName(property))
		if !method.IsValid() {
			t.Fatalf("missing generated method for %s", property)
		}
		method.Call([]reflect.Value{reflect.ValueOf(Inherit)})
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
		BackgroundColor(Hex("#20242b")).
		Padding(Px(16)).
		BorderRadius(Px(8)).
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
	toolbar := Div().Display(DisplayFlex).FlexDirection(FlexDirectionRow).
		AlignItems(AlignItemsCenter).Gap(Px(8)).
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
		Div().CSSRaw("does-not-exist", "1px"),
		Div().Attr("style", "color:red"),
		Div().Attr("onclick", "javascript:alert(1)"),
		Div().Attr("bad name", "value"),
		Tag("script").Text("alert(1)"),
		Input().Children(Span()),
		Input().Text("not allowed"),
		Div().CSSRaw("color", ""),
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

// TestAttributeEntriesOverflowInlineCapacity exercises the spill from the
// inline attributes into the element's overflow slice, replacement of an
// existing name, and sorted output.
func TestAttributeEntriesOverflowInlineCapacity(t *testing.T) {
	e := Div().Attr("a", "1").Attr("b", "2").Attr("c", "3").Attr("d", "4").
		Attr("e", "5").Attr("a", "9")
	if got, want := e.attrs.len(), inlineAttrEntries+1; got != want {
		t.Fatalf("attrs len = %d, want %d", got, want)
	}
	markup, err := Render(e)
	if err != nil {
		t.Fatal(err)
	}
	const want = `<div a="9" b="2" c="3" d="4" e="5"></div>`
	if markup != want {
		t.Fatalf("Render = %q, want %q", markup, want)
	}
}

func propertyMethodName(property string) string {
	parts := strings.Split(strings.TrimPrefix(property, "-"), "-")
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}
