package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

// TestDataHoverRulesTargetIdBearingParts pins the hover rules that only match
// once the row, bar, or copy control carries an id.
func TestDataHoverRulesTargetIdBearingParts(t *testing.T) {
	t.Parallel()
	for _, selector := range []string{
		".of-data-table-body .of-data-table-row:hover",
		".of-chart-bar:hover",
		".of-code-block-copy:hover",
	} {
		if !strings.Contains(chadcn.Stylesheet, selector+" {") {
			t.Fatalf("stylesheet dropped %s", selector)
		}
	}
}

// TestCodeBlockCopyStaysDecorativeWithoutCallback keeps the copy chip as a
// hidden label when the host passes neither an action nor OnCopy. The code
// body stays text.
func TestCodeBlockCopyStaysDecorativeWithoutCallback(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.CodeBlock(chadcn.CodeBlockOptions{
		Filename: "main.go",
		Language: "go",
		Code:     "fmt.Println(1)",
	}))
	compositeWant(t, markup,
		`<span aria-hidden="true" class="of-code-block-copy">Copy</span>`,
		`<pre class="of-code-block-body"><code class="of-code-block-code">fmt.Println(1)</code></pre>`,
	)
	compositeDeny(t, markup, "<button", `id=`, "<input", "<textarea", "contenteditable")
}

// TestCodeBlockCopyButtonWiresOnCopy turns the chip into a button with a
// stable id and wires OnCopy through OnClick. The code body stays text.
func TestCodeBlockCopyButtonWiresOnCopy(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.CodeBlock(chadcn.CodeBlockOptions{
		ID:     "snippet",
		Code:   "a < b",
		OnCopy: func(context.Context) error { return nil },
	}))
	compositeWant(t, markup,
		`id="snippet"`,
		`id="snippet-copy"`,
		`type="button"`,
		`class="of-code-block-copy"`,
		`data-action=`,
		`>Copy</button>`,
		`<code class="of-code-block-code">a &lt; b</code>`,
	)
	button := compositeTag(t, markup, `id="snippet-copy"`)
	if !strings.Contains(button, "<button") || strings.Contains(button, "aria-hidden") {
		t.Fatalf("copy control is not an exposed button: %s", button)
	}
	code := compositeTag(t, markup, `class="of-code-block-code"`)
	if strings.Contains(code, "id=") {
		t.Fatalf("code body took an id: %s", code)
	}
	compositeDeny(t, markup, "<input", "<textarea", "contenteditable", "<span aria-hidden=\"true\" class=\"of-code-block-copy\">")
}

// TestCodeBlockCopyButtonUsesAction keeps a host data-action when OnCopy is
// absent, and still gives the button an id. OnCopy wins when both are set.
func TestCodeBlockCopyButtonUsesAction(t *testing.T) {
	t.Parallel()
	markup := renderCompositeTree(t, chadcn.CodeBlock(chadcn.CodeBlockOptions{
		ID:     "snippet",
		Action: "copy:snippet",
		Code:   "package main",
	}))
	compositeWant(t, markup, `id="snippet-copy"`, `data-action="copy:snippet"`, `type="button"`, `>Copy</button>`)
	compositeDeny(t, markup, "ownframe-ui-", "aria-hidden=\"true\" class=\"of-code-block-copy\"")

	unnamed := renderCompositeTree(t, chadcn.CodeBlock(chadcn.CodeBlockOptions{
		Action: "copy:snippet",
		Code:   "package main",
	}))
	button := compositeTag(t, unnamed, `data-action="copy:snippet"`)
	if !strings.Contains(button, "<button") || !strings.Contains(button, "id=") {
		t.Fatalf("action copy control has no id: %s", button)
	}

	both := renderCompositeTree(t, chadcn.CodeBlock(chadcn.CodeBlockOptions{
		ID:     "snippet",
		Action: "copy:snippet",
		Code:   "package main",
		OnCopy: func(context.Context) error { return nil },
	}))
	compositeWant(t, both, `id="snippet-copy"`, `data-action="ownframe-ui-`)
	compositeDeny(t, both, `data-action="copy:snippet"`)
}

// TestHexNamedCellExposesLabel keeps a labeled cell in the accessibility
// tree under its label. A cell with an empty label stays hidden. The three
// solid parts stay hidden either way.
func TestHexNamedCellExposesLabel(t *testing.T) {
	t.Parallel()
	named := renderCompositeTree(t, chadcn.HexCellElement(chadcn.HexCell{Label: "Plumbing", Tone: 1}))
	cell := compositeTag(t, named, `class="of-hex"`)
	for _, want := range []string{`aria-label="Plumbing"`, `role="img"`, `title="Plumbing"`, `data-tone="1"`} {
		if !strings.Contains(cell, want) {
			t.Fatalf("named cell missing %q in %s", want, cell)
		}
	}
	if strings.Contains(cell, "aria-hidden") {
		t.Fatalf("named cell is hidden: %s", cell)
	}
	if got := strings.Count(named, `aria-hidden="true"`); got != 3 {
		t.Fatalf("named cell hid %d parts, want 3 in %s", got, named)
	}

	empty := renderCompositeTree(t, chadcn.HexCellElement(chadcn.HexCell{Tone: 0}))
	hidden := compositeTag(t, empty, `class="of-hex"`)
	if !strings.Contains(hidden, `aria-hidden="true"`) || strings.Contains(hidden, "aria-label") {
		t.Fatalf("empty cell is exposed: %s", hidden)
	}

	grid := renderCompositeTree(t, chadcn.HexGrid(chadcn.HexGridOptions{
		Label: "Portfolio health",
		Cells: []chadcn.HexCell{
			{Label: "Plumbing", Tone: 1},
			{Tone: 2},
			{Label: "HVAC", Tone: 3},
		},
	}))
	compositeWant(t, grid, `aria-label="Portfolio health"`, `aria-label="Plumbing"`, `aria-label="HVAC"`)
	if got := strings.Count(grid, `class="of-hex"`); got != 3 {
		t.Fatalf("cells=%d, want 3 in %s", got, grid)
	}
}
