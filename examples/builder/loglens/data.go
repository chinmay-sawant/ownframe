package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

type logEntry struct {
	Time     string
	Level    string
	Service  string
	Message  string
	Trace    string
	Duration string
	Status   int
}

type serviceHealth struct {
	Name     string
	Status   string
	Requests string
	Errors   string
	Latency  string
	Owner    string
}

type traceSpan struct {
	Name     string
	Service  string
	Detail   string
	Start    float64
	Width    float64
	Status   string
	Duration string
}

type chatMessage struct {
	Name       string
	Time       string
	Text       string
	Side       string
	Initials   string
	Attachment string
}

type telemetryMetrics struct {
	Total      int
	Errors     int
	Warnings   int
	ErrorRate  float64
	P95Latency time.Duration
}

//nolint:gochecknoglobals // static dashboard fixture
var sampleLogs = []logEntry{
	{"14:32:08.412", "error", "api-gateway", "upstream timeout after 30s", "7f3a9c21", "30.0s", 504},
	{"14:31:55.008", "warn", "billing", "retrying invoice sync (attempt 2)", "5d21b8e0", "1.8s", 200},
	{"14:31:42.771", "info", "auth", "token refreshed for user 8412", "a10e77c9", "42ms", 200},
	{"14:31:28.190", "error", "worker", "job queue saturation: 1842 pending", "3bc41d70", "0ms", 503},
	{"14:31:11.604", "info", "web", "GET /dashboard 200", "f2a8b104", "118ms", 200},
	{"14:30:57.332", "warn", "api-gateway", "rate limit crossed for tenant acme", "9d8e2a41", "2ms", 429},
	{"14:30:42.019", "info", "search", "index refresh finished in 2.1s", "8c7d3f06", "2.1s", 200},
	{"14:30:16.884", "error", "billing", "card authorization declined", "6e2f9a12", "640ms", 402},
	{"14:29:51.402", "info", "worker", "nightly export completed", "4a1b7c99", "12.4s", 200},
	{"14:29:33.155", "warn", "web", "slow render: /reports (1.9s)", "2f7d8a50", "1.9s", 200},
	{"14:29:10.782", "info", "auth", "session created for user 5521", "1c8e4b30", "61ms", 200},
	{"14:28:44.918", "info", "search", "query served from cache", "0d4a6f21", "8ms", 200},
	{"14:28:19.227", "error", "api-gateway", "connection reset by peer", "9b1c5e73", "14ms", 502},
	{"14:27:58.640", "info", "web", "GET /settings 200", "7a3d2c18", "94ms", 200},
	{"14:27:31.005", "warn", "worker", "retry budget exhausted for job 9842", "5f8b1d42", "0ms", 500},
	{"14:27:04.883", "info", "billing", "invoice paid: inv_1f2a", "3e7c9b60", "221ms", 200},
	{"14:26:48.120", "info", "auth", "password rotation completed", "2d9a4c88", "310ms", 200},
	{"14:26:20.571", "error", "search", "shard replica unavailable", "1b6e8f35", "0ms", 503},
	{"14:25:59.304", "info", "web", "POST /api/views 201", "9c4f7a02", "143ms", 201},
	{"14:25:33.912", "warn", "api-gateway", "deprecated header X-Tenant used", "8a2d6c14", "1ms", 200},
	{"14:25:08.447", "info", "worker", "report cache warmed", "7f1b3e95", "3.2s", 200},
	{"14:24:41.688", "info", "search", "index shard promoted", "6d8c2a71", "480ms", 200},
	{"14:24:17.250", "error", "auth", "invalid refresh token", "5b3e9f08", "18ms", 401},
	{"14:23:52.630", "info", "web", "GET /logs 200", "4a7c1d86", "76ms", 200},
}

//nolint:gochecknoglobals // static dashboard fixture
var sampleServices = []serviceHealth{
	{"api-gateway", "degraded", "1.2M", "0.42%", "310ms", "Platform"},
	{"auth", "healthy", "480K", "0.01%", "48ms", "Identity"},
	{"billing", "healthy", "211K", "0.08%", "220ms", "Revenue"},
	{"worker", "warning", "96K", "0.31%", "1.1s", "Platform"},
	{"search", "healthy", "640K", "0.02%", "62ms", "Discovery"},
	{"web", "healthy", "820K", "0.03%", "118ms", "Frontend"},
}

//nolint:gochecknoglobals // static dashboard fixture
var sampleSpans = []traceSpan{
	{"GET /api/logs", "web", "gateway", 0, 18, "success", "1.24s"},
	{"auth.verify", "auth", "token lookup", 18, 8, "success", "180ms"},
	{"logs.query", "api-gateway", "postgres primary", 26, 34, "warning", "740ms"},
	{"search.enrich", "search", "index shard 4", 60, 22, "success", "420ms"},
	{"response.write", "web", "gzip stream", 82, 18, "success", "160ms"},
}

//nolint:gochecknoglobals // static dashboard fixture
var sampleMessages = []chatMessage{
	{"Ada", "14:28", "I’m seeing the same 504s on api-gateway. Pulling the trace now.", "start", "AK", ""},
	{"You", "14:29", "The queue is at 1842 pending. I’ll scale the worker pool.", "end", "YO", ""},
	{"Linus", "14:30", "Here’s the heap snapshot from the worker that stalled.", "start", "LT", "worker-heap-14-30.txt"},
	{"You", "14:31", "Thanks. Adding it to the incident timeline.", "end", "YO", ""},
	{"Ada", "14:32", "Customer impact is limited to the EU region. Status page updated.", "start", "AK", ""},
}

var (
	//nolint:gochecknoglobals // static dashboard fixture
	timeRangeOptions = []string{"15 minutes", "1 hour", "24 hours", "7 days", "30 days"}
	//nolint:gochecknoglobals // static dashboard fixture
	serviceOptions = []string{"All services", "api-gateway", "auth", "billing", "search", "web", "worker"}
	//nolint:gochecknoglobals // static dashboard fixture
	severityOptions = []string{"error", "warn", "info", "debug"}
	//nolint:gochecknoglobals // static dashboard fixture
	sortOptions = []string{"Newest first", "Oldest first", "Slowest first", "Status code"}
	//nolint:gochecknoglobals // static dashboard fixture
	channelOptions = []string{"Slack", "Email", "PagerDuty", "Webhook"}
)

func (a *app) visibleLogs() []logEntry {
	records := sampleLogs
	if a.feed != nil {
		records = a.feed.logs
	}
	out := make([]logEntry, 0, len(records))
	query := strings.ToLower(strings.TrimSpace(a.query))
	for _, entry := range records {
		if a.service != "" && a.service != "All services" && entry.Service != a.service {
			continue
		}
		if len(a.severity) > 0 && !containsString(a.severity, entry.Level) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(entry.Message+" "+entry.Service+" "+entry.Trace), query) {
			continue
		}
		out = append(out, entry)
	}
	switch a.sortValue {
	case "Oldest first":
		for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
			out[left], out[right] = out[right], out[left]
		}
	case "Slowest first":
		sort.SliceStable(out, func(i, j int) bool {
			left, _ := time.ParseDuration(out[i].Duration)
			right, _ := time.ParseDuration(out[j].Duration)
			return left > right
		})
	case "Status code":
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Status > out[j].Status
		})
	}
	return out
}

func (a *app) pageLogs() []logEntry {
	logs := a.visibleLogs()
	if len(logs) == 0 {
		return nil
	}
	start := a.pageIndex * logPageSize
	if start >= len(logs) {
		start = (a.logPageCount() - 1) * logPageSize
	}
	end := min(start+logPageSize, len(logs))
	return logs[start:end]
}

func (a *app) logPageCount() int {
	logs := len(a.visibleLogs())
	if logs == 0 {
		return 1
	}
	return (logs + logPageSize - 1) / logPageSize
}

func (a *app) selectedLog() (logEntry, bool) {
	logs := a.visibleLogs()
	if a.drawerIndex < 0 || a.drawerIndex >= len(logs) {
		return logEntry{}, false
	}
	return logs[a.drawerIndex], true
}

func logLevelVariant(level string) chadcn.BadgeVariant {
	switch level {
	case "error":
		return chadcn.BadgeDanger
	case "warn":
		return chadcn.BadgeOutline
	case "info":
		return chadcn.BadgeSecondary
	default:
		return chadcn.BadgeSecondary
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func statusLabel(status int) string {
	return fmt.Sprintf("%d", status)
}

func (a *app) telemetryMetrics() telemetryMetrics {
	if a.feed == nil {
		return telemetryMetrics{}
	}
	metrics := telemetryMetrics{Total: len(a.feed.logs)}
	durations := make([]time.Duration, 0, metrics.Total)
	for _, entry := range a.feed.logs {
		if entry.Level == "error" || entry.Status >= 500 {
			metrics.Errors++
		} else if entry.Level == "warn" {
			metrics.Warnings++
		}
		if duration, err := time.ParseDuration(entry.Duration); err == nil {
			durations = append(durations, duration)
		}
	}
	if metrics.Total > 0 {
		metrics.ErrorRate = 100 * float64(metrics.Errors) / float64(metrics.Total)
	}
	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		index := (95*len(durations)+99)/100 - 1
		metrics.P95Latency = durations[max(0, min(index, len(durations)-1))]
	}
	return metrics
}

func (a *app) liveServices() []serviceHealth {
	if a.feed == nil {
		return sampleServices
	}
	counts := make(map[string]int, len(serviceOptions))
	errors := make(map[string]int, len(serviceOptions))
	durations := make(map[string][]time.Duration, len(serviceOptions))
	for _, entry := range a.feed.logs {
		counts[entry.Service]++
		if entry.Level == "error" || entry.Status >= 500 {
			errors[entry.Service]++
		}
		if duration, err := time.ParseDuration(entry.Duration); err == nil {
			durations[entry.Service] = append(durations[entry.Service], duration)
		}
	}
	services := make([]serviceHealth, 0, len(sampleServices))
	for _, sample := range sampleServices {
		count := counts[sample.Name]
		if count == 0 {
			continue
		}
		rate := 100 * float64(errors[sample.Name]) / float64(count)
		status := "healthy"
		if rate >= 40 {
			status = "degraded"
		} else if rate >= 20 {
			status = "warning"
		}
		values := durations[sample.Name]
		latency := time.Duration(0)
		if len(values) > 0 {
			sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
			index := (95*len(values)+99)/100 - 1
			latency = values[max(0, min(index, len(values)-1))]
		}
		services = append(services, serviceHealth{
			Name: sample.Name, Status: status,
			Requests: fmt.Sprintf("%d events", count),
			Errors:   fmt.Sprintf("%.1f%%", rate),
			Latency:  latency.String(), Owner: sample.Owner,
		})
	}
	return services
}

func (a *app) eventSeries() ([]string, []float64, []float64) {
	services := []string{"api-gateway", "auth", "billing", "worker", "search", "web"}
	current := make([]float64, len(services))
	previous := make([]float64, len(services))
	if a.feed == nil {
		return services, current, previous
	}
	split := min(demoFeedPageSize, len(a.feed.logs))
	for index, entry := range a.feed.logs {
		for serviceIndex, service := range services {
			if entry.Service == service {
				if index < split {
					current[serviceIndex]++
				} else if index < split*2 {
					previous[serviceIndex]++
				}
				break
			}
		}
	}
	return services, current, previous
}
