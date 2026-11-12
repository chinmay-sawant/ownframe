package web

import (
	"html"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestBrowserDocumentInput(t *testing.T) {
	t.Parallel()

	const value = "ada@ex.com"
	const src = "<body>" +
		`<span id="email" class="field" placeholder="Email" data-bind="Email" data-ownframe-field="input"></span>` +
		"</body>"
	got := browserDocument(src, "", []page.WebControl{{
		ID:    "email",
		Tag:   "input",
		Type:  "text",
		Value: value,
	}}, "")

	tag := openTag(t, got, "input")
	for _, want := range []string{
		`type="` + html.EscapeString("text") + `"`,
		`value="` + html.EscapeString(value) + `"`,
		`class="field"`,
		`placeholder="Email"`,
		`data-bind="Email"`,
		`id="email"`,
	} {
		if !strings.Contains(tag, want) {
			t.Fatalf("input tag missing %q: %s\n%s", want, tag, got)
		}
	}
	if strings.Contains(got, "data-ownframe-field") || strings.Contains(got, "<span") {
		t.Fatalf("field span left behind:\n%s", got)
	}
}

func TestBrowserDocumentPassword(t *testing.T) {
	t.Parallel()

	// Painted password text is bullets. The control value is the secret.
	const value = "secret"
	bullets := strings.Repeat("•", len(value))
	src := `<body><span id="pw" data-ownframe-field="input">` + bullets + `</span></body>`
	got := browserDocument(src, "", []page.WebControl{{
		ID:    "pw",
		Tag:   "input",
		Type:  "password",
		Value: value,
	}}, "")

	tag := openTag(t, got, "input")
	for _, want := range []string{
		`<input`,
		`type="` + html.EscapeString("password") + `"`,
		`value="` + html.EscapeString(value) + `"`,
	} {
		if !strings.Contains(tag, want) {
			t.Fatalf("input tag missing %q: %s\n%s", want, tag, got)
		}
	}
	if strings.Contains(got, bullets) {
		t.Fatalf("bullets used as the value:\n%s", got)
	}
}

func TestBrowserDocumentTextarea(t *testing.T) {
	t.Parallel()

	const value = "a<b&c"
	const src = `<body><span id="note" data-ownframe-field="textarea">draft</span></body>`
	got := browserDocument(src, "", []page.WebControl{{
		ID:    "note",
		Tag:   "textarea",
		Value: value,
	}}, "")

	body := elementBody(t, got, "textarea")
	if body != html.EscapeString(value) {
		t.Fatalf("textarea body = %q, want %q\n%s", body, html.EscapeString(value), got)
	}
	if strings.Contains(got, "draft") || strings.Contains(got, "data-ownframe-field") || strings.Contains(got, value) {
		t.Fatalf("raw field left behind:\n%s", got)
	}
}

func TestBrowserDocumentSelect(t *testing.T) {
	t.Parallel()

	const src = `<body><select id="city" data-ownframe-field="select">Rome</select></body>`
	got := browserDocument(src, "", []page.WebControl{{
		ID:  "city",
		Tag: "select",
		Options: []page.SelectOption{
			{Value: "rm", Label: "Rome", Selected: true},
			{Value: "mi", Label: "Milan"},
		},
	}}, "")

	if strings.Contains(got, "data-ownframe-field") {
		t.Fatalf("field attr left behind:\n%s", got)
	}
	if strings.Count(got, "<select") != 1 || strings.Count(got, "<option") != 2 {
		t.Fatalf("select shape:\n%s", got)
	}
	if !strings.Contains(openTag(t, got, "select"), `id="city"`) {
		t.Fatalf("select id dropped:\n%s", got)
	}
	rome := optionByValue(t, got, "rm")
	milan := optionByValue(t, got, "mi")
	if !strings.Contains(rome, "selected") || !strings.Contains(rome, html.EscapeString("Rome")) {
		t.Fatalf("selected option: %s\n%s", rome, got)
	}
	if strings.Contains(milan, "selected") || !strings.Contains(milan, html.EscapeString("Milan")) {
		t.Fatalf("other option: %s\n%s", milan, got)
	}
	if strings.Count(got, "Rome") != 1 {
		t.Fatalf("painted label left behind:\n%s", got)
	}
}

func TestBrowserDocumentDropsNestedSelection(t *testing.T) {
	t.Parallel()

	const value = "ab"
	const src = `<body><span id="e" data-ownframe-field="input">` +
		`<span data-ownframe-selection="1">ab</span></span></body>`
	got := browserDocument(src, "", []page.WebControl{{
		ID:    "e",
		Tag:   "input",
		Type:  "text",
		Value: value,
	}}, "")

	if strings.Count(got, "<input") != 1 {
		t.Fatalf("inputs:\n%s", got)
	}
	if strings.Contains(got, "data-ownframe-selection") ||
		strings.Contains(got, "data-ownframe-field") ||
		strings.Contains(got, "<span") {
		t.Fatalf("selection markup left behind:\n%s", got)
	}
	tag := openTag(t, got, "input")
	if !strings.Contains(tag, `value="`+html.EscapeString(value)+`"`) {
		t.Fatalf("input tag: %s\n%s", tag, got)
	}
}

func TestBrowserDocumentBootOnce(t *testing.T) {
	t.Parallel()

	// A document that already carries the boot marker must not gain a second one.
	const boot = "var ownframeBoot = 1;"
	const marker = `<script id="ownframe-boot">`
	fresh := `<html><head></head><body><p id="m">hi</p></body></html>`
	got := browserDocument(fresh, "", nil, boot)
	if strings.Count(got, `id="ownframe-boot"`) != 1 || !strings.Contains(got, marker) {
		t.Fatalf("boot marker:\n%s", got)
	}
	if body := elementBody(t, got, "script"); body != boot {
		t.Fatalf("boot body = %q\n%s", body, got)
	}

	const already = `<html><head><script id="ownframe-boot">var old = 1;</script></head>` +
		`<body><p id="m">hi</p></body></html>`
	again := browserDocument(already, "", nil, boot)
	if strings.Count(again, `id="ownframe-boot"`) != 1 {
		t.Fatalf("duplicated boot:\n%s", again)
	}
}

func TestBrowserDocumentTheme(t *testing.T) {
	t.Parallel()

	const theme = ":root{color:red}"
	const src = `<html><head></head><body><p id="m">hi</p></body></html>`
	got := browserDocument(src, theme, nil, "")

	const open = `<style id="of-theme">`
	i := strings.Index(got, open)
	if i < 0 {
		t.Fatalf("missing %s\n%s", open, got)
	}
	rest := got[i+len(open):]
	j := strings.Index(rest, "</style>")
	if j < 0 {
		t.Fatalf("unclosed theme style\n%s", got)
	}
	if rest[:j] != theme {
		t.Fatalf("theme style = %q, want %q\n%s", rest[:j], theme, got)
	}
	if strings.Count(got, `id="of-theme"`) != 1 {
		t.Fatalf("theme count:\n%s", got)
	}
}

func TestBrowserDocumentPreservesPlainHTML(t *testing.T) {
	t.Parallel()

	const src = `<!doctype html><html><head><title>Keep</title></head><body><p id="m">hi</p></body></html>`
	got := browserDocument(src, "", nil, "")
	if !strings.Contains(got, `<p id="m">hi</p>`) || !strings.Contains(got, `<title>Keep</title>`) {
		t.Fatalf("plain html changed:\n%s", got)
	}
}

func TestBrowserDocumentOmitsFrame(t *testing.T) {
	t.Parallel()

	const src = `<html><head></head><body><p id="m">hi</p></body></html>`
	got := browserDocument(src, ":root{color:red}", nil, "var ownframeBoot = 1;")
	lower := strings.ToLower(got)
	if strings.Contains(lower, "<iframe") || strings.Contains(lower, "srcdoc") {
		t.Fatalf("frame markup in\n%s", got)
	}
}

func openTag(t *testing.T, doc, tag string) string {
	t.Helper()

	i := strings.Index(strings.ToLower(doc), "<"+tag)
	if i < 0 {
		t.Fatalf("no <%s in\n%s", tag, doc)
	}
	rest := doc[i:]
	j := strings.IndexByte(rest, '>')
	if j < 0 {
		t.Fatalf("unclosed <%s in\n%s", tag, doc)
	}

	return rest[:j]
}

func elementBody(t *testing.T, doc, tag string) string {
	t.Helper()

	open := openTag(t, doc, tag)
	i := strings.Index(doc, open)
	if i < 0 {
		t.Fatalf("missing %s in\n%s", open, doc)
	}
	start := i + len(open) + 1
	close := "</" + tag + ">"
	end := strings.Index(doc[start:], close)
	if end < 0 {
		t.Fatalf("no %s in\n%s", close, doc)
	}

	return doc[start : start+end]
}

func optionByValue(t *testing.T, doc, value string) string {
	t.Helper()

	needle := `value="` + html.EscapeString(value) + `"`
	i := strings.Index(doc, needle)
	if i < 0 {
		t.Fatalf("missing %s\n%s", needle, doc)
	}
	start := strings.LastIndex(doc[:i], "<option")
	if start < 0 {
		t.Fatalf("option value outside an option\n%s", doc)
	}
	end := strings.Index(doc[start:], "</option>")
	if end < 0 {
		t.Fatalf("unclosed option\n%s", doc)
	}

	return doc[start : start+end]
}
