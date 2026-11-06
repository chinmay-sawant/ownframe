// Package benchscore compares two `go test -bench` outputs and applies the
// same class budgets to either project.
//
// Every benchmark row is placed in a class by a name rule or by an optional
// JSON manifest. The time, allocation, and retained-heap medians of the
// candidate are then held against the reference. A row fails when it exceeds
// its class budget.
package benchscore

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/perf/benchfmt"
)

// Metric names use the units printed by `go test -bench` where possible.
// Retained-heap metrics keep the unit the benchmark reported, as long as the
// unit names a heap or retained quantity.
const (
	MetricTime   = "ns/op"
	MetricBytes  = "B/op"
	MetricAllocs = "allocs/op"
)

const (
	benchmarkPrefix = "Benchmark"
	nanosPerSecond  = 1e9
	medianDivisor   = 2
)

// Benchmark is one benchmark name with the median of every comparable metric
// across all result lines that shared the name.
type Benchmark struct {
	Name   string
	Values map[string]float64
}

// Parse reads Go benchmark text and returns one Benchmark per name.
//
// Repeated results for the same name, such as several -count runs, collapse
// into one median per unit. Benchmarks with no comparable metric are skipped.
func Parse(r io.Reader, source string) ([]Benchmark, error) {
	reader := benchfmt.NewReader(r, source)
	values := make(map[string]map[string][]float64)
	var order []string

	for reader.Scan() {
		switch record := reader.Result().(type) {
		case *benchfmt.Result:
			name := benchmarkPrefix + record.Name.String()
			samples, ok := values[name]
			if !ok {
				samples = make(map[string][]float64)
				values[name] = samples
				order = append(order, name)
			}
			for _, value := range record.Values {
				metric, scaled, ok := canonicalMetric(value)
				if !ok {
					continue
				}
				samples[metric] = append(samples[metric], scaled)
			}
		case *benchfmt.SyntaxError:
			return nil, fmt.Errorf("%s: %w", source, record)
		}
	}
	if err := reader.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", source, err)
	}

	benchmarks := make([]Benchmark, 0, len(order))
	for _, name := range order {
		samples := values[name]
		if len(samples) == 0 {
			continue
		}
		medians := make(map[string]float64, len(samples))
		for metric, runs := range samples {
			medians[metric] = median(runs)
		}
		benchmarks = append(benchmarks, Benchmark{Name: name, Values: medians})
	}
	return benchmarks, nil
}

// ParseFile reads a benchmark output file.
func ParseFile(path string) ([]Benchmark, error) {
	data, err := os.ReadFile(path) //nolint:gosec // reason: the path is an explicit CLI input.
	if err != nil {
		return nil, fmt.Errorf("read benchmark file %q: %w", path, err)
	}
	return Parse(bytes.NewReader(data), path)
}

// canonicalMetric maps a benchfmt value to a comparable metric. Benchfmt
// tidies ns/op to sec/op, so the time value is scaled back to nanoseconds.
func canonicalMetric(value benchfmt.Value) (string, float64, bool) {
	switch value.Unit {
	case "sec/op":
		if value.OrigUnit == "ns/op" {
			return MetricTime, value.OrigValue, true
		}
		return MetricTime, value.Value * nanosPerSecond, true
	case "ns/op":
		return MetricTime, value.Value, true
	case MetricBytes:
		return MetricBytes, value.Value, true
	case MetricAllocs:
		return MetricAllocs, value.Value, true
	}
	if isRetainedUnit(value.Unit) || isRetainedUnit(value.OrigUnit) {
		unit := value.Unit
		if unit == "" {
			unit = value.OrigUnit
		}
		return unit, value.Value, true
	}
	return "", 0, false
}

func isRetainedUnit(unit string) bool {
	unit = strings.ToLower(strings.TrimSpace(unit))
	return strings.Contains(unit, "heap") || strings.Contains(unit, "retained") || strings.Contains(unit, "retention")
}

func median(runs []float64) float64 {
	sorted := append([]float64(nil), runs...)
	sort.Float64s(sorted)
	middle := len(sorted) / medianDivisor
	if len(sorted)%medianDivisor == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / medianDivisor
}

func metricOrder(values map[string]float64) []string {
	order := make([]string, 0, len(values))
	for _, metric := range []string{MetricTime, MetricBytes, MetricAllocs} {
		if _, ok := values[metric]; ok {
			order = append(order, metric)
		}
	}
	var retained []string
	for metric := range values {
		if metric == MetricTime || metric == MetricBytes || metric == MetricAllocs {
			continue
		}
		retained = append(retained, metric)
	}
	sort.Strings(retained)
	return append(order, retained...)
}
