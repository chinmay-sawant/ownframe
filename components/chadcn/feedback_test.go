package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func feedbackHTML(t *testing.T, element *ui.Element) string {
	t.Helper()
	source, err := ui.Render(element)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestAlertVariantClassAndSlots(t *testing.T) {
	t.Parallel()
	variants := map[chadcn.AlertVariant]string{
		chadcn.AlertInfo:    "of-alert--info",
		chadcn.AlertSuccess: "of-alert--success",
		chadcn.AlertWarning: "of-alert--warning",
		chadcn.AlertDanger:  "of-alert--danger",
	}
	for variant, want := range variants {
		source := feedbackHTML(t, chadcn.Alert(chadcn.AlertOptions{Variant: variant, Title: "Status"}))
		if !strings.Contains(source, want) {
			t.Fatalf("variant missing class %q in %s", want, source)
		}
	}
	source := feedbackHTML(t, chadcn.Alert(chadcn.AlertOptions{
		Variant:     chadcn.AlertDanger,
		Title:       "Payment failed",
		Description: "Your card was declined.",
		Icon:        ui.Text("!"),
		Action:      chadcn.Button("Retry", chadcn.ButtonOptions{}),
	}))
	for _, want := range []string{
		`class="of-alert of-alert--danger"`, `role="alert"`,
		`class="of-alert-icon"`, `aria-hidden="true"`,
		`class="of-alert-title"`, `Payment failed`,
		`class="of-alert-description"`, `Your card was declined.`,
		`class="of-alert-action"`, `>Retry<`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("alert missing %q in %s", want, source)
		}
	}
	plain := feedbackHTML(t, chadcn.Alert(chadcn.AlertOptions{Title: "Heads up"}))
	if strings.Contains(plain, "of-alert-description") || strings.Contains(plain, "of-alert-action") {
		t.Fatalf("empty slots rendered markup: %s", plain)
	}
}

func TestToastClosedAndOpenRendering(t *testing.T) {
	t.Parallel()
	if toast := chadcn.Toast(chadcn.ToastOptions{Open: false, Title: "Event created"}); toast != nil {
		t.Fatal("closed toast rendered an element")
	}
	source := feedbackHTML(t, chadcn.Toast(chadcn.ToastOptions{
		Open:        true,
		Title:       "Event created",
		Description: "Sunday, December 3 at 9:00 AM",
		Action:      chadcn.Button("Undo", chadcn.ButtonOptions{}),
		CloseAction: "toast:close",
	}))
	for _, want := range []string{
		`class="of-toast"`, `role="status"`, `aria-live="polite"`,
		`class="of-toast-body"`, `class="of-toast-title"`, `Event created`,
		`class="of-toast-description"`, `Sunday, December 3 at 9:00 AM`,
		`class="of-toast-action"`, `>Undo<`,
		`class="of-toast-close"`, `aria-label="Close"`, `data-action="toast:close"`,
		`position:fixed`, `width:356px`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("toast missing %q in %s", want, source)
		}
	}
	noAction := feedbackHTML(t, chadcn.Toast(chadcn.ToastOptions{Open: true, Title: "Saved"}))
	if strings.Contains(noAction, "data-action") {
		t.Fatalf("empty close action rendered an action: %s", noAction)
	}
}

func TestAccordionRendersOnlyOpenPanels(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.Accordion(chadcn.AccordionOptions{
		Open:     []string{"shipping"},
		OnToggle: func(context.Context, string) error { return nil },
		Items: []chadcn.AccordionItem{
			{ID: "shipping", Title: "What are your shipping options?", Content: ui.Text("Free shipping on international orders.")},
			{ID: "returns", Title: "What is your return policy?", Content: ui.Text("Return within 30 days.")},
		},
	}))
	if got := strings.Count(source, `aria-expanded="true"`); got != 1 {
		t.Fatalf("expanded triggers=%d in %s", got, source)
	}
	if got := strings.Count(source, `aria-expanded="false"`); got != 1 {
		t.Fatalf("collapsed triggers=%d in %s", got, source)
	}
	if got := strings.Count(source, `class="of-accordion-item"`); got != 2 {
		t.Fatalf("items=%d in %s", got, source)
	}
	for _, want := range []string{
		`class="of-accordion"`, `class="of-accordion-trigger"`, `class="of-accordion-title"`,
		`id="shipping-trigger"`, `aria-controls="shipping-panel"`,
		`class="of-accordion-content"`, `id="shipping-panel"`,
		`aria-labelledby="shipping-trigger"`, `role="region"`,
		`Free shipping on international orders.`, `What is your return policy?`,
		`data-action=`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("accordion missing %q in %s", want, source)
		}
	}
	if strings.Contains(source, "Return within 30 days.") {
		t.Fatalf("closed panel rendered its content: %s", source)
	}
}

func TestCollapsibleRendersTriggerAndOpenContent(t *testing.T) {
	t.Parallel()
	render := func(open bool, children ...*ui.Element) string {
		t.Helper()
		return feedbackHTML(t, chadcn.Collapsible(chadcn.CollapsibleOptions{
			ID:           "order",
			Open:         open,
			Trigger:      chadcn.Button("Order #4189", chadcn.ButtonOptions{}).ID("order-toggle"),
			OnOpenChange: func(context.Context, bool) error { return nil },
		}, children...))
	}
	closed := render(false, ui.Text("Status: Shipped"))
	for _, want := range []string{
		`class="of-collapsible"`, `id="order"`, `id="order-toggle"`,
		`aria-expanded="false"`, `Order #4189`, `data-action=`,
	} {
		if !strings.Contains(closed, want) {
			t.Fatalf("closed collapsible missing %q in %s", want, closed)
		}
	}
	if strings.Contains(closed, `aria-controls="order-content"`) {
		t.Fatalf("closed collapsible points at an absent panel: %s", closed)
	}
	if strings.Contains(closed, "of-collapsible-content") || strings.Contains(closed, "Status: Shipped") {
		t.Fatalf("closed collapsible rendered content: %s", closed)
	}
	open := render(true, ui.Text("Status: Shipped"))
	for _, want := range []string{
		`aria-expanded="true"`, `aria-controls="order-content"`,
		`class="of-collapsible-content"`, `id="order-content"`, `Status: Shipped`,
	} {
		if !strings.Contains(open, want) {
			t.Fatalf("open collapsible missing %q in %s", want, open)
		}
	}
	if empty := chadcn.Collapsible(chadcn.CollapsibleOptions{}); empty != nil {
		t.Fatal("collapsible without a trigger rendered an element")
	}
}

func TestTabsRendersActiveTabAndPanel(t *testing.T) {
	t.Parallel()
	items := []chadcn.TabItem{
		{ID: "overview", Label: "Overview", Content: ui.Text("Overview body")},
		{ID: "analytics", Label: "Analytics", Content: ui.Text("Analytics body")},
		{ID: "settings", Label: "Settings", Content: ui.Text("Settings body")},
	}
	source := feedbackHTML(t, chadcn.Tabs(chadcn.TabsOptions{
		Active:   "analytics",
		OnSelect: func(context.Context, string) error { return nil },
		Items:    items,
	}))
	if got := strings.Count(source, `role="tablist"`); got != 1 {
		t.Fatalf("tablists=%d in %s", got, source)
	}
	if got := strings.Count(source, `role="tab"`); got != 3 {
		t.Fatalf("tabs=%d in %s", got, source)
	}
	if got := strings.Count(source, `role="tabpanel"`); got != 1 {
		t.Fatalf("tab panels=%d in %s", got, source)
	}
	if got := strings.Count(source, `aria-selected="true"`); got != 1 {
		t.Fatalf("selected tabs=%d in %s", got, source)
	}
	if got := strings.Count(source, `aria-selected="false"`); got != 2 {
		t.Fatalf("unselected tabs=%d in %s", got, source)
	}
	for _, want := range []string{
		`aria-controls="analytics-panel"`, `id="analytics-tab"`,
		`aria-labelledby="analytics-tab"`, `id="analytics-panel"`,
		`role="tab" tabindex="0"`, `role="tab" tabindex="-1"`,
		`class="of-tabs-panel"`, `Analytics body`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("tabs missing %q in %s", want, source)
		}
	}
	if strings.Contains(source, "Overview body") || strings.Contains(source, "Settings body") {
		t.Fatalf("inactive panels rendered content: %s", source)
	}
	if got := strings.Count(source, "data-action="); got != 3 {
		t.Fatalf("tab actions=%d in %s", got, source)
	}
	defaulted := feedbackHTML(t, chadcn.Tabs(chadcn.TabsOptions{Items: items}))
	if !strings.Contains(defaulted, "Overview body") || strings.Contains(defaulted, "Analytics body") {
		t.Fatalf("empty Active did not select the first tab: %s", defaulted)
	}
}

func TestBreadcrumbMarksCurrentPage(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.Breadcrumb([]chadcn.BreadcrumbItem{
		{Label: "Home", Action: "nav:home"},
		{Label: "Components", Action: "nav:components"},
		{Label: "Breadcrumb"},
	}))
	for _, want := range []string{
		`aria-label="Breadcrumb"`, `class="of-breadcrumb-list"`,
		`class="of-breadcrumb-link"`, `data-action="nav:home"`, `data-action="nav:components"`,
		`class="of-breadcrumb-page"`, `aria-current="page"`, `>Breadcrumb</span>`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("breadcrumb missing %q in %s", want, source)
		}
	}
	if got := strings.Count(source, `class="of-breadcrumb-separator"`); got != 2 {
		t.Fatalf("separators=%d in %s", got, source)
	}
	if got := strings.Count(source, `aria-current="page"`); got != 1 {
		t.Fatalf("current pages=%d in %s", got, source)
	}
	flagged := feedbackHTML(t, chadcn.Breadcrumb([]chadcn.BreadcrumbItem{
		{Label: "Home", Action: "nav:home"},
		{Label: "Settings", Action: "nav:settings", Current: true},
	}))
	if strings.Contains(flagged, `data-action="nav:settings"`) {
		t.Fatalf("current page kept its action: %s", flagged)
	}
	if !strings.Contains(flagged, `aria-current="page"`) {
		t.Fatalf("explicit current page lost its state: %s", flagged)
	}
}

func TestPaginationBoundsAndPageRange(t *testing.T) {
	t.Parallel()
	first := feedbackHTML(t, chadcn.Pagination(chadcn.PaginationOptions{
		Page:       1,
		TotalPages: 3,
		OnSelect:   func(context.Context, int) error { return nil },
	}))
	if !strings.Contains(first, `aria-label="Go to previous page" class="of-pagination-previous" disabled="disabled"`) {
		t.Fatalf("first page did not disable previous: %s", first)
	}
	if strings.Contains(first, `aria-label="Go to next page" class="of-pagination-next" disabled="disabled"`) {
		t.Fatalf("first page disabled next: %s", first)
	}
	if !strings.Contains(first, `<button aria-current="page" aria-label="Page 1" class="of-pagination-link"`) {
		t.Fatalf("first page is not current: %s", first)
	}
	last := feedbackHTML(t, chadcn.Pagination(chadcn.PaginationOptions{Page: 5, TotalPages: 5}))
	if !strings.Contains(last, `aria-label="Go to next page" class="of-pagination-next" disabled="disabled"`) {
		t.Fatalf("last page did not disable next: %s", last)
	}
	if strings.Contains(last, `aria-label="Go to previous page" class="of-pagination-previous" disabled="disabled"`) {
		t.Fatalf("last page disabled previous: %s", last)
	}
	if got := strings.Count(first, `disabled="disabled"`); got != 1 {
		t.Fatalf("disabled bounds=%d in %s", got, first)
	}
	clamped := feedbackHTML(t, chadcn.Pagination(chadcn.PaginationOptions{Page: 99, TotalPages: 5}))
	if !strings.Contains(clamped, `<button aria-current="page" aria-label="Page 5" class="of-pagination-link"`) {
		t.Fatalf("out-of-range page was not clamped: %s", clamped)
	}
	window := feedbackHTML(t, chadcn.Pagination(chadcn.PaginationOptions{Page: 5, TotalPages: 10}))
	if got := strings.Count(window, `class="of-pagination-ellipsis"`); got != 2 {
		t.Fatalf("ellipses=%d in %s", got, window)
	}
	for _, want := range []string{"Page 1", "Page 4", "Page 5", "Page 6", "Page 10"} {
		if !strings.Contains(window, `aria-label="`+want+`"`) {
			t.Fatalf("window missing %q in %s", want, window)
		}
	}
	for _, absent := range []string{"Page 2", "Page 3", "Page 7", "Page 8", "Page 9"} {
		if strings.Contains(window, `aria-label="`+absent+`"`) {
			t.Fatalf("window kept %q in %s", absent, window)
		}
	}
	if page := chadcn.Pagination(chadcn.PaginationOptions{}); page != nil {
		t.Fatal("empty pagination rendered an element")
	}
}
