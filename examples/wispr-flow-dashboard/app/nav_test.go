package app

import (
	"context"
	"testing"
)

// openInsights clicks the sidebar's Insights link and leaves the dashboard
// on screen.
func openInsights(t *testing.T, app *App) {
	t.Helper()

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	nav, ok := findBox(app.Boxes(), "nav-insights")
	if !ok {
		t.Fatal("no nav-insights box")
	}

	clickBox(t, app, nav)
}

// TestSidebarNavigation checks the sidebar starts open on the dictation
// page, collapses on the toggle, and switches pages on a nav click.
func TestSidebarNavigation(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if app.View().ActivePage != "dictation" {
		t.Fatalf("ActivePage = %q", app.View().ActivePage)
	}

	toggle, ok := findBox(app.Boxes(), "sidebar-toggle")
	if !ok {
		t.Fatal("no sidebar-toggle box")
	}

	clickBox(t, app, toggle)

	if !app.View().SidebarCollapsed {
		t.Fatal("sidebar did not collapse")
	}

	openInsights(t, app)

	if app.View().ActivePage != "insights" {
		t.Fatalf("ActivePage = %q", app.View().ActivePage)
	}
}
