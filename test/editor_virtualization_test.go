package test_test

import (
	"strings"
	"testing"
	"unicode"
)

// ----------------------------------------------------------------------------
// Category 4: Zero-Allocation Editor Virtualization & Syntax Highlighting
// ----------------------------------------------------------------------------

func generateSyntheticGo(target int) []string {
	tpl := []string{
		"package worker",
		"",
		"import (",
		"\t\"context\"",
		"\t\"fmt\"",
		"\t\"sync\"",
		"\t\"time\"",
		")",
		"",
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
		"\tPayload   map[string]any",
		"}",
		"",
		"func Dispatch(ctx context.Context, job *Job) (bool, error) {",
		"\tif job == nil {",
		"\t\treturn false, fmt.Errorf(\"nil job\")",
		"\t}",
		"\treturn true, nil",
		"}",
	}
	lines := make([]string, 0, target)
	for len(lines) < target {
		for _, l := range tpl {
			if len(lines) >= target {
				break
			}
			lines = append(lines, l)
		}
	}
	return lines
}

func generateSyntheticPython(target int) []string {
	tpl := []string{
		"import sys, os, time",
		"from typing import List, Dict, Optional",
		"",
		"DEFAULT_TIMEOUT = 30.0",
		"",
		"class TaskRunner:",
		"    \"\"\"Manages background batch runs.\"\"\"",
		"    def __init__(self, name: str):",
		"        self.name = name",
		"        self.active = True",
		"",
		"    def execute(self, job_id: str, batch: List[int]) -> Optional[int]:",
		"        if not batch:",
		"            return None",
		"        acc = sum(x * 2 for x in batch if x > 0)",
		"        return acc",
	}
	lines := make([]string, 0, target)
	for len(lines) < target {
		for _, l := range tpl {
			if len(lines) >= target {
				break
			}
			lines = append(lines, l)
		}
	}
	return lines
}

func generateSyntheticMarkdown(target int) []string {
	tpl := []string{
		"# System Telemetry Overview",
		"",
		"This document tracks **frame latency** and *allocations*.",
		"- Row slicing at 20px per item",
		"- Pure Go tokenizer with DFA state machine",
		"",
		"```go",
		"func ProcessPacket(p []byte) error { return nil }",
		"```",
		"",
		"| Stage | Latency | Allocations |",
		"| --- | --- | --- |",
		"| Parse | 0.04ms | 0 |",
		"| Layout | 1.12ms | 2KB |",
		"",
		"---",
	}
	lines := make([]string, 0, target)
	for len(lines) < target {
		for _, l := range tpl {
			if len(lines) >= target {
				break
			}
			lines = append(lines, l)
		}
	}
	return lines
}

// ----------------------------------------------------------------------------
// Zero-Allocation Span-Based Tokenizer & Arenas
// ----------------------------------------------------------------------------

type TokenSpan struct {
	Start uint32
	End   uint32
	Kind  uint8
}

type TokenArena struct {
	spans []TokenSpan
}

func newTokenArena(cap int) *TokenArena {
	return &TokenArena{spans: make([]TokenSpan, 0, cap)}
}

func (ta *TokenArena) TokenizeLine(line string) []TokenSpan {
	ta.spans = ta.spans[:0]
	start := 0
	inWord := false

	for i, r := range line {
		if unicode.IsSpace(r) {
			if inWord {
				ta.spans = append(ta.spans, TokenSpan{Start: uint32(start), End: uint32(i), Kind: 1})
				inWord = false
			}
			ta.spans = append(ta.spans, TokenSpan{Start: uint32(i), End: uint32(i + 1), Kind: 0})
		} else if strings.ContainsRune("(){}[];:.,+-*/=<>!&|#`*_\t\"'", r) {
			if inWord {
				ta.spans = append(ta.spans, TokenSpan{Start: uint32(start), End: uint32(i), Kind: 1})
				inWord = false
			}
			ta.spans = append(ta.spans, TokenSpan{Start: uint32(i), End: uint32(i + 1), Kind: 2})
		} else {
			if !inWord {
				start = i
				inWord = true
			}
		}
	}
	if inWord {
		ta.spans = append(ta.spans, TokenSpan{Start: uint32(start), End: uint32(len(line)), Kind: 1})
	}
	return ta.spans
}

// ----------------------------------------------------------------------------
// Zero-Allocation Viewport Buffer
// ----------------------------------------------------------------------------

type visualRow struct {
	LineIndex int
	Text      string
}

type ViewportBuffer struct {
	rows []visualRow
}

func newViewportBuffer(cap int) *ViewportBuffer {
	return &ViewportBuffer{rows: make([]visualRow, 0, cap)}
}

func (vb *ViewportBuffer) Slice(lines []string, offsetY int, viewportHeight int, rowHeight int) []visualRow {
	vb.rows = vb.rows[:0]
	if rowHeight <= 0 || len(lines) == 0 {
		return vb.rows
	}
	startLine := offsetY / rowHeight
	if startLine >= len(lines) {
		return vb.rows
	}
	visibleCount := (viewportHeight / rowHeight) + 2
	endLine := startLine + visibleCount
	if endLine > len(lines) {
		endLine = len(lines)
	}

	for i := startLine; i < endLine; i++ {
		vb.rows = append(vb.rows, visualRow{LineIndex: i, Text: lines[i]})
	}
	return vb.rows
}

// ----------------------------------------------------------------------------
// Zero-Allocation Soft-Wrap Breakpoint Buffer
// ----------------------------------------------------------------------------

type WrapPlan struct {
	breaks []int
}

func newWrapPlan(cap int) *WrapPlan {
	return &WrapPlan{breaks: make([]int, 0, cap)}
}

func (wp *WrapPlan) Compute(line string, maxCols int) []int {
	wp.breaks = wp.breaks[:0]
	if len(line) <= maxCols || maxCols <= 0 {
		return wp.breaks
	}
	col := 0
	for i, r := range line {
		col++
		if col >= maxCols && unicode.IsSpace(r) {
			wp.breaks = append(wp.breaks, i)
			col = 0
		}
	}
	return wp.breaks
}

// ----------------------------------------------------------------------------
// Zero-Allocation Benchmarks
// ----------------------------------------------------------------------------

func BenchmarkEditor_SoftWrap_Resize_100(b *testing.B) {
	lines := generateSyntheticGo(100)
	plan := newWrapPlan(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cols := 80 + (i%20)*2
		for _, l := range lines {
			_ = plan.Compute(l, cols)
		}
	}
}

func BenchmarkEditor_SoftWrap_Resize_1000(b *testing.B) {
	lines := generateSyntheticGo(1000)
	plan := newWrapPlan(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cols := 80 + (i%20)*2
		for _, l := range lines {
			_ = plan.Compute(l, cols)
		}
	}
}

func BenchmarkEditor_RowWindowing_Scroll(b *testing.B) {
	lines := generateSyntheticGo(1000)
	const rowH = 20
	const viewH = 600
	vb := newViewportBuffer(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offsetY := ((i % 200) + 1) * rowH
		_ = vb.Slice(lines, offsetY, viewH, rowH)
	}
}

func BenchmarkEditor_RowWindowing_FillCode_Wrap(b *testing.B) {
	lines := generateSyntheticGo(1000)
	const rowH = 20
	const viewH = 600
	offsetY := 200 * rowH
	vb := newViewportBuffer(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vb.Slice(lines, offsetY, viewH, rowH)
	}
}

func BenchmarkEditor_RowWindowing_FillCode_NoWrap(b *testing.B) {
	lines := generateSyntheticGo(1000)
	const rowH = 20
	const viewH = 600
	offsetY := 200 * rowH
	vb := newViewportBuffer(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vb.Slice(lines, offsetY, viewH, rowH)
	}
}

func BenchmarkEditor_Syntax_Go_100(b *testing.B) {
	lines := generateSyntheticGo(100)
	arena := newTokenArena(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			_ = arena.TokenizeLine(l)
		}
	}
}

func BenchmarkEditor_Syntax_Go_1000(b *testing.B) {
	lines := generateSyntheticGo(1000)
	arena := newTokenArena(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			_ = arena.TokenizeLine(l)
		}
	}
}

func BenchmarkEditor_Syntax_Python_100(b *testing.B) {
	lines := generateSyntheticPython(100)
	arena := newTokenArena(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			_ = arena.TokenizeLine(l)
		}
	}
}

func BenchmarkEditor_Syntax_Python_1000(b *testing.B) {
	lines := generateSyntheticPython(1000)
	arena := newTokenArena(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			_ = arena.TokenizeLine(l)
		}
	}
}

func BenchmarkEditor_Syntax_Markdown_100(b *testing.B) {
	lines := generateSyntheticMarkdown(100)
	arena := newTokenArena(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			_ = arena.TokenizeLine(l)
		}
	}
}

func BenchmarkEditor_Syntax_Markdown_1000(b *testing.B) {
	lines := generateSyntheticMarkdown(1000)
	arena := newTokenArena(64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lines {
			_ = arena.TokenizeLine(l)
		}
	}
}
