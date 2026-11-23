package benchscore

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMedians(t *testing.T) {
	const text = `
goos: linux
goarch: amd64
pkg: example.com/demo

BenchmarkIdleFrame-8   100   100000 ns/op   0 B/op   0 allocs/op
BenchmarkIdleFrame-8   100   140000 ns/op   0 B/op   0 allocs/op
BenchmarkIdleFrame-8   100   120000 ns/op   0 B/op   0 allocs/op
BenchmarkIdleFrame-8   100   160000 ns/op   0 B/op   0 allocs/op
BenchmarkHeap-8         10     5000 ns/op   4096 B/op   2 allocs/op   2097152 heap-B/op
BenchmarkCustom-8       10     5000 ns/op   4096 B/op   2 allocs/op   1024 retained-B/op
`

	benchmarks, err := Parse(strings.NewReader(text), "inline")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(benchmarks) != 3 {
		t.Fatalf("Parse() returned %d benchmarks, want 3", len(benchmarks))
	}

	idle := benchmarks[0]
	if idle.Name != "BenchmarkIdleFrame-8" {
		t.Fatalf("first benchmark = %q, want BenchmarkIdleFrame-8", idle.Name)
	}
	if got := idle.Values[MetricTime]; got != 130000 {
		t.Fatalf("idle median ns/op = %v, want 130000", got)
	}
	if got := idle.Values[MetricBytes]; got != 0 {
		t.Fatalf("idle median B/op = %v, want 0", got)
	}

	heap := benchmarks[1]
	if got := heap.Values["heap-B/op"]; got != 2097152 {
		t.Fatalf("heap retained = %v, want 2097152", got)
	}
	custom := benchmarks[2]
	if got := custom.Values["retained-B/op"]; got != 1024 {
		t.Fatalf("custom retained = %v, want 1024", got)
	}
}

func TestParseSyntaxError(t *testing.T) {
	_, err := Parse(strings.NewReader("BenchmarkBroken 10 5\n"), "bad.txt")
	if err == nil {
		t.Fatal("Parse() error = nil, want a syntax error")
	}
	if !strings.Contains(err.Error(), "bad.txt") {
		t.Fatalf("Parse() error = %q, want the source name", err)
	}
}

func TestParseFileFixture(t *testing.T) {
	benchmarks, err := ParseFile(filepath.Join("testdata", "reference.txt"))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if len(benchmarks) != 6 {
		t.Fatalf("ParseFile() returned %d benchmarks, want 6", len(benchmarks))
	}
	if got := benchmarks[0].Values[MetricTime]; got != 105000 {
		t.Fatalf("idle fixture median = %v, want 105000", got)
	}
	class, sameData := Classify(benchmarks[0].Name, Manifest{"BenchmarkIdleFrame": "interaction"})
	if class != ClassInteraction || sameData {
		t.Fatalf("fixture manifest override = (%s, %v), want (Interaction, false)", class, sameData)
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
		want     Class
		sameData bool
	}{
		{name: "BenchmarkIdleFrame-8", want: ClassIdle},
		{name: "BenchmarkSameDataFrame-8", want: ClassIdle, sameData: true},
		{name: "BenchmarkClickOpen-8", want: ClassInteraction},
		{name: "BenchmarkPointerInteraction-8", want: ClassInteraction},
		{name: "BenchmarkInvoiceColdBuild-8", want: ClassRebuild},
		{name: "BenchmarkRebuildTree-8", want: ClassRebuild},
		{name: "BenchmarkResizeWindow-8", want: ClassResize},
		{name: "BenchmarkTypingRune-8", want: ClassTyping},
		{name: "BenchmarkKeystrokeLatency-8", want: ClassTyping},
		{name: "BenchmarkHeapRetained-8", want: ClassHeap},
		{name: "BenchmarkAllocsPerFrame-8", want: ClassHeap},
		{name: "BenchmarkUnmatchedRow-8", want: ClassOther},
		{
			name:     "BenchmarkIdlePreview-8",
			manifest: Manifest{"BenchmarkIdlePreview": "interaction"},
			want:     ClassInteraction,
		},
		{
			name:     "BenchmarkInvoiceColdBuild/sub-8",
			manifest: Manifest{"BenchmarkInvoiceColdBuild": "resize"},
			want:     ClassResize,
		},
		{
			name:     "BenchmarkPress-8",
			manifest: Manifest{"BenchmarkPress": "same-data"},
			want:     ClassIdle,
			sameData: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			class, sameData := Classify(test.name, test.manifest)
			if class != test.want || sameData != test.sameData {
				t.Fatalf("Classify(%q) = (%s, %v), want (%s, %v)", test.name, class, sameData, test.want, test.sameData)
			}
		})
	}
}

func TestParseManifest(t *testing.T) {
	const text = `{
		"BenchmarkInvoiceColdBuild": "rebuild",
		"BenchmarkClickOpen": "click",
		"BenchmarkSameData": "same-data"
	}`
	manifest, err := ParseManifest(strings.NewReader(text))
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}
	if len(manifest) != 3 {
		t.Fatalf("ParseManifest() returned %d entries, want 3", len(manifest))
	}
	if manifest["BenchmarkClickOpen"] != "click" {
		t.Fatalf("click entry = %q, want click", manifest["BenchmarkClickOpen"])
	}
	if class, sameData := Classify("BenchmarkSameData", manifest); class != ClassIdle || !sameData {
		t.Fatalf("same-data manifest = (%s, %v), want (Idle, true)", class, sameData)
	}

	if _, err := ParseManifest(strings.NewReader(`{"BenchmarkX": "fast"}`)); err == nil {
		t.Fatal("ParseManifest() error = nil, want an unknown-class error")
	}
	if _, err := ParseManifest(strings.NewReader(`{" ": "idle"}`)); err == nil {
		t.Fatal("ParseManifest() error = nil, want an empty-name error")
	}
	if _, err := ParseManifest(strings.NewReader(`not json`)); err == nil {
		t.Fatal("ParseManifest() error = nil, want a decode error")
	}
}

func TestLoadManifest(t *testing.T) {
	manifest, err := LoadManifest(filepath.Join("testdata", "classes.json"))
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if manifest["BenchmarkInvoiceColdBuild"] != "rebuild" {
		t.Fatalf("invoice class = %q, want rebuild", manifest["BenchmarkInvoiceColdBuild"])
	}
	empty, err := LoadManifest("")
	if err != nil {
		t.Fatalf("LoadManifest(empty) error = %v", err)
	}
	if empty != nil {
		t.Fatalf("LoadManifest(empty) = %v, want nil", empty)
	}
}

func benchmark(name string, values map[string]float64) Benchmark {
	return Benchmark{Name: name, Values: values}
}

func TestCompareBudgets(t *testing.T) {
	tests := []struct {
		name      string
		reference []Benchmark
		candidate []Benchmark
		manifest  Manifest
		wantPass  bool
		wantFails int
	}{
		{
			name:      "generic within budget",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100, MetricBytes: 100, MetricAllocs: 2})},
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 150, MetricBytes: 150, MetricAllocs: 3})},
			wantPass:  true,
		},
		{
			name:      "generic over budget",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 250})},
			wantFails: 1,
		},
		{
			name:      "generic at budget boundary",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 200})},
			wantPass:  true,
		},
		{
			name:      "idle over strict budget",
			reference: []Benchmark{benchmark("BenchmarkIdleFrame-8", map[string]float64{MetricTime: 100})},
			candidate: []Benchmark{benchmark("BenchmarkIdleFrame-8", map[string]float64{MetricTime: 110})},
			wantFails: 1,
		},
		{
			name:      "idle exact budget with no allocs",
			reference: []Benchmark{benchmark("BenchmarkIdleFrame-8", map[string]float64{MetricTime: 100, MetricAllocs: 0})},
			candidate: []Benchmark{benchmark("BenchmarkIdleFrame-8", map[string]float64{MetricTime: 100, MetricAllocs: 0})},
			wantPass:  true,
		},
		{
			name:      "idle allocates",
			reference: []Benchmark{benchmark("BenchmarkIdleFrame-8", map[string]float64{MetricTime: 100, MetricAllocs: 0})},
			candidate: []Benchmark{benchmark("BenchmarkIdleFrame-8", map[string]float64{MetricTime: 90, MetricAllocs: 1})},
			wantFails: 1,
		},
		{
			name:      "same-data strict budget",
			reference: []Benchmark{benchmark("BenchmarkSameDataFrame-8", map[string]float64{MetricTime: 100})},
			candidate: []Benchmark{benchmark("BenchmarkSameDataFrame-8", map[string]float64{MetricTime: 101})},
			wantFails: 1,
		},
		{
			name:      "bytes over budget",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100, MetricBytes: 100, MetricAllocs: 1})},
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100, MetricBytes: 250, MetricAllocs: 1})},
			wantFails: 1,
		},
		{
			name:      "allocs over budget",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100, MetricAllocs: 2})},
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100, MetricAllocs: 5})},
			wantFails: 1,
		},
		{
			name:      "retained heap ratio over budget",
			reference: []Benchmark{benchmark("BenchmarkHeap-8", map[string]float64{MetricTime: 100, "heap-B/op": 1000})},
			candidate: []Benchmark{benchmark("BenchmarkHeap-8", map[string]float64{MetricTime: 100, "heap-B/op": 2100})},
			wantFails: 1,
		},
		{
			name:      "retained heap at growth limit",
			reference: []Benchmark{benchmark("BenchmarkHeap-8", map[string]float64{MetricTime: 100, "heap-B/op": 3 << 20})},
			candidate: []Benchmark{benchmark("BenchmarkHeap-8", map[string]float64{MetricTime: 100, "heap-B/op": 4 << 20})},
			wantPass:  true,
		},
		{
			name:      "retained heap over growth limit",
			reference: []Benchmark{benchmark("BenchmarkHeap-8", map[string]float64{MetricTime: 100, "heap-B/op": 3 << 20})},
			candidate: []Benchmark{benchmark("BenchmarkHeap-8", map[string]float64{MetricTime: 100, "heap-B/op": 4<<20 + 1})},
			wantFails: 1,
		},
		{
			name:      "zero baseline",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 0, MetricBytes: 0})},
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 1, MetricBytes: 0})},
			wantFails: 1,
		},
		{
			name:      "missing benchmark",
			reference: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
			candidate: nil,
			wantFails: 1,
		},
		{
			name:      "new candidate benchmark",
			reference: nil,
			candidate: []Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
			wantPass:  true,
		},
		{
			name:      "manifest lifts an idle name",
			reference: []Benchmark{benchmark("BenchmarkIdlePreview-8", map[string]float64{MetricTime: 100})},
			candidate: []Benchmark{benchmark("BenchmarkIdlePreview-8", map[string]float64{MetricTime: 150})},
			manifest:  Manifest{"BenchmarkIdlePreview": "interaction"},
			wantPass:  true,
		},
		{
			name:      "manifest applies same-data budget",
			reference: []Benchmark{benchmark("BenchmarkPress-8", map[string]float64{MetricTime: 100})},
			candidate: []Benchmark{benchmark("BenchmarkPress-8", map[string]float64{MetricTime: 101})},
			manifest:  Manifest{"BenchmarkPress": "same-data"},
			wantFails: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := Compare(test.reference, test.candidate, test.manifest)
			if got := report.Passed(); got != test.wantPass {
				t.Fatalf("Passed() = %v, want %v; rows = %+v", got, test.wantPass, report.Rows)
			}
			if got := len(report.FailedRows()); got != test.wantFails {
				t.Fatalf("FailedRows() = %d, want %d; rows = %+v", got, test.wantFails, report.Rows)
			}
		})
	}
}

func TestCompareMissingRows(t *testing.T) {
	report := Compare(
		[]Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100, MetricBytes: 10})},
		[]Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
		nil,
	)
	failed := report.FailedRows()
	if len(failed) != 1 {
		t.Fatalf("FailedRows() = %d, want 1", len(failed))
	}
	if failed[0].Reason != "metric missing from candidate" {
		t.Fatalf("failure reason = %q, want metric missing from candidate", failed[0].Reason)
	}

	report = Compare(
		[]Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
		nil,
		nil,
	)
	failed = report.FailedRows()
	if len(failed) != 1 || failed[0].Reason != "benchmark missing from candidate" {
		t.Fatalf("missing benchmark failure = %+v, want benchmark missing from candidate", failed)
	}
}

func TestCompareFixtureAndReport(t *testing.T) {
	reference, err := ParseFile(filepath.Join("testdata", "reference.txt"))
	if err != nil {
		t.Fatalf("ParseFile(reference) error = %v", err)
	}
	candidate, err := ParseFile(filepath.Join("testdata", "candidate.txt"))
	if err != nil {
		t.Fatalf("ParseFile(candidate) error = %v", err)
	}
	manifest, err := LoadManifest(filepath.Join("testdata", "classes.json"))
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}

	report := Compare(reference, candidate, manifest)
	if report.Passed() {
		t.Fatal("Passed() = true, want failures from the fixture")
	}
	failed := report.FailedRows()
	if len(failed) != 3 {
		t.Fatalf("FailedRows() = %d, want 3: %+v", len(failed), failed)
	}
	if failed[0].Benchmark != "BenchmarkResizeWindow-8" || failed[0].Metric != MetricTime {
		t.Fatalf("first failure = %+v, want ResizeWindow ns/op", failed[0])
	}
	if failed[1].Benchmark != "BenchmarkTypingRune-8" || failed[1].Metric != MetricAllocs {
		t.Fatalf("second failure = %+v, want TypingRune allocs/op", failed[1])
	}
	if failed[2].Benchmark != "BenchmarkHeapRetained-8" || !strings.Contains(failed[2].Reason, "growth") {
		t.Fatalf("third failure = %+v, want HeapRetained growth", failed[2])
	}

	var output bytes.Buffer
	if err := report.Write(&output); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	text := output.String()
	for _, want := range []string{
		"CLASS",
		"BENCHMARK",
		"REFERENCE",
		"CANDIDATE",
		"RATIO",
		"RESULT",
		"BenchmarkResizeWindow-8",
		"FAILED ROWS",
		"retained heap growth",
		"3 of 19 rows failed",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("report output missing %q:\n%s", want, text)
		}
	}
}

func TestReportPassSummary(t *testing.T) {
	report := Compare(
		[]Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
		[]Benchmark{benchmark("BenchmarkClick-8", map[string]float64{MetricTime: 100})},
		nil,
	)
	var output bytes.Buffer
	if err := report.Write(&output); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !strings.Contains(output.String(), "1 rows checked, all within budget") {
		t.Fatalf("report output = %q, want pass summary", output.String())
	}
}
