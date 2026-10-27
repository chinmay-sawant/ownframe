package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// previewMaxHeadingPx is the largest heading face that still fits the 20 px
// preview row. The mono stack reports at most 1.32 em of ascent plus descent
// (JetBrains Mono: 1020 + 300 over 1000), so 14 * 1.32 = 18.5 px stays inside
// var(--row-h).
const previewMaxHeadingPx = 14

// previewRule is one parsed "selectors { declarations }" block.
type previewRule struct {
	selectors []string
	body      string
}

// readPreviewRules parses components/preview.css. The stylesheet has no
// at-rules, so one pass over braces is enough.
func readPreviewRules(t *testing.T) []previewRule {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("components", "preview.css"))
	if err != nil {
		t.Fatalf("read preview.css: %v", err)
	}

	comments := regexp.MustCompile(`(?s)/\*.*?\*/`)
	source := comments.ReplaceAllString(string(raw), "")

	pair := regexp.MustCompile(`(?s)([^{}]*)\{([^{}]*)\}`)
	matches := pair.FindAllStringSubmatch(source, -1)
	rules := make([]previewRule, 0, len(matches))

	for _, match := range matches {
		selectors := strings.Split(match[1], ",")
		for i := range selectors {
			selectors[i] = strings.Join(strings.Fields(selectors[i]), " ")
		}

		rules = append(rules, previewRule{selectors: selectors, body: match[2]})
	}

	return rules
}

// previewRulesFor returns the declarations of every rule whose selector list
// names the exact selector. The last declaration wins within one block.
func previewRulesFor(rules []previewRule, selector string) []map[string]string {
	var out []map[string]string

	for _, rule := range rules {
		named := false

		for _, candidate := range rule.selectors {
			if candidate == selector {
				named = true

				break
			}
		}

		if !named {
			continue
		}

		decls := map[string]string{}

		for _, decl := range strings.Split(rule.body, ";") {
			prop, value, ok := strings.Cut(decl, ":")
			if !ok {
				continue
			}

			decls[strings.TrimSpace(prop)] = strings.TrimSpace(value)
		}

		out = append(out, decls)
	}

	return out
}

// TestPreviewRowsKeepTheRowGrid holds the geometry the window relies on: the
// preview slices visible rows by scrollY / rowH and positions the track from
// scrollY % rowH, so every .prow and every blank .prow-gap must stay exactly
// one var(--row-h) tall. A shorter gap row would desync scrolling even though
// it would tighten block spacing.
func TestPreviewRowsKeepTheRowGrid(t *testing.T) {
	rules := readPreviewRules(t)

	for _, selector := range []string{".prow", ".prow-gap"} {
		blocks := previewRulesFor(rules, selector)
		if len(blocks) == 0 {
			t.Fatalf("preview.css has no %s rule", selector)
		}

		sized := false

		for _, block := range blocks {
			for _, prop := range []string{"height", "min-height", "max-height"} {
				value, ok := block[prop]
				if !ok {
					continue
				}

				if prop == "height" {
					sized = true
				}

				if value != "var(--row-h)" {
					t.Errorf("%s sets %s: %s, want var(--row-h)", selector, prop, value)
				}
			}
		}

		if !sized {
			t.Errorf("%s sets no height", selector)
		}
	}
}

// TestPreviewHeadingsFitTheRow checks that no heading asks for a face taller
// than the 20 px row can hold, that no heading overrides the row line height,
// and that H1 keeps title ink instead of the link blue.
func TestPreviewHeadingsFitTheRow(t *testing.T) {
	rules := readPreviewRules(t)

	for level := 1; level <= 6; level++ {
		selector := ".prow-h" + strconv.Itoa(level) + " .ptext"

		blocks := previewRulesFor(rules, selector)
		if len(blocks) == 0 {
			t.Fatalf("preview.css has no %s rule", selector)
		}

		for _, block := range blocks {
			if raw, ok := block["font-size"]; ok {
				if !strings.HasSuffix(raw, "px") {
					t.Fatalf("%s font-size is %q, want a px length", selector, raw)
				}

				size, err := strconv.ParseFloat(strings.TrimSuffix(raw, "px"), 64)
				if err != nil {
					t.Fatalf("%s font-size is %q: %v", selector, raw, err)
				}

				if size > previewMaxHeadingPx {
					t.Errorf("%s font-size is %.1f px, want at most %d px inside the 20 px row",
						selector, size, previewMaxHeadingPx)
				}
			}

			if raw, ok := block["line-height"]; ok && raw != "var(--row-h)" {
				t.Errorf("%s sets line-height: %s, want var(--row-h)", selector, raw)
			}
		}
	}

	for _, block := range previewRulesFor(rules, ".prow-h1 .ptext") {
		switch color := strings.ToLower(block["color"]); color {
		case "#82aaff", "var(--accent)":
			t.Errorf("H1 text uses the link ink %s, want title ink", color)
		}
	}
}
