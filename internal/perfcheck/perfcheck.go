// Package perfcheck checks page redraw stages against budgets that scale with
// document size, so one set of limits applies to any page.
//
// The caller reads the page counters with Perf enabled and passes plain values:
//
//	stats := p.Probe().Stats()
//	perfcheck.Check(t, perfcheck.Stats{
//		Template: stats.LastTemplate,
//		Style:    stats.StyleTime,
//		Layout:   stats.LayoutTime,
//		Paint:    stats.PaintTime,
//		Other:    stats.OtherTime,
//		Alloc:    stats.AllocFrame,
//	}, stats.Boxes, stats.Ops)
//
// A zero box or op count skips the rows that need that denominator. The
// default budgets target a warm redraw; pass a Budgets value for a cold render.
package perfcheck

import (
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
)

const (
	defaultTemplatePerBox = 1.0
	defaultTemplatePerOp  = 0.25
	defaultStylePerBox    = 2.0
	defaultStylePerOp     = 0.5
	defaultLayoutPerBox   = 3.0
	defaultLayoutPerOp    = 0.75
	defaultPaintPerBox    = 3.0
	defaultPaintPerOp     = 0.75
	defaultOtherPerBox    = 1.0
	defaultOtherPerOp     = 0.25
	defaultAllocPerBox    = 128 * 1024
	tablePadding          = 2
)

// Stats is the pipeline cost of one redraw. The fields hold the same values as
// the page counters: LastTemplate, StyleTime, LayoutTime, PaintTime,
// OtherTime, and AllocFrame.
type Stats struct {
	Template time.Duration
	Style    time.Duration
	Layout   time.Duration
	Paint    time.Duration
	Other    time.Duration
	Alloc    uint64
}

// StageBudget is one stage's ceiling in microseconds per box and per op. A zero
// or negative ceiling disables that check.
type StageBudget struct {
	PerBox float64
	PerOp  float64
}

// Budgets is the set of ceilings Check and Report apply. Stage values are
// microseconds; AllocPerBox is bytes. A zero or negative ceiling disables that
// check.
type Budgets struct {
	Template    StageBudget
	Style       StageBudget
	Layout      StageBudget
	Paint       StageBudget
	Other       StageBudget
	AllocPerBox float64
}

// DefaultBudgets returns the ceilings Check and Report use. Stage values are
// microseconds; the allocation value is bytes.
//
// The timing ceilings target a warm redraw: 10 µs per box and 2.5 µs per op
// across all five stages. The allocation ceiling is loose because AllocFrame
// samples about twice a second, so a fast test can still carry the first cold
// redraw in its only sample. Pass a Budgets value for a cold render or a host
// with a different cost class.
func DefaultBudgets() Budgets {
	return Budgets{
		Template: StageBudget{
			PerBox: defaultTemplatePerBox,
			PerOp:  defaultTemplatePerOp,
		},
		Style: StageBudget{
			PerBox: defaultStylePerBox,
			PerOp:  defaultStylePerOp,
		},
		Layout: StageBudget{
			PerBox: defaultLayoutPerBox,
			PerOp:  defaultLayoutPerOp,
		},
		Paint: StageBudget{
			PerBox: defaultPaintPerBox,
			PerOp:  defaultPaintPerOp,
		},
		Other: StageBudget{
			PerBox: defaultOtherPerBox,
			PerOp:  defaultOtherPerOp,
		},
		AllocPerBox: defaultAllocPerBox,
	}
}

// Violation is one metric over its ceiling. Actual and Budget carry the row's
// unit: microseconds for a stage row and bytes for the allocation row.
type Violation struct {
	Stage  string
	Unit   string
	Actual float64
	Budget float64
}

// Violations returns every metric over budget. A zero box or op count skips
// the rows that need that denominator, and a zero or negative ceiling skips
// its row.
func Violations(budgets Budgets, stats Stats, boxes, ops int) []Violation {
	rows := measurements(budgets, stats, boxes, ops)
	violations := make([]Violation, 0, len(rows))
	for _, row := range rows {
		if !row.over {
			continue
		}
		violations = append(violations, Violation{
			Stage:  row.stage,
			Unit:   row.unit,
			Actual: row.actual,
			Budget: row.budget,
		})
	}
	return violations
}

// Check fails t with a table when any default budget is exceeded.
func Check(t *testing.T, stats Stats, boxes, ops int) {
	t.Helper()
	check(t, DefaultBudgets(), stats, boxes, ops)
}

// CheckWith fails t with a table when any budget in budgets is exceeded.
func CheckWith(t *testing.T, budgets Budgets, stats Stats, boxes, ops int) {
	t.Helper()
	check(t, budgets, stats, boxes, ops)
}

// Report formats the default-budget table whether or not a budget is exceeded.
func Report(stats Stats, boxes, ops int) string {
	return ReportWith(DefaultBudgets(), stats, boxes, ops)
}

// ReportWith formats the budget table whether or not a budget is exceeded.
func ReportWith(budgets Budgets, stats Stats, boxes, ops int) string {
	rows := measurements(budgets, stats, boxes, ops)
	over := 0
	for _, row := range rows {
		if row.over {
			over++
		}
	}

	var out strings.Builder
	if over == 0 {
		_, _ = fmt.Fprintf(&out, "perfcheck: boxes=%d ops=%d status=ok\n", boxes, ops)
	} else {
		_, _ = fmt.Fprintf(
			&out,
			"perfcheck: boxes=%d ops=%d status=FAIL (%d over budget)\n",
			boxes,
			ops,
			over,
		)
	}

	table := tabwriter.NewWriter(&out, 0, 0, tablePadding, ' ', 0)
	_, _ = fmt.Fprintln(table, "stage\tunit\tactual\tbudget\tstatus")
	for _, row := range rows {
		_, _ = fmt.Fprintf(
			table,
			"%s\t%s\t%s\t%s\t%s\n",
			row.stage,
			row.unit,
			row.actualText(),
			row.budgetText(),
			row.statusText(),
		)
	}
	_ = table.Flush()

	return out.String()
}

// testReporter is the part of *testing.T Check needs, so tests can drive the
// failure path without failing themselves.
type testReporter interface {
	Helper()
	Errorf(format string, args ...any)
}

func check(t testReporter, budgets Budgets, stats Stats, boxes, ops int) {
	t.Helper()
	if len(Violations(budgets, stats, boxes, ops)) == 0 {
		return
	}
	t.Errorf("%s", ReportWith(budgets, stats, boxes, ops))
}

type stageCost struct {
	name   string
	total  time.Duration
	budget StageBudget
}

func stageCosts(stats Stats, budgets Budgets) []stageCost {
	return []stageCost{
		{
			name:   "template",
			total:  stats.Template,
			budget: budgets.Template,
		},
		{
			name:   "style",
			total:  stats.Style,
			budget: budgets.Style,
		},
		{
			name:   "layout",
			total:  stats.Layout,
			budget: budgets.Layout,
		},
		{
			name:   "paint",
			total:  stats.Paint,
			budget: budgets.Paint,
		},
		{
			name:   "other",
			total:  stats.Other,
			budget: budgets.Other,
		},
	}
}

type measurementRow struct {
	stage  string
	unit   string
	actual float64
	budget float64
	bytes  bool
	skip   bool
	off    bool
	over   bool
}

func measurements(budgets Budgets, stats Stats, boxes, ops int) []measurementRow {
	stages := stageCosts(stats, budgets)
	rows := make([]measurementRow, 0, len(stages)*2+1)
	for _, stage := range stages {
		rows = append(rows,
			stageRow(stage, "box", boxes),
			stageRow(stage, "op", ops),
		)
	}
	rows = append(rows, allocRow(stats.Alloc, boxes, budgets.AllocPerBox))
	return rows
}

func stageRow(stage stageCost, unit string, count int) measurementRow {
	budget := stage.budget.PerOp
	if unit == "box" {
		budget = stage.budget.PerBox
	}
	return normalizeRow(
		measurementRow{
			stage:  stage.name,
			unit:   unit,
			budget: budget,
		},
		durationMicros(stage.total),
		count,
	)
}

func allocRow(alloc uint64, boxes int, budget float64) measurementRow {
	return normalizeRow(
		measurementRow{
			stage:  "alloc",
			unit:   "box",
			budget: budget,
			bytes:  true,
		},
		float64(alloc),
		boxes,
	)
}

func normalizeRow(row measurementRow, total float64, count int) measurementRow {
	if count <= 0 {
		row.skip = true
		return row
	}

	row.actual = total / float64(count)
	row.off = row.budget <= 0
	row.over = !row.off && row.actual > row.budget
	return row
}

func durationMicros(d time.Duration) float64 {
	return float64(d) / float64(time.Microsecond)
}

func (r measurementRow) actualText() string {
	if r.skip {
		return "-"
	}
	return r.valueText(r.actual)
}

func (r measurementRow) budgetText() string {
	if r.skip {
		return "-"
	}
	if r.off {
		return "off"
	}
	return r.valueText(r.budget)
}

func (r measurementRow) statusText() string {
	switch {
	case r.skip:
		if r.unit == "box" {
			return "skipped (no boxes)"
		}
		return "skipped (no ops)"
	case r.off:
		return "off"
	case r.over:
		return "over by " + r.valueText(r.actual-r.budget)
	default:
		return "ok"
	}
}

func (r measurementRow) valueText(value float64) string {
	if r.bytes {
		return fmt.Sprintf("%.0f B", value)
	}
	return fmt.Sprintf("%.2f µs", value)
}
