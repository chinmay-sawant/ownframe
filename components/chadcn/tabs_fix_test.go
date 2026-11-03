package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestTabsFixRenderBuildsOnlyActivePanel(t *testing.T) {
	t.Parallel()
	var calls []string
	items := []chadcn.TabItem{
		{ID: "overview", Label: "Overview", Content: ui.Text("Eager overview body")},
		{ID: "analytics", Label: "Analytics", Content: ui.Text("Eager analytics body")},
		{ID: "settings", Label: "Settings", Content: ui.Text("Eager settings body")},
	}
	source := feedbackHTML(t, chadcn.Tabs(chadcn.TabsOptions{
		Active: "analytics",
		Items:  items,
		Render: func(id string) *ui.Element {
			calls = append(calls, id)
			return ui.Text("Lazy " + id + " body")
		},
	}))
	if len(calls) != 1 || calls[0] != "analytics" {
		t.Fatalf("Render calls=%v, want [analytics]", calls)
	}
	if got := strings.Count(source, `role="tabpanel"`); got != 1 {
		t.Fatalf("tab panels=%d in %s", got, source)
	}
	if got := strings.Count(source, `role="tab"`); got != 3 {
		t.Fatalf("tabs=%d in %s", got, source)
	}
	for _, want := range []string{
		`aria-controls="analytics-panel"`, `id="analytics-panel"`,
		`aria-labelledby="analytics-tab"`, `class="of-tabs-panel"`,
		`Lazy analytics body`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("tabs missing %q in %s", want, source)
		}
	}
	for _, unwanted := range []string{
		"Lazy overview body", "Lazy settings body",
		"Eager overview body", "Eager analytics body", "Eager settings body",
	} {
		if strings.Contains(source, unwanted) {
			t.Fatalf("inactive or eager panel content %q rendered: %s", unwanted, source)
		}
	}
}

func TestTabsFixRenderNilResultKeepsPanelARIA(t *testing.T) {
	t.Parallel()
	var calls []string
	source := feedbackHTML(t, chadcn.Tabs(chadcn.TabsOptions{
		Items: []chadcn.TabItem{
			{ID: "overview", Label: "Overview"},
			{ID: "analytics", Label: "Analytics"},
		},
		Render: func(id string) *ui.Element {
			calls = append(calls, id)
			return nil
		},
	}))
	if len(calls) != 1 || calls[0] != "overview" {
		t.Fatalf("Render calls=%v, want [overview]", calls)
	}
	for _, want := range []string{
		`role="tabpanel"`, `id="overview-panel"`, `aria-labelledby="overview-tab"`,
		`class="of-tabs-panel of-tabs-panel--empty"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("nil panel missing %q in %s", want, source)
		}
	}
	if got := strings.Count(source, `aria-selected="true"`); got != 1 {
		t.Fatalf("selected tabs=%d in %s", got, source)
	}
}
