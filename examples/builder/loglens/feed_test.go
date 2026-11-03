package main

import (
	"testing"
	"time"
)

func TestDemoTelemetryRefreshAddsNewestFirstAndPaginates(t *testing.T) {
	now := time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC)
	feed := newTelemetryFeed(func() time.Time { return now }, 7)
	initial, initialCount, initialPages := feed.logPage(0)
	if len(initial) != demoFeedPageSize || initialCount != len(sampleLogs) || initialPages != 3 {
		t.Fatalf("initial log page = %d rows, %d total, %d pages", len(initial), initialCount, initialPages)
	}
	added := feed.refresh()
	latest, total, pages := feed.logPage(0)
	if added < 1 || added > 3 {
		t.Fatalf("refresh added %d logs, want 1 to 3", added)
	}
	if total != initialCount+added || pages != 3 {
		t.Fatalf("refreshed feed = %d rows across %d pages, want %d rows and 3 pages", total, pages, initialCount+added)
	}
	if latest[0].Time != now.Format("15:04:05.000") || latest[0].Trace == initial[0].Trace {
		t.Fatalf("newest page did not start with a fresh record: first=%+v initial=%+v", latest[0], initial[0])
	}
	older, _, _ := feed.logPage(pages - 1)
	if len(older) == 0 || older[0].Trace == latest[0].Trace {
		t.Fatalf("oldest page is empty or overlaps the newest page: %+v", older)
	}
	clamped, _, _ := feed.logPage(pages + 10)
	if len(clamped) == 0 || clamped[0].Trace != older[0].Trace {
		t.Fatal("out-of-range pages should clamp to the oldest page")
	}
}

func TestDemoTelemetryTracePageMatchesNewestLog(t *testing.T) {
	now := time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC)
	feed := newTelemetryFeed(func() time.Time { return now }, 11)
	feed.refresh()
	logs, _, _ := feed.logPage(0)
	traces, _, _ := feed.tracePage(0)
	if len(logs) == 0 || len(traces) == 0 {
		t.Fatal("expected generated logs and traces")
	}
	if traces[0].ID != logs[0].Trace || !traces[0].Timestamp.Equal(now) {
		t.Fatalf("newest trace does not correspond to newest log: trace=%+v log=%+v", traces[0], logs[0])
	}
	if len(traces[0].Spans) != len(sampleSpans) {
		t.Fatalf("trace has %d spans, want %d", len(traces[0].Spans), len(sampleSpans))
	}
}

func TestDemoTelemetryBoundsRetainedHistory(t *testing.T) {
	now := time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC)
	feed := newTelemetryFeed(func() time.Time { return now }, 3)
	for range 250 {
		feed.refresh()
	}
	_, logCount, _ := feed.logPage(0)
	_, traceCount, _ := feed.tracePage(0)
	if logCount != demoFeedLimit || traceCount != demoFeedLimit {
		t.Fatalf("history counts = logs %d, traces %d; want both capped at %d", logCount, traceCount, demoFeedLimit)
	}
}
