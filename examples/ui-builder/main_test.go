package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestBuilderDashboardRendersInteractiveSamples(t *testing.T) {
	t.Parallel()
	a, err := newMonitorApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	for _, word := range []string{
		"System overview", "CPU utilization", "System memory",
		"Disk usage", "CPU history", "Memory history",
		"data-action=\"refresh\"", "data-action=\"view:charts\"",
		"data-action=\"metric:cpu\"", "data-action=\"file:main.go\"",
		"not live system telemetry",
	} {
		if !strings.Contains(html, word) {
			t.Errorf("dashboard HTML missing %q", word)
		}
	}
	if strings.Contains(html, "<script") {
		t.Fatal("dashboard should not depend on scripts")
	}
	if _, err := ui.Render(a.screen()); err != nil {
		t.Fatalf("builder cannot serialize dashboard: %v", err)
	}
}

func TestBuilderDashboardActionRoutingAndRefresh(t *testing.T) {
	t.Parallel()
	a, err := newMonitorApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"view:charts", "metric:memory", "refresh", "file:README.md"} {
		if err := pagetest.ClickAction(ctx, a.page, action); err != nil {
			t.Fatalf("click %s: %v", action, err)
		}
	}
	if a.tab != "charts" || a.metric != "Memory" || a.sample != 1 || a.file != "README.md" {
		t.Fatalf("unexpected dashboard state: %+v", a)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "cancel"}); err != nil {
		t.Fatal(err)
	}
	if a.tab != "overview" || a.metric != "CPU" || a.file != "main.go" {
		t.Fatal("cancel should reset the demo selection")
	}
}
