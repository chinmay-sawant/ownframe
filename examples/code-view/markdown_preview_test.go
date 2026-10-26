package main

import (
	"strings"
	"testing"
)

func TestPreviewRows(t *testing.T) {
	t.Parallel()

	lines := []string{
		"# Title",
		"",
		"Text with **bold** and `code`.",
		"",
		"- one",
		"- two",
		"",
		"> quote",
		"",
		"```go",
		"fmt.Println(\"x\")",
		"```",
		"",
		"| a | b |",
	}

	rows, anchors := previewRows(lines)
	kinds := map[string]bool{}
	var text []string

	for _, row := range rows {
		kinds[row.Kind] = true
		text = append(text, lineText(row.Segments))
	}

	for _, want := range []string{"h1", "p", "li", "quote", "code", "table", "gap", "rule"} {
		if want == "rule" {
			continue
		}

		if !kinds[want] {
			t.Errorf("missing row kind %q", want)
		}
	}

	all := strings.Join(text, "\n")
	for _, want := range []string{"Title", "bold", "code", "one", "quote", "fmt.Println"} {
		if !strings.Contains(all, want) {
			t.Errorf("preview missing %q: %q", want, all)
		}
	}

	for _, unwanted := range []string{"# Title", "**bold**", "`code`"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("preview kept markers %q", unwanted)
		}
	}

	if anchors[0] != 0 {
		t.Fatalf("heading anchor = %d", anchors[0])
	}
}
