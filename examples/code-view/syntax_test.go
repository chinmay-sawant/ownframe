package main

import (
	"strings"
	"testing"
)

func lineText(segs []segment) string {
	var b strings.Builder

	for _, seg := range segs {
		b.WriteString(seg.Text)
	}

	return b.String()
}

func classesOf(lines [][]segment) map[string]bool {
	out := map[string]bool{}

	for _, line := range lines {
		for _, seg := range line {
			for _, class := range strings.Fields(seg.Class) {
				out[class] = true
			}
		}
	}

	return out
}

func TestLanguageOf(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"backend/probe.go":     langGo,
		"tools/report.py":      langPython,
		"web/app.js":           langJS,
		"web/app.ts":           langJS,
		"README.md":            langMarkdown,
		"web/styles.css":       langCSS,
		"config/package.json":  langJSON,
		"web/index.html":       langHTML,
		"config/services.yaml": langYAML,
		"scripts/deploy.sh":    langShell,
		"Makefile":             langShell,
		"LICENSE":              langPlain,
	}

	for path, want := range cases {
		if got := languageOf(path); got != want {
			t.Errorf("%s: got %s, want %s", path, got, want)
		}
	}
}

func TestJavaScriptTokens(t *testing.T) {
	t.Parallel()

	src := []string{
		"// fetch the latest round",
		"export async function loadRound(tag = 'all') {",
		"  const response = await fetch(`/api/checks?tag=${tag}`);",
		"  if (response.ok === true) return response.json();",
		"  throw new Error('bad status');",
		"}",
	}
	got := tokenize(langJS, src)
	classes := classesOf(got)

	for _, want := range []string{classKw, classFn, classStr, classCom, classVar, classLit} {
		if !classes[want] {
			t.Errorf("missing token class %q", want)
		}
	}

	if lineText(got[2]) != src[2] {
		t.Errorf("line 3 text changed: %q", lineText(got[2]))
	}

	if !strings.Contains(lineText(got[1]), "loadRound") {
		t.Fatal("function name disappeared")
	}

	for i, line := range src {
		if line != "" && lineText(got[i]) != line {
			t.Errorf("line %d text changed", i+1)
		}
	}
}

func TestPythonTripleQuotedString(t *testing.T) {
	t.Parallel()

	src := []string{
		"class Report:",
		"    \"\"\"Summarise one window.",
		"",
		"    The docstring spans lines.",
		"    \"\"\"",
		"    def rows(self, since):",
		"        return self._query(since)  # comment",
	}
	got := tokenize(langPython, src)

	if !classesOf(got)[classStr] {
		t.Fatal("no string token")
	}

	if !classesOf(got)[classCom] {
		t.Fatal("no comment token")
	}

	for i, line := range src {
		if line != "" && lineText(got[i]) != line {
			t.Errorf("line %d text changed", i+1)
		}
	}
}

func TestGoFunctionReceivers(t *testing.T) {
	t.Parallel()

	got := tokenize(langGo, []string{
		"func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {",
		"type Result struct {",
		"func NewProber(timeout time.Duration) *Prober {",
	})
	classes := classesOf(got)

	if !classes[classFn] || !classes[classType] {
		t.Fatalf("classes = %v", classes)
	}

	// The method name must still be a function token, not part of the receiver.
	found := false

	for _, seg := range got[0] {
		if strings.TrimSpace(seg.Text) == "ServeHTTP" && seg.Class == classFn {
			found = true
		}
	}

	if !found {
		t.Fatalf("ServeHTTP not highlighted as a function: %+v", got[0])
	}
}

func TestMarkdownStructure(t *testing.T) {
	t.Parallel()

	src := []string{
		"# Title",
		"",
		"## Section",
		"",
		"```go",
		"fmt.Println(\"inside a fence\")",
		"```",
		"",
		"Use **bold** and [`link`](https://example.com).",
		"- one item",
	}
	got := tokenize(langMarkdown, src)
	classes := classesOf(got)

	for _, want := range []string{classH1, classH2, classCode, classBold, classLink, classPunc} {
		if !classes[want] {
			t.Errorf("missing %q: %v", want, classes)
		}
	}

	if lineText(got[5]) != src[5] {
		t.Errorf("fence line changed: %q", lineText(got[5]))
	}

	for i, line := range src {
		if line != "" && lineText(got[i]) != line {
			t.Errorf("line %d text changed: %q", i+1, lineText(got[i]))
		}
	}
}

func TestMarkupTokens(t *testing.T) {
	t.Parallel()

	cases := map[string][]string{
		langCSS:   {"@media (max-width: 900px) {", "  .card { color: #ff00aa; }"},
		langJSON:  {"{", "  \"name\": \"beacon\",", "  \"port\": 8099", "}"},
		langHTML:  {"<button class=\"chip\" data-tag=\"api\">API</button>"},
		langYAML:  {"services:", "  - name: checkout", "    tags: [api]"},
		langShell: {"#!/bin/sh", "set -eu", "for file in *.go; do echo \"$file\"; done"},
	}

	for lang, lines := range cases {
		got := tokenize(lang, lines)

		for i, line := range lines {
			if lineText(got[i]) != line {
				t.Errorf("%s line %d text changed", lang, i+1)
			}
		}
	}
}

func TestHighlightLineAlwaysHasText(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langGo, langPython, langJS, langMarkdown, langCSS, langJSON, langHTML, langYAML, langShell, langPlain} {
		got := highlightLine(lang, "")
		if len(got) == 0 || lineText(got) == "" {
			t.Errorf("%s: empty line produced no text", lang)
		}
	}
}

func TestSpacesAtTokenBoundaries(t *testing.T) {
	t.Parallel()

	got := highlightLine(langGo, "Service string")
	if lineText(got) != "Service string" {
		t.Fatalf("text = %q", lineText(got))
	}

	for i := 0; i+1 < len(got); i++ {
		if strings.HasSuffix(got[i].Text, " ") {
			t.Fatalf("segment %d ends with a space: %+v", i, got)
		}
	}

	line := diffLine{Kind: '+', Text: "const total = 20", Emph: true, EmphAt: 14, EmphTo: 16}
	marked := diffSegments(line, langGo)

	for i := 0; i+1 < len(marked); i++ {
		if strings.HasSuffix(marked[i].Text, " ") {
			t.Fatalf("marked segment %d ends with a space: %+v", i, marked)
		}
	}
}

func TestSyntaxCache_IncrementalReuse(t *testing.T) {
	lines := []string{
		"package main",
		"",
		"func add(a, b int) int {",
		"\treturn a + b",
		"}",
		"",
		"func sub(a, b int) int {",
		"\treturn a - b",
		"}",
	}

	cache := newSyntaxCache(langGo)
	initial := cache.tokenize(lines)
	if len(initial) != len(lines) {
		t.Fatalf("expected %d lines, got %d", len(lines), len(initial))
	}

	// Capture segment slice pointers for lines 0..2
	p0 := &initial[0][0]
	p2 := &initial[2][0]

	// Edit only line 7 ("\treturn a - b" -> "\treturn a - b - 1")
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)
	linesCopy[7] = "\treturn a - b - 1"

	cache.markDirty(7)
	updated := cache.tokenize(linesCopy)

	// Verify line 7 changed
	if lineText(updated[7]) != "\treturn a - b - 1" {
		t.Fatalf("line 7 not updated: %q", lineText(updated[7]))
	}

	// Verify unmodified lines 0 and 2 retained identical underlying token segment addresses
	if &updated[0][0] != p0 {
		t.Errorf("line 0 was re-allocated instead of reused")
	}
	if &updated[2][0] != p2 {
		t.Errorf("line 2 was re-allocated instead of reused")
	}
}

func TestSyntaxCache_MultilinePropagation(t *testing.T) {
	lines := []string{
		"package main",
		"/*",
		"start of block comment",
		"end of block comment",
		"*/",
		"func hello() {}",
	}

	cache := newSyntaxCache(langGo)
	initial := cache.tokenize(lines)

	if !classesOf([][]segment{initial[2]})[classCom] {
		t.Fatalf("line 2 should be comment")
	}
	if !classesOf([][]segment{initial[5]})[classKw] {
		t.Fatalf("line 5 should be func keyword")
	}

	// Remove "*/" on line 4, so line 5 becomes part of comment
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)
	linesCopy[4] = "// no longer closing"

	cache.markDirty(4)
	updated := cache.tokenize(linesCopy)

	// Now line 5 should have comment class, not keyword class
	if classesOf([][]segment{updated[5]})[classKw] {
		t.Errorf("line 5 still has keyword class after block comment extended")
	}
	if !classesOf([][]segment{updated[5]})[classCom] {
		t.Errorf("line 5 missing comment class after block comment extended")
	}
}

func TestWrapPlanArena_CapacityRetained(t *testing.T) {
	lines := []string{
		"package main",
		"func longFunctionSignature(ctx context.Context, firstArgument int, secondArgument string) error {",
		"\treturn nil",
		"}",
	}

	doc := &fileDoc{
		Path:   "test.go",
		Lang:   langGo,
		Lines:  lines,
		Tokens: tokenize(langGo, lines),
	}

	plan1 := doc.wrapped(40)
	if plan1 == nil || len(plan1.Rows) == 0 {
		t.Fatalf("plan1 empty")
	}
	if len(plan1.arena.offsets) != len(plan1.Rows) || len(plan1.arena.lengths) != len(plan1.Rows) {
		t.Errorf("arena offsets/lengths mismatch: %d vs %d", len(plan1.arena.offsets), len(plan1.Rows))
	}

	rowsCap1 := cap(plan1.Rows)
	firstCap1 := cap(plan1.First)

	// Resize to cols 60
	plan2 := doc.wrapped(60)
	if plan2 != plan1 {
		t.Errorf("wrapPlan object was re-allocated instead of reused on resize")
	}
	if cap(plan2.Rows) < rowsCap1 {
		t.Errorf("rows capacity decreased")
	}
	if cap(plan2.First) < firstCap1 {
		t.Errorf("first capacity decreased")
	}
}
