package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	demoFeedPageSize     = 10
	demoFeedLimit        = 500
	feedMaxNewRecords    = 3
	feedPercentCap       = 100
	feedRegionJitterSpan = 11
	feedRegionJitterMid  = 5
	feedDurationSpanMS   = 1800
	feedDurationBaseMS   = 12
)

// telemetryTrace is one generated trace and its waterfall. It is local demo
// telemetry, not a connection to a production collector.
type telemetryTrace struct {
	ID        string
	Name      string
	Service   string
	Owner     string
	Timestamp time.Time
	Duration  string
	Status    string
	Spans     []traceSpan
}

// telemetryFeed keeps a bounded, newest-first stream for the LogLens showcase.
// Refresh adds changing sample records until the app is connected to a real
// telemetry backend.
type telemetryFeed struct {
	now      func() time.Time
	random   *rand.Rand
	sequence uint64
	logs     []logEntry
	traces   []telemetryTrace
	regions  []float64
}

func newTelemetryFeed(now func() time.Time, seed int64) *telemetryFeed {
	if now == nil {
		now = time.Now
	}
	feed := &telemetryFeed{
		now:     now,
		random:  rand.New(rand.NewPCG(uint64(seed), uint64(seed)>>1)), //nolint:gosec // reason: deterministic demo feed shuffle, not security-sensitive.
		logs:    make([]logEntry, 0, demoFeedLimit),
		traces:  make([]telemetryTrace, 0, demoFeedLimit),
		regions: overviewRegionErrors(),
	}
	started := now()
	for index, entry := range sampleLogs {
		stamp := started.Add(-time.Duration(index) * 15 * time.Second)
		entry.Time = stamp.Format("15:04:05.000")
		feed.logs = append(feed.logs, entry)
		feed.traces = append(feed.traces, demoTraceFromLog(entry, stamp, index))
	}
	return feed
}

// Refresh adds one to three generated log and trace records to the top of the
// stream and returns the number added. The deterministic seed keeps showcase
// runs reproducible while each refresh evolves the data.
func (feed *telemetryFeed) refresh() int {
	count := feed.random.IntN(feedMaxNewRecords) + 1
	refreshedAt := feed.now()
	for index := count - 1; index >= 0; index-- {
		feed.sequence++
		entry, trace := feed.generatedRecord(refreshedAt.Add(-time.Duration(index) * time.Millisecond))
		feed.logs = append([]logEntry{entry}, feed.logs...)
		feed.traces = append([]telemetryTrace{trace}, feed.traces...)
		point := feed.random.IntN(len(feed.regions))
		feed.regions[point] = float64(max(0, min(feedPercentCap, int(feed.regions[point])+feed.random.IntN(feedRegionJitterSpan)-feedRegionJitterMid)))
	}
	feed.logs = keepNewestLogs(feed.logs, demoFeedLimit)
	feed.traces = keepNewestTraces(feed.traces, demoFeedLimit)
	return count
}

func (feed *telemetryFeed) regionErrors() []float64 {
	return append([]float64(nil), feed.regions...)
}

// LogPage returns a copied page of logs, the total matching-free record count,
// and the page count. Page zero is always the newest page.
func (feed *telemetryFeed) logPage(page int) ([]logEntry, int, int) {
	return pageSlice(feed.logs, page, demoFeedPageSize)
}

// TracePage returns a copied page of traces, the total record count, and the
// page count. Page zero is always the newest page.
func (feed *telemetryFeed) tracePage(page int) ([]telemetryTrace, int, int) {
	return pageSlice(feed.traces, page, demoFeedPageSize)
}

func (feed *telemetryFeed) generatedRecord(at time.Time) (logEntry, telemetryTrace) {
	services := [...]string{"api-gateway", "auth", "billing", "worker", "search", "web"}
	owners := [...]string{"AK", "RB", "LT", "MC"}
	events := [...]struct {
		level   string
		message string
		status  int
	}{
		{"info", "request completed successfully", 200},
		{"info", "cache entry refreshed", 200},
		{"warn", "upstream latency exceeded the service target", 200},
		{"warn", "retry scheduled after transient failure", 503},
		{"error", "upstream request timed out", 504},
		{"error", "connection reset while reading response", 502},
	}
	service := services[feed.random.IntN(len(services))]
	event := events[feed.random.IntN(len(events))]
	id := fmt.Sprintf("%08x", feed.sequence)
	durationMS := feed.random.IntN(feedDurationSpanMS) + feedDurationBaseMS
	duration := fmt.Sprintf("%dms", durationMS)
	entry := logEntry{
		Time:     at.Format("15:04:05.000"),
		Level:    event.level,
		Service:  service,
		Message:  event.message,
		Trace:    id,
		Duration: duration,
		Status:   event.status,
	}
	trace := demoTraceFromLog(entry, at, feed.random.IntN(len(owners)))
	trace.Owner = owners[feed.random.IntN(len(owners))]
	trace.Duration = duration
	return entry, trace
}

func demoTraceFromLog(entry logEntry, at time.Time, variant int) telemetryTrace {
	spans := append([]traceSpan(nil), sampleSpans...)
	for index := range spans {
		spans[index].Status = "success"
		spans[index].Service = entry.Service
		spans[index].Duration = entry.Duration
	}
	status := "success"
	switch entry.Level {
	case "warn":
		status = "warning"
	case "error":
		status = "error"
	}
	if len(spans) > 0 && status != "success" {
		spans[len(spans)-1].Status = status
	}
	name := [...]string{"GET /api/logs", "POST /api/events", "GET /api/search", "PUT /api/invoices"}[variant%4]
	return telemetryTrace{
		ID:        entry.Trace,
		Name:      name,
		Service:   entry.Service,
		Owner:     "AK",
		Timestamp: at,
		Duration:  entry.Duration,
		Status:    status,
		Spans:     spans,
	}
}

func pageSlice[T any](records []T, page, pageSize int) ([]T, int, int) {
	count := len(records)
	pages := (count + pageSize - 1) / pageSize
	if pages == 0 {
		return nil, count, 0
	}
	page = max(0, min(page, pages-1))
	start := page * pageSize
	end := min(start+pageSize, count)
	return append([]T(nil), records[start:end]...), count, pages
}

func keepNewestLogs(records []logEntry, limit int) []logEntry {
	if len(records) > limit {
		return records[:limit]
	}
	return records
}

func keepNewestTraces(records []telemetryTrace, limit int) []telemetryTrace {
	if len(records) > limit {
		return records[:limit]
	}
	return records
}
