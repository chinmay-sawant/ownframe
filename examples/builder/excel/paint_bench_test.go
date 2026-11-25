package main

import (
	"testing"
)

func benchApp(b *testing.B, full bool) *app {
	b.Helper()
	if full {
		return newFullApp(b)
	}
	a, err := newApp()
	if err != nil {
		b.Fatal(err)
	}
	a.redraw()
	return a
}

// BenchmarkExcelFirstPaintFull is the first mount and paint of a fully
// populated 1000 x 52 sheet (EXL-80 case 1).
func BenchmarkExcelFirstPaintFull(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		a, err := newApp()
		if err != nil {
			b.Fatal(err)
		}
		fillFullSheet(a.sheet())
		b.StartTimer()
		a.redraw()
		if a.lastErr != nil {
			b.Fatal(a.lastErr)
		}
	}
}

// BenchmarkExcelScrollStep scrolls one row down and back, redrawing each
// time (EXL-80 case 3).
func BenchmarkExcelScrollStep(b *testing.B) {
	a := benchApp(b, true)
	b.ReportAllocs()
	b.ResetTimer()
	d := 1
	for b.Loop() {
		if !a.scrollBy(d, 0) {
			d = -d
			a.scrollBy(d, 0)
		}
		a.redraw()
	}
}

// BenchmarkExcelSelect moves the active cell one step and repaints.
func BenchmarkExcelSelect(b *testing.B) {
	a := benchApp(b, true)
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		a.selectCell(CellRef{Row: 2 + i%10, Col: 2 + (i/10)%6})
		i++
	}
}

// BenchmarkExcelEditCommitRecalc edits the head of a 999-cell dependency
// chain and redraws (EXL-80 case 2).
func BenchmarkExcelEditCommitRecalc(b *testing.B) {
	a := benchApp(b, true)
	a.selectCell(CellRef{0, 0})
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		a.beginEdit("")
		a.editBuf = string(rune('1' + i%9))
		a.commitEdit()
		i++
	}
}

// BenchmarkExcelThemeToggle flips Light/Dark and repaints (EXL-80 case 4).
func BenchmarkExcelThemeToggle(b *testing.B) {
	a := benchApp(b, true)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		a.toggleTheme()
	}
}

// BenchmarkExcelRibbonTabSwitch cycles the ribbon through two tabs.
func BenchmarkExcelRibbonTabSwitch(b *testing.B) {
	a := benchApp(b, false)
	tabs := []string{"xl:tab:Insert", "xl:tab:Home"}
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		a.runAction(tabs[i%2])
		i++
	}
}

// BenchmarkExcelChartDrag switches to the Charts sheet (6 seeded charts) and
// drags a chart by small steps, repainting each frame like live DragMove.
func BenchmarkExcelChartDrag(b *testing.B) {
	a := benchApp(b, false)
	for i, s := range a.wb.Sheets {
		if s.Name == "Charts" {
			a.wb.Active = i
			break
		}
	}
	a.redraw()
	var target *ChartSpec
	for _, c := range a.x.charts {
		if c.Sheet == a.wb.Sheet().Name {
			target = c
			break
		}
	}
	if target == nil {
		target = a.addChart("column")
		a.redraw()
	}
	a.x.selChart, a.x.dragChart = target.ID, target.ID
	a.redraw()
	b.ReportAllocs()
	b.ResetTimer()
	dx := 4.0
	for i := 0; b.Loop(); i++ {
		if i%20 == 19 {
			dx = -dx
		}
		a.chartDragMove(dx, 2.0)
		a.redraw()
		if a.lastErr != nil {
			b.Fatal(a.lastErr)
		}
	}
	a.chartDragEnd()
}
