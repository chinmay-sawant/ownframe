package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestLayoutFixToggleGroupWraps checks that a horizontal group carries the
// wrapping modifier while a vertical group keeps its single column.
func TestLayoutFixToggleGroupWraps(t *testing.T) {
	t.Parallel()
	horizontal := feedbackHTML(t, chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
		Label:    "Severity",
		Multiple: true,
		Selected: []string{"error"},
		Items: []chadcn.ToggleItem{
			{ID: "severity-error", Value: "error", Label: "Error"},
			{ID: "severity-warn", Value: "warn", Label: "Warn"},
		},
	}))
	for _, want := range []string{
		`class="of-toggle-group of-toggle-group--wrap"`,
		`role="group"`,
		`aria-label="Severity"`,
		`data-multiple="true"`,
		`aria-pressed="true"`,
		`aria-pressed="false"`,
		`data-value="error"`,
		`data-value="warn"`,
	} {
		if !strings.Contains(horizontal, want) {
			t.Fatalf("wrapping group missing %q in %s", want, horizontal)
		}
	}

	vertical := feedbackHTML(t, chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
		Vertical: true,
		Items:    []chadcn.ToggleItem{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}},
	}))
	if !strings.Contains(vertical, `class="of-toggle-group of-toggle-group--vertical"`) {
		t.Fatalf("vertical group lost its class in %s", vertical)
	}
	if strings.Contains(vertical, "of-toggle-group--wrap") {
		t.Fatalf("vertical group must not wrap: %s", vertical)
	}
	if strings.Contains(horizontal, "of-toggle-group--vertical") {
		t.Fatalf("horizontal group picked up the vertical class: %s", horizontal)
	}
}

// TestLayoutFixMessageParts checks the meta and content parts render in order
// so the stylesheet spacing has a stable target.
func TestLayoutFixMessageParts(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.Message(chadcn.MessageOptions{
		Name: "Ada Kim",
		Time: "14:28",
		Children: []*ui.Element{
			chadcn.Bubble(chadcn.BubbleOptions{Variant: chadcn.BubblePrimary}, ui.Text("Pulling the trace now.")),
		},
	}))
	for _, want := range []string{
		`class="of-message of-message--start"`,
		`class="of-message-avatar"`,
		`aria-hidden="true"`,
		">AK</div>",
		`class="of-message-body"`,
		`class="of-message-meta"`,
		`class="of-message-name"`,
		`class="of-message-time"`,
		`class="of-message-content"`,
		`class="of-bubble of-bubble--primary"`,
		"Ada Kim",
		"14:28",
		"Pulling the trace now.",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("message missing %q in %s", want, source)
		}
	}
	if strings.Index(source, "of-message-avatar") > strings.Index(source, "of-message-body") {
		t.Fatalf("avatar must paint before the body: %s", source)
	}
	if strings.Index(source, "of-message-name") > strings.Index(source, "of-message-time") {
		t.Fatalf("name must paint before the time: %s", source)
	}
	if strings.Index(source, "of-message-meta") > strings.Index(source, "of-message-content") {
		t.Fatalf("meta must paint above the content: %s", source)
	}
	if strings.Index(source, "of-message-content") > strings.Index(source, "of-bubble") {
		t.Fatalf("bubble must sit inside the content column: %s", source)
	}

	timeOnly := feedbackHTML(t, chadcn.Message(chadcn.MessageOptions{Time: "14:28"}))
	if strings.Contains(timeOnly, "of-message-name") {
		t.Fatalf("a message without a name should not paint one: %s", timeOnly)
	}
	if !strings.Contains(timeOnly, "of-message-meta") {
		t.Fatalf("a time-only message keeps its meta line: %s", timeOnly)
	}
}

// TestLayoutFixSpacingRules pins the spacing contract in the stylesheet.
func TestLayoutFixSpacingRules(t *testing.T) {
	t.Parallel()
	sheet := chadcn.Stylesheet
	for _, tt := range []struct {
		selector string
		wants    []string
	}{
		{
			selector: ".of-toggle-group",
			wants:    []string{"flex-wrap: wrap;", "gap: var(--of-toggle-group-gap, 8px);"},
		},
		{
			selector: ".of-toggle-group--wrap",
			wants:    []string{"flex-wrap: wrap;"},
		},
		{
			selector: ".of-toggle-group--vertical",
			wants:    []string{"flex-direction: column;", "align-items: stretch;", "flex-wrap: nowrap;"},
		},
		{
			selector: ".of-message",
			wants:    []string{"gap: 12px;"},
		},
		{
			selector: ".of-message-body",
			wants:    []string{"gap: 6px;"},
		},
		{
			selector: ".of-message-meta",
			wants:    []string{"gap: 8px;", "padding: 0 6px;", "line-height: 1.4;"},
		},
		{
			selector: ".of-bubble",
			wants:    []string{"padding: var(--of-bubble-padding, 11px 15px);", "line-height: 1.5;"},
		},
	} {
		rule := layoutFixRule(t, sheet, tt.selector)
		for _, want := range tt.wants {
			if !strings.Contains(rule, want) {
				t.Fatalf("rule %s missing %q in %q", tt.selector, want, rule)
			}
		}
	}
}

func layoutFixRule(t *testing.T, sheet, selector string) string {
	t.Helper()
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		t.Fatalf("stylesheet missing rule %s", selector)
	}
	rest := sheet[start:]
	end := strings.Index(rest, "}")
	if end < 0 {
		t.Fatalf("stylesheet rule %s has no closing brace", selector)
	}
	return rest[:end]
}
