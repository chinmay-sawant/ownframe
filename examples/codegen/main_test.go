package main

import (
	"context"
	"strings"
	"testing"
)

func displayHasText(t *testing.T, app *App, want string) bool {
	t.Helper()
	for _, op := range app.Page().Probe().Display().Ops {
		if strings.Contains(op.Text, want) {
			return true
		}
	}

	return false
}

func TestCodegenDashboardActions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := New(false, false)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("initial Redraw failed: %v", err)
	}

	for _, want := range []string{
		"OwnFrame Codegen Dashboard",
		"Online",
		"Embedded ui.Element slot",
		"Compile view.html with ownframegen",
	} {
		if !displayHasText(t, app, want) {
			t.Fatalf("initial display missing %q", want)
		}
	}

	cases := []struct {
		action   string
		wantText string
	}{
		{ActionToggleStatus, "Offline"},
		{ActionToggleStatus, "Online"},
		{ActionIncrementDone, "Completed task #2"},
		{ActionAppendTask, "Generated task #5"},
		{ActionResetTasks, "Reset task list"},
	}

	for _, tc := range cases {
		if err := app.handleAction(ctx, tc.action); err != nil {
			t.Fatalf("action %q failed: %v", tc.action, err)
		}
		if !displayHasText(t, app, tc.wantText) {
			t.Fatalf("after action %q, display missing %q", tc.action, tc.wantText)
		}
	}

	// Verify real click dispatch via Page.Boxes() hits the generated data-action.
	var clickX, clickY float64
	found := false
	for _, b := range app.Page().Boxes() {
		if b.Action == ActionAppendTask {
			clickX = b.X + b.W/2
			clickY = b.Y + b.H/2
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("box for action %q not found", ActionAppendTask)
	}
	if err := app.Page().Click(ctx, clickX, clickY); err != nil {
		t.Fatalf("Click(%s) failed: %v", ActionAppendTask, err)
	}
	if !displayHasText(t, app, "Generated task #5") {
		t.Fatal("Click(append-task) did not paint Generated task #5")
	}
}

func TestCodegenEscaping(t *testing.T) {
	t.Parallel()

	v := NewDashboardView().
		WithTitle(`<script>alert("xss")</script>`).
		WithBannerHTML(`<div class="safe-slot">ok</div>`)

	out := v.RenderString()
	if strings.Contains(out, `<script>alert`) {
		t.Fatal("untrusted title was not HTML-escaped")
	}
	if !strings.Contains(out, `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`) {
		t.Fatal("expected escaped title in rendered HTML")
	}
	if !strings.Contains(out, `<div class="safe-slot">ok</div>`) {
		t.Fatal("raw BannerHTML slot was unexpectedly escaped")
	}
}
