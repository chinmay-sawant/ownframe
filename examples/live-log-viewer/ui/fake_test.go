package ui

import (
	"context"
	"strings"
	"sync"
	"time"
)

// fakeFeed is an in-memory Feed for headless tests.
type fakeFeed struct {
	mu       sync.Mutex
	entries  []Entry
	sources  []SourceInfo
	settings Settings
	delay    time.Duration
	pages    int
	exports  int
}

// seed fills the feed with n entries, IDs 1..n, mixed severities.
func (f *fakeFeed) seed(n int) {
	sevs := []string{"debug", "info", "warn", "error"}
	f.mu.Lock()
	defer f.mu.Unlock()

	f.sources = []SourceInfo{{Key: "s1", Name: "app.log", State: "live", Count: n}}
	f.settings = Settings{Follow: true}

	for i := 1; i <= n; i++ {
		f.entries = append(f.entries, Entry{
			ID: int64(i), TimeOK: true, Time: time.Unix(int64(i), 0),
			Source: "app.log", SourceKey: "s1", Severity: sevs[i%len(sevs)],
			Text: "line " + strings.Repeat("x", i%7), Lines: 1,
		})
	}
}

// add appends entries as a live source would.
func (f *fakeFeed) add(es ...Entry) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.entries = append(f.entries, es...)
}

func (f *fakeFeed) wait(ctx context.Context) error {
	if f.delay <= 0 {
		return ctx.Err()
	}

	select {
	case <-time.After(f.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *fakeFeed) match(e Entry, q Query) bool {
	if q.MaxID > 0 && e.ID > q.MaxID {
		return false
	}

	if q.FromID > 0 && e.ID < q.FromID {
		return false
	}

	if q.ToID > 0 && e.ID > q.ToID {
		return false
	}

	if len(q.Sources) > 0 && !inList(q.Sources, e.SourceKey) {
		return false
	}

	if q.MinSev != "" && sevRank(e.Severity) < sevRank(q.MinSev) {
		return false
	}

	if q.Text != "" && !strings.Contains(strings.ToLower(e.Text), strings.ToLower(q.Text)) {
		return false
	}

	return true
}
