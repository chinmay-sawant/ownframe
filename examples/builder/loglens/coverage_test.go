package main

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/ui"
)

// TestLoglensRendersEveryComponentRoot renders every panel and overlay from
// host state and checks that the combined markup contains every ChadCN
// component root. A root is matched as a whole class token, so rendering only
// .of-alert-dialog does not satisfy the .of-alert check.
func TestLoglensRendersEveryComponentRoot(t *testing.T) {
	roots := []string{
		// Primitives and display.
		".of-aspect-ratio", ".of-avatar", ".of-badge", ".of-card", ".of-direction",
		".of-empty", ".of-item", ".of-kbd", ".of-label", ".of-progress",
		".of-separator", ".of-skeleton", ".of-spinner", ".of-table", ".of-typography",
		// Actions and forms.
		".of-button", ".of-button-group", ".of-toggle", ".of-toggle-group", ".of-field",
		".of-input", ".of-textarea", ".of-native-select", ".of-checkbox", ".of-radio-group",
		".of-switch", ".of-slider", ".of-input-group", ".of-input-otp",
		// Feedback and disclosure.
		".of-alert", ".of-toast", ".of-accordion", ".of-collapsible", ".of-tabs",
		".of-breadcrumb", ".of-pagination",
		// Overlays and menus.
		".of-dialog", ".of-alert-dialog", ".of-sheet", ".of-drawer", ".of-popover",
		".of-tooltip", ".of-hover-card", ".of-dropdown-menu", ".of-context-menu",
		".of-menubar", ".of-navigation-menu", ".of-command",
		// Selection.
		".of-select", ".of-combobox", ".of-calendar", ".of-date-picker",
		// Host-owned composites.
		".of-scroll-area", ".of-resizable", ".of-sidebar", ".of-carousel",
		".of-data-table", ".of-chart", ".of-code-block", ".of-message", ".of-bubble",
		".of-attachment", ".of-marker", ".of-message-scroller", ".of-questionnaire",
		".of-trace",
	}
	if len(roots) != 66 {
		t.Fatalf("component root list has %d entries, want 66", len(roots))
	}

	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}

	var markup strings.Builder
	render := func(name string, element *ui.Element) {
		t.Helper()
		if element == nil {
			t.Fatalf("%s returned nil", name)
		}
		fragment, renderErr := ui.Render(element)
		if renderErr != nil {
			t.Fatalf("ui.Render(%s): %v", name, renderErr)
		}
		markup.WriteString(fragment)
	}

	// Overview. loading turns on the skeleton and spinner states.
	a.tab = "overview"
	a.loading = true
	render("overviewPanel", a.overviewPanel())

	// Logs. The row menu, detail drawer, and all three filter pickers are open
	// so their surfaces render, not only their triggers.
	a.tab = "logs"
	a.loading = false
	a.contextOpen = true
	a.contextX, a.contextY = -150, 34
	a.drawerOpen = true
	a.drawerIndex = 0
	a.selectOpen = true
	a.comboboxOpen = true
	a.dateOpen = true
	render("logsPanel", a.logsPanel())
	render("logFilters", a.logFilters())

	// Traces. RTL flips the Arabic card; the split, scroll area, and waterfall
	// render in either mode.
	a.tab = "traces"
	a.rtl = true
	render("tracesPanel", a.tracesPanel())

	// Incident room seeds its own accordion on the first render.
	a.tab = "incident"
	render("incidentPanel", a.incidentPanel())

	// Topbar menus and popovers.
	a.navOpen = "product-platform"
	a.menubarOpen = "menu-file"
	a.userMenuOpen = true
	render("topMenus", a.topMenus())

	a.tooltipOpen = true
	a.hoverOpen = true
	a.notifOpen = true
	render("topPopovers", a.topPopovers())

	// Settings sheet.
	a.settingsOpen = true
	render("settingsSheet", a.settingsSheet())

	// Page-level dialog, alert dialog, and toast.
	a.dialogOpen = true
	a.alertOpen = true
	a.toastOpen = true
	a.toastTitle = "Filters applied"
	a.toastBody = "24 hours · All services · Newest first"
	render("globalDialogs", a.globalDialogs())

	// Command palette.
	a.commandOpen = true
	a.commandActive = "overview"
	render("commandOverlay", a.commandOverlay())

	// The shell owns the sidebar, breadcrumb, and tab list, so render the whole
	// screen once as well. The active tab panel repeats incidentPanel, which is
	// harmless; this call is what puts .of-sidebar, .of-breadcrumb, and
	// .of-tabs in the fragment.
	render("screen", a.screen())

	html := markup.String()
	if strings.TrimSpace(html) == "" {
		t.Fatal("rendered dashboard is empty")
	}
	for _, root := range roots {
		if !loglensHasClassToken(html, strings.TrimPrefix(root, ".")) {
			t.Errorf("rendered dashboard is missing component root %s", root)
		}
	}
}

// loglensHasClassToken reports whether html contains want as one whole class
// token.
func loglensHasClassToken(html, want string) bool {
	const prefix = `class="`
	for {
		start := strings.Index(html, prefix)
		if start < 0 {
			return false
		}
		html = html[start+len(prefix):]
		end := strings.IndexByte(html, '"')
		if end < 0 {
			return false
		}
		for _, token := range strings.Fields(html[:end]) {
			if token == want {
				return true
			}
		}
		html = html[end+1:]
	}
}
