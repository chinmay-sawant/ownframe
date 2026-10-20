package app

import "testing"

// TestFilesOpens checks the Files section starts on the recent filter.
func TestFilesOpens(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	if app.View().Section != "chat" {
		t.Fatalf("Section = %q", app.View().Section)
	}

	openSection(t, app, "files")

	for _, id := range []string{"files-item-f1", "files-row-f1"} {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no %s box", id)
		}
	}

	if d := app.View().Files; d.Filter != "recent" || len(d.Files) != 10 {
		t.Fatalf("d = %+v", d)
	}
}

// TestFilesStarred checks the starred filter and its unstar drop.
func TestFilesStarred(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	openSection(t, app, "files")

	clickID(t, app, "files-filter-starred")

	if d := app.View().Files; d.Filter != "starred" || len(d.Files) != 3 {
		t.Fatalf("d = %+v", d)
	}

	clickID(t, app, "files-item-f3")
	clickID(t, app, "files-detail-star-f3")

	if d := app.View().Files; d.Active != "" || len(d.Files) != 2 {
		t.Fatalf("d = %+v", d)
	}
}

// TestFilesToggle checks a row star flips f2 and a row click shows detail.
func TestFilesToggle(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	openSection(t, app, "files")

	clickID(t, app, "files-star-f2")

	f2 := false

	for _, f := range app.View().Files.Files {
		if f.ID == "f2" {
			f2 = f.Starred
		}
	}

	if !f2 {
		t.Fatal("f2 is not starred")
	}

	clickID(t, app, "files-row-f1")

	if app.View().Files.Active != "f1" {
		t.Fatalf("Active = %q", app.View().Files.Active)
	}

	if _, ok := findBox(app.Boxes(), "files-detail-star-f1"); !ok {
		t.Fatal("no f1 detail star")
	}
}
