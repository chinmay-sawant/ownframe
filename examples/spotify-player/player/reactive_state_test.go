package player

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestReactiveApplyStateUpdate(t *testing.T) {
	api := &testAPI{}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)

	app, err := NewAt(srv.URL)
	if err != nil {
		t.Fatalf("NewAt: %v", err)
	}
	redraw(t, app)

	app.Page().ClearChangedProperties()
	if err := app.Load(context.Background(), "canned"); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(app.Page().ChangedProperties()) == 0 {
		t.Fatal("Load left no changed properties")
	}
	if got := app.View().Status; !containsSubstring(got, "Live") {
		t.Fatalf("Status = %q", got)
	}
	redraw(t, app)
}

func TestReactiveOfflineStateUpdate(t *testing.T) {
	app, err := NewAt("http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("NewAt: %v", err)
	}
	redraw(t, app)

	app.Page().ClearChangedProperties()
	if err := app.Load(context.Background(), "canned"); err == nil {
		t.Fatal("Load: want an error")
	}
	if len(app.Page().ChangedProperties()) == 0 {
		t.Fatal("offline Load left no changed properties")
	}
	if got := app.View().Status; got != sampleStatus {
		t.Fatalf("Status = %q", got)
	}
	redraw(t, app)
}

func containsSubstring(s, frag string) bool {
	for i := 0; i+len(frag) <= len(s); i++ {
		if s[i:i+len(frag)] == frag {
			return true
		}
	}
	return false
}
