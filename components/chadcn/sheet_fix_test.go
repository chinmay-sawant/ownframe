package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestSheetFixPanelParts keeps the right-side default and renders the header,
// scrolling body, and fixed footer in that order.
func TestSheetFixPanelParts(t *testing.T) {
	t.Parallel()
	root := chadcn.Sheet(chadcn.SheetOptions{
		ID:          "settings-sheet",
		Open:        true,
		Title:       "Project settings",
		Description: "Notifications, routing, and verification.",
		Footer: []*ui.Element{
			chadcn.Button("Cancel", chadcn.ButtonOptions{Variant: chadcn.ButtonText}),
			chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}),
		},
		OnOpenChange: func(context.Context, bool) error { return nil },
	}, ui.Text("Destination email"))
	markup := overlayMarkup(t, root)
	overlayWant(t, markup,
		`class="of-sheet-panel of-sheet--right"`,
		`class="of-sheet-header"`,
		`class="of-sheet-body"`,
		`class="of-sheet-footer"`,
		`class="of-sheet-footer-slot"`,
		`id="settings-sheet-panel"`,
		`id="settings-sheet-backdrop"`,
		`role="dialog"`,
		"Destination email",
		"Cancel",
		"Save",
	)
	if got := strings.Count(markup, `class="of-sheet-footer-slot"`); got != 2 {
		t.Fatalf("footer slots=%d, want 2:\n%s", got, markup)
	}
	header := strings.Index(markup, `class="of-sheet-header"`)
	body := strings.Index(markup, `class="of-sheet-body"`)
	footer := strings.Index(markup, `class="of-sheet-footer"`)
	if header < 0 || body <= header || footer <= body {
		t.Fatalf("sheet parts out of order header=%d body=%d footer=%d:\n%s", header, body, footer, markup)
	}
}

// TestSheetFixStylesheetWidthAndScroll pins the panel width, the narrow-window
// fallback, and the body and footer rules that keep a long form scrollable.
func TestSheetFixStylesheetWidthAndScroll(t *testing.T) {
	t.Parallel()
	panel := sheetFixBlock(t, chadcn.Stylesheet, ".of-sheet-panel {")
	if !strings.Contains(panel, "width: 400px;") {
		t.Fatalf("panel does not set the desktop width:\n%s", panel)
	}
	if strings.Contains(panel, "width: 90%;") {
		t.Fatalf("panel keeps the old percentage width:\n%s", panel)
	}
	if strings.Contains(panel, "max-width: 400px;") {
		t.Fatalf("panel still clamps through max-width:\n%s", panel)
	}

	narrow := sheetFixBlock(t, chadcn.Stylesheet, "@media (max-width: 520px)")
	for _, want := range []string{".of-sheet--right", ".of-sheet--left", "width: 90%;"} {
		if !strings.Contains(narrow, want) {
			t.Fatalf("narrow-window rule is missing %q:\n%s", want, narrow)
		}
	}
	if strings.Contains(narrow, ".of-sheet--top") || strings.Contains(narrow, ".of-sheet--bottom") {
		t.Fatalf("narrow-window rule shrinks edge sheets:\n%s", narrow)
	}

	body := sheetFixBlock(t, chadcn.Stylesheet, ".of-sheet-body {")
	for _, want := range []string{"min-height: 0;", "overflow-y: auto;"} {
		if !strings.Contains(body, want) {
			t.Fatalf("sheet body is missing %q:\n%s", want, body)
		}
	}

	footer := sheetFixBlock(t, chadcn.Stylesheet, ".of-sheet-footer {")
	if !strings.Contains(footer, "flex-shrink: 0;") {
		t.Fatalf("sheet footer can shrink:\n%s", footer)
	}
}

// sheetFixBlock returns one stylesheet block, braces included, from the first
// occurrence of marker to the brace that closes it.
func sheetFixBlock(t *testing.T, sheet, marker string) string {
	t.Helper()
	start := strings.Index(sheet, marker)
	if start < 0 {
		t.Fatalf("stylesheet is missing %q", marker)
	}
	open := strings.IndexByte(sheet[start:], '{')
	if open < 0 {
		t.Fatalf("stylesheet marker %q has no opening brace", marker)
	}
	depth := 0
	for index := start + open; index < len(sheet); index++ {
		switch sheet[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return sheet[start : index+1]
			}
		}
	}
	t.Fatalf("stylesheet marker %q is unterminated", marker)
	return ""
}
