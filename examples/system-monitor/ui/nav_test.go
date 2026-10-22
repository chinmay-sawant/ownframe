package ui

import "testing"

func TestClickNavigationAndTheme(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)

	clickBox(t, app, boxByAction(t, app, actProcesses))

	if app.state.nav != "processes" {
		t.Fatalf("nav after click = %q", app.state.nav)
	}

	if _, ok := boxByID(app.page.Boxes(), "proc-search"); !ok {
		t.Fatal("process screen did not render the search box")
	}

	clickBox(t, app, boxByAction(t, app, prefixSort+string(sortName)))

	if app.state.table.key != sortName {
		t.Fatalf("sort after click = %q", app.state.table.key)
	}

	clickBox(t, app, boxByAction(t, app, actTheme))

	if !app.state.dark {
		t.Fatal("theme click did not flip to dark")
	}

	if app.page.Display() == nil {
		t.Fatal("dark theme made the page fall back to the bitmap path")
	}
}
