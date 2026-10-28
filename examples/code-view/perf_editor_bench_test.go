package main

import (
	"context"
	"testing"
)

// generateGoLines creates synthetic but realistic Go code for benchmark testing.
func generateGoLines(target int) []string {
	template := []string{
		"package worker",
		"",
		"import (",
		"\t\"context\"",
		"\t\"fmt\"",
		"\t\"sync\"",
		"\t\"time\"",
		")",
		"",
		"// JobState indicates status of a worker job.",
		"type JobState int",
		"",
		"const (",
		"\tStatePending JobState = iota",
		"\tStateRunning",
		"\tStateCompleted",
		"\tStateFailed",
		")",
		"",
		"type Job struct {",
		"\tID        string",
		"\tPriority  int",
		"\tState     JobState",
		"\tRetries   int",
		"\tCreatedAt time.Time",
		"\tPayload   map[string]any",
		"}",
		"",
		"type Pool struct {",
		"\tmu       sync.Mutex",
		"\tactive   []*Job",
		"\tcapacity int",
		"\tclosed   bool",
		"}",
		"",
		"func (p *Pool) Dispatch(ctx context.Context, job *Job) (bool, error) {",
		"\tp.mu.Lock()",
		"\tdefer p.mu.Unlock()",
		"\tif p.closed {",
		"\t\treturn false, fmt.Errorf(\"pool is closed\")",
		"\t}",
		"\tif len(p.active) >= p.capacity {",
		"\t\treturn false, nil",
		"\t}",
		"\tjob.State = StateRunning",
		"\tp.active = append(p.active, job)",
		"\treturn true, nil",
		"}",
	}

	lines := make([]string, 0, target)
	for len(lines) < target {
		for _, line := range template {
			if len(lines) >= target {
				break
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// generatePythonLines creates synthetic Python code for benchmark testing.
func generatePythonLines(target int) []string {
	template := []string{
		"import os",
		"import sys",
		"import time",
		"from typing import List, Dict, Optional, Tuple",
		"",
		"# Module level constant configuration",
		"DEFAULT_TIMEOUT: float = 30.0",
		"MAX_RETRIES: int = 5",
		"",
		"class TaskRunner:",
		"    \"\"\"TaskRunner manages asynchronous worker jobs.",
		"    It handles retries, state machine transitions, and backoff.",
		"    \"\"\"",
		"",
		"    def __init__(self, name: str, retries: int = MAX_RETRIES):",
		"        self.name: str = name",
		"        self.retries: int = retries",
		"        self.active: bool = True",
		"        self.history: List[Dict[str, int]] = []",
		"",
		"    @property",
		"    def is_available(self) -> bool:",
		"        return self.active and self.retries > 0",
		"",
		"    def execute(self, job_id: str, batch: List[int]) -> Optional[int]:",
		"        # Calculate checksum for the current batch",
		"        if not batch:",
		"            return None",
		"        accumulator = 0",
		"        for idx, val in enumerate(batch):",
		"            if val < 0:",
		"                continue",
		"            accumulator += (val * 2) ^ (idx + 1)",
		"        record = {'id': job_id, 'sum': accumulator}",
		"        self.history.append(record)",
		"        return accumulator",
	}

	lines := make([]string, 0, target)
	for len(lines) < target {
		for _, line := range template {
			if len(lines) >= target {
				break
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// generateMarkdownLines creates synthetic Markdown content for benchmark testing.
func generateMarkdownLines(target int) []string {
	template := []string{
		"# Systems Architecture Documentation",
		"",
		"## Overview and Key Constraints",
		"",
		"This service handles **high-throughput** streaming data with *low latency* bounds.",
		"For reference, consult the [API Specification](https://example.org/spec) and internal docs.",
		"",
		"### Operational Characteristics",
		"",
		"- Monolithic memory arena with 64MB chunks",
		"- Zero-allocation ring buffer for socket reads",
		"- Deterministic garbage collector pacing",
		"- Visual row tracking at 20px per item",
		"",
		"> Warning: Avoid holding long locks in the critical section.",
		"",
		"```go",
		"func ProcessStream(ctx context.Context, ch <-chan Packet) error {",
		"\tfor pkt := range ch {",
		"\t\tif err := pkt.Validate(); err != nil {",
		"\t\t\treturn err",
		"\t\t}",
		"\t}",
		"\treturn nil",
		"}",
		"```",
		"",
		"| Component | P99 Latency | Memory Overhead |",
		"| --- | --- | --- |",
		"| Parser | 0.42ms | 12KB |",
		"| Router | 0.15ms | 4KB |",
		"| Shaper | 0.88ms | 32KB |",
		"",
		"---",
	}

	lines := make([]string, 0, target)
	for len(lines) < target {
		for _, line := range template {
			if len(lines) >= target {
				break
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// ----------------------------------------------------------------------------
// 1. Syntax Highlighting Benchmarks
// ----------------------------------------------------------------------------

func BenchmarkEditor_Syntax_Go_100(b *testing.B) {
	lines := generateGoLines(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(langGo, lines)
	}
}

func BenchmarkEditor_Syntax_Go_1000(b *testing.B) {
	lines := generateGoLines(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(langGo, lines)
	}
}

func BenchmarkEditor_Syntax_Python_100(b *testing.B) {
	lines := generatePythonLines(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(langPython, lines)
	}
}

func BenchmarkEditor_Syntax_Python_1000(b *testing.B) {
	lines := generatePythonLines(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(langPython, lines)
	}
}

func BenchmarkEditor_Syntax_Markdown_100(b *testing.B) {
	lines := generateMarkdownLines(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(langMarkdown, lines)
	}
}

func BenchmarkEditor_Syntax_Markdown_1000(b *testing.B) {
	lines := generateMarkdownLines(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(langMarkdown, lines)
	}
}

// ----------------------------------------------------------------------------
// 2. Soft-Wrap Engine Plan Generation Benchmarks
// ----------------------------------------------------------------------------

func BenchmarkEditor_SoftWrap_Resize_100(b *testing.B) {
	lines := generateGoLines(100)
	doc := &fileDoc{
		Path:   "worker.go",
		Lang:   langGo,
		Lines:  lines,
		Tokens: tokenize(langGo, lines),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cols := 80 + (i%20)*2
		_ = doc.wrapped(cols)
	}
}

func BenchmarkEditor_SoftWrap_Resize_1000(b *testing.B) {
	lines := generateGoLines(1000)
	doc := &fileDoc{
		Path:   "worker.go",
		Lang:   langGo,
		Lines:  lines,
		Tokens: tokenize(langGo, lines),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cols := 80 + (i%20)*2
		_ = doc.wrapped(cols)
	}
}

// ----------------------------------------------------------------------------
// 3. Editor Row Windowing During Scrolling Benchmarks
// ----------------------------------------------------------------------------

func setupBenchApp(b *testing.B, lineCount int, wrap bool) (*App, *fileDoc) {
	b.Helper()
	app, err := New("")
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	lines := generateGoLines(lineCount)
	doc := &fileDoc{
		Path:     "backend/worker.go",
		Lang:     langGo,
		Lines:    lines,
		Tokens:   tokenize(langGo, lines),
		ByteSize: int64(len(lines) * 35),
	}
	app.docs["backend/worker.go"] = doc
	app.tabs = []string{"backend/worker.go"}
	app.active = 0
	app.wrap = wrap
	app.refresh()
	if err := app.Page().Redraw(context.Background()); err != nil {
		b.Fatalf("redraw: %v", err)
	}
	return app, doc
}

func BenchmarkEditor_RowWindowing_Scroll(b *testing.B) {
	app, _ := setupBenchApp(b, 1000, true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offsetY := ((i % 200) + 1) * rowH
		app.onScroll(offsetY, defaultHeight)
	}
}

func BenchmarkEditor_RowWindowing_Scroll_NoWrap(b *testing.B) {
	app, _ := setupBenchApp(b, 1000, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offsetY := ((i % 200) + 1) * rowH
		app.onScroll(offsetY, defaultHeight)
	}
}

func BenchmarkEditor_RowWindowing_FillCode_Wrap(b *testing.B) {
	app, _ := setupBenchApp(b, 1000, true)
	v := app.buildView()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.scrollY = ((i % 200) + 1) * rowH
		app.fillCode(&v)
	}
}

func BenchmarkEditor_RowWindowing_FillCode_NoWrap(b *testing.B) {
	app, _ := setupBenchApp(b, 1000, false)
	v := app.buildView()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.scrollY = ((i % 200) + 1) * rowH
		app.fillCode(&v)
	}
}
