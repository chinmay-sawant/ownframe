package benchscore

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Budgets applied to every project. The reference is the baseline, so a ratio
// above 1.0x is a regression.
const (
	maxSlowdownRatio   = 2.0
	strictRatio        = 1.0
	bytesPerMiB        = 1 << 20
	maxHeapGrowthBytes = bytesPerMiB
	ratioTolerance     = 1e-9
)

// Status is the gate result for one benchmark metric.
type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusNew  Status = "NEW"
)

// Row is one class, benchmark, metric comparison.
type Row struct {
	Class        Class
	Benchmark    string
	Metric       string
	Reference    float64
	Candidate    float64
	HasReference bool
	HasCandidate bool
	Ratio        float64
	Status       Status
	Reason       string
}

// Report holds every compared row in reference order, followed by benchmarks
// that exist only in the candidate.
type Report struct {
	Rows []Row
}

// Passed reports whether every compared row stayed within budget. New
// candidate benchmarks do not fail the gate.
func (r Report) Passed() bool {
	return len(r.FailedRows()) == 0
}

// FailedRows returns the rows that exceeded a budget or lost a benchmark or
// metric from the candidate.
func (r Report) FailedRows() []Row {
	var failed []Row
	for _, row := range r.Rows {
		if row.Status == StatusFail {
			failed = append(failed, row)
		}
	}
	return failed
}

// Compare holds the candidate medians against the reference medians under the
// class budgets. Every reference benchmark must remain in the candidate.
func Compare(reference, candidate []Benchmark, manifest Manifest) Report {
	candidateByName := make(map[string]Benchmark, len(candidate))
	for _, benchmark := range candidate {
		candidateByName[benchmark.Name] = benchmark
	}
	seen := make(map[string]bool, len(reference))
	report := Report{Rows: make([]Row, 0, len(reference))}

	for _, ref := range reference {
		seen[ref.Name] = true
		class, sameData := Classify(ref.Name, manifest)
		cand, ok := candidateByName[ref.Name]
		if !ok {
			report.Rows = append(report.Rows, Row{
				Class:        class,
				Benchmark:    ref.Name,
				Metric:       "-",
				Status:       StatusFail,
				Reason:       "benchmark missing from candidate",
				HasReference: false,
			})
			continue
		}
		for _, metric := range metricOrder(ref.Values) {
			refValue := ref.Values[metric]
			candValue, present := cand.Values[metric]
			if !present {
				report.Rows = append(report.Rows, Row{
					Class:        class,
					Benchmark:    ref.Name,
					Metric:       metric,
					Reference:    refValue,
					HasReference: true,
					Status:       StatusFail,
					Reason:       "metric missing from candidate",
				})
				continue
			}
			ratio, reasons := judge(metric, refValue, candValue, class, sameData)
			status := StatusPass
			if len(reasons) > 0 {
				status = StatusFail
			}
			report.Rows = append(report.Rows, Row{
				Class:        class,
				Benchmark:    ref.Name,
				Metric:       metric,
				Reference:    refValue,
				Candidate:    candValue,
				HasReference: true,
				HasCandidate: true,
				Ratio:        ratio,
				Status:       status,
				Reason:       strings.Join(reasons, "; "),
			})
		}
	}

	var newNames []string
	for _, benchmark := range candidate {
		if !seen[benchmark.Name] {
			newNames = append(newNames, benchmark.Name)
		}
	}
	sort.Strings(newNames)
	for _, name := range newNames {
		cand := candidateByName[name]
		class, _ := Classify(name, manifest)
		for _, metric := range metricOrder(cand.Values) {
			report.Rows = append(report.Rows, Row{
				Class:        class,
				Benchmark:    name,
				Metric:       metric,
				Candidate:    cand.Values[metric],
				HasCandidate: true,
				Status:       StatusNew,
				Reason:       "new benchmark in candidate",
			})
		}
	}
	return report
}

func judge(metric string, reference, candidate float64, class Class, sameData bool) (float64, []string) {
	ratio := ratioOf(reference, candidate)
	if isRetainedUnit(metric) {
		var reasons []string
		if ratio > maxSlowdownRatio+ratioTolerance {
			reasons = append(reasons, fmt.Sprintf("%s %.2fx > %.2fx", metric, ratio, maxSlowdownRatio))
		}
		if growth := candidate - reference; growth > maxHeapGrowthBytes {
			reasons = append(reasons, fmt.Sprintf(
				"retained heap growth %.0f B > %d B (%.2f MiB)",
				growth, maxHeapGrowthBytes, float64(maxHeapGrowthBytes)/bytesPerMiB,
			))
		}
		return ratio, reasons
	}

	strict := class == ClassIdle || sameData
	limit := maxSlowdownRatio
	if strict {
		limit = strictRatio
	}
	if strict && metric == MetricAllocs && candidate > 0 {
		return ratio, []string{fmt.Sprintf("allocs %.0f, idle and same-data rows must stay at 0", candidate)}
	}
	if ratio > limit+ratioTolerance {
		label := metric
		if strict {
			label = "idle/same-data " + metric
		}
		return ratio, []string{fmt.Sprintf("%s %.2fx > %.2fx", label, ratio, limit)}
	}
	return ratio, nil
}

func ratioOf(reference, candidate float64) float64 {
	if reference == 0 {
		if candidate == 0 {
			return 1
		}
		return math.Inf(1)
	}
	if reference < 0 {
		return math.Inf(1)
	}
	return candidate / reference
}
