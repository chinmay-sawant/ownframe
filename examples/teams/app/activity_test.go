package app

import "testing"

// activityApp opens the app on the activity menu.
func activityApp(t *testing.T) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	openSection(t, app, "activity")

	return app
}

// activityClick clicks the box with the id.
func activityClick(t *testing.T, app *App, id string) {
	t.Helper()

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no %s box", id)
	}

	clickBox(t, app, box)
}

func TestActivityOpen(t *testing.T) {
	t.Parallel()
	app := activityApp(t)
	activityClick(t, app, "activity-item-a1")

	if v := app.View().Activity; v.Active != "a1" || v.Items[0].Unread {
		t.Fatalf("Active = %q, a1 unread = %v", v.Active, v.Items[0].Unread)
	}
}

func TestActivityFilter(t *testing.T) {
	t.Parallel()
	app := activityApp(t)

	for _, tc := range []struct{ filter, show, hide string }{
		{"mentions", "activity-item-a1", "activity-item-a2"},
		{"replies", "activity-item-a2", "activity-item-a1"},
	} {
		activityClick(t, app, "activity-filter-"+tc.filter)

		if v := app.View().Activity; v.Filter != tc.filter {
			t.Fatalf("Filter = %q", v.Filter)
		}

		if _, ok := findBox(app.Boxes(), tc.show); !ok {
			t.Fatalf("%s hidden", tc.show)
		}

		if _, ok := findBox(app.Boxes(), tc.hide); ok {
			t.Fatalf("%s shown", tc.hide)
		}
	}
}

func TestActivityReply(t *testing.T) {
	t.Parallel()
	app := activityApp(t)
	activityClick(t, app, "activity-item-a1")
	activityClick(t, app, "activity-reply")

	if it := app.View().Activity.Items[0]; !it.Replied || it.Replies != 3 {
		t.Fatalf("Replied = %v, Replies = %d", it.Replied, it.Replies)
	}

	if _, ok := findBox(app.Boxes(), "activity-reply"); ok {
		t.Fatal("reply button still shown")
	}
}

func TestActivityMarkRead(t *testing.T) {
	t.Parallel()
	app := activityApp(t)
	activityClick(t, app, "activity-mark")

	for _, it := range app.View().Activity.Items {
		if it.Unread {
			t.Fatalf("%s still unread", it.ID)
		}
	}
}
