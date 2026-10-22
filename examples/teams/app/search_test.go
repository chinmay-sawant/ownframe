package app

import (
	"context"
	"strings"
	"testing"
)

// TestFilesSearchNarrowsData checks the Files search box filters the table.
func TestFilesSearchNarrowsData(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	openSection(t, app, "files")
	all := len(app.View().Files.Files)

	clickID(t, app, "files-search")

	if err := app.Page().Testing().Type(context.Background(), "vibranium"); err != nil {
		t.Fatalf("Type: %v", err)
	}

	files := app.View().Files.Files
	if len(files) == 0 || len(files) >= all {
		t.Fatalf("search list = %d of %d", len(files), all)
	}

	for _, f := range files {
		if !strings.Contains(strings.ToLower(f.Name), "vibranium") {
			t.Fatalf("%s does not match the query", f.Name)
		}
	}
}

// TestCallsSearchNarrowsData checks the Calls search box filters history.
func TestCallsSearchNarrowsData(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	openSection(t, app, "calls")
	all := len(app.View().Calls.History)

	clickID(t, app, "calls-search")

	if err := app.Page().Testing().Type(context.Background(), "shuri"); err != nil {
		t.Fatalf("Type: %v", err)
	}

	history := app.View().Calls.History
	if len(history) == 0 || len(history) >= all {
		t.Fatalf("search list = %d of %d", len(history), all)
	}

	for _, c := range history {
		if !strings.Contains(strings.ToLower(c.Name), "shuri") {
			t.Fatalf("%s does not match the query", c.Name)
		}
	}
}
