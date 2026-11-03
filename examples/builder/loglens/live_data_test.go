package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/ui"
)

func TestManualRefreshUpdatesSharedTelemetryAndReturnsToNewestPages(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.pageIndex = 2
	a.tracePageIndex = 2
	a.severity = nil
	before := len(a.feed.logs)
	if err := a.refreshTelemetry(context.Background()); err != nil {
		t.Fatal(err)
	}
	if a.pageIndex != 0 || a.tracePageIndex != 0 {
		t.Fatalf("refresh left old pages selected: logs=%d traces=%d", a.pageIndex, a.tracePageIndex)
	}
	if a.lastRefreshCount < 1 || len(a.feed.logs) != before+a.lastRefreshCount {
		t.Fatalf("refresh count=%d, feed grew %d to %d", a.lastRefreshCount, before, len(a.feed.logs))
	}
	logs := a.pageLogs()
	traces, total, pages := a.feed.tracePage(a.tracePageIndex)
	if len(logs) != demoFeedPageSize || len(traces) != demoFeedPageSize || total != before+a.lastRefreshCount || pages < 1 {
		t.Fatalf("newest page sizes logs=%d traces=%d total=%d pages=%d", len(logs), len(traces), total, pages)
	}
	if logs[0].Trace != traces[0].ID || logs[0].Trace != a.feed.logs[0].Trace {
		t.Fatalf("newest log and trace are not paired: log=%+v trace=%+v", logs[0], traces[0])
	}

	firstPage, err := ui.Render(a.tracesPanel())
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(firstPage, `id="recent-trace-`); got != demoFeedPageSize {
		t.Fatalf("trace page rendered %d rows, want %d", got, demoFeedPageSize)
	}
	if !strings.Contains(firstPage, traces[0].ID) || !strings.Contains(firstPage, "page 1 of") {
		t.Fatal("trace page did not render the newest record and pagination state")
	}
	a.tracePageIndex = 1
	secondPage, err := ui.Render(a.tracesPanel())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(secondPage, `id="recent-trace-`+traces[0].ID+`"`) {
		t.Fatal("trace page two repeated the newest record from page one")
	}
}
