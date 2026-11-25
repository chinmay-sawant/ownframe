package main

import (
	"encoding/hex"
	"image"
	"image/color"
	"regexp"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// tokColor returns the #rrggbb value of a --xl-* token in the given theme.
func tokColor(t *testing.T, dark bool, name string) color.RGBA {
	t.Helper()
	src := xlLightTokens
	if dark {
		src = xlDarkTokens
	}
	m := regexp.MustCompile(`--` + regexp.QuoteMeta(name) + `:#([0-9a-fA-F]{6})`).FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("token --%s not found", name)
	}
	rgb, err := hex.DecodeString(m[1])
	if err != nil {
		t.Fatal(err)
	}
	return color.RGBA{rgb[0], rgb[1], rgb[2], 255}
}

func near(got color.Color, want color.RGBA, tol int) bool {
	r, g, b, _ := got.RGBA()
	d := func(a uint32, b uint8) int {
		x := int(a>>8) - int(b)
		if x < 0 {
			x = -x
		}
		return x
	}
	return d(r, want.R) <= tol && d(g, want.G) <= tol && d(b, want.B) <= tol
}

// countNear counts pixels of img inside rect within tol of want.
func countNear(img image.Image, rect image.Rectangle, want color.RGBA, tol int) int {
	n := 0
	rect = rect.Intersect(img.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if near(img.At(x, y), want, tol) {
				n++
			}
		}
	}
	return n
}

// cellRect is the pixel rect of a visible cell in window coordinates.
func cellRect(t *testing.T, a *app, c CellRef) image.Rectangle {
	t.Helper()
	cw, ch := a.canvasSize()
	col, okc := slotFor(a.layoutCols(cw), c.Col)
	row, okr := slotFor(a.layoutRows(ch), c.Row)
	if !okc || !okr {
		t.Fatalf("cell %s is not visible", c.A1())
	}
	gx, gy := a.gridOrigin()
	x := int(gx) + a.rowHdrW() + col.pos
	y := int(gy) + a.colHdrH() + row.pos
	return image.Rect(x, y, x+col.size, y+row.size)
}

func setTheme(t *testing.T, a *app, dark bool) {
	t.Helper()
	if a.dark != dark {
		a.toggleTheme()
	}
	if a.lastErr != nil {
		t.Fatal(a.lastErr)
	}
}

// TestThemePixels pins EXL-81: the painted frame carries the --xl-* token
// colors of each theme for the title bar, grid canvas, headers, gridlines
// and the active-cell selection border.
func TestThemePixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		name := "light"
		if dark {
			name = "dark"
		}
		t.Run(name, func(t *testing.T) {
			a, _ := newTestApp(t)
			setTheme(t, a, dark)
			a.selectCell(CellRef{9, 3}) // D10: empty, away from the sample data
			img := framePNG(t, a)

			title := tokColor(t, dark, "xl-titlebar-bg")
			if got := img.At(windowWidth/2, 4); !near(got, title, 2) {
				t.Errorf("title bar pixel = %v, want %v", got, title)
			}
			sheetBg := tokColor(t, dark, "xl-sheet-bg")
			empty := cellRect(t, a, CellRef{20, 8}).Inset(3)
			if n := countNear(img, empty, sheetBg, 2); n < empty.Dx()*empty.Dy()*9/10 {
				t.Errorf("empty cell paints %d/%d sheet-bg pixels", n, empty.Dx()*empty.Dy())
			}
			gx, gy := a.gridOrigin()
			hdr := image.Rect(int(gx)+a.rowHdrW()+4, int(gy), int(gx)+a.rowHdrW()+30, int(gy)+a.colHdrH())
			if n := countNear(img, hdr, tokColor(t, dark, "xl-header-bg"), 2); n == 0 {
				t.Errorf("column header paints no xl-header-bg pixels")
			}
			// The active column header switches to the active tint.
			act := cellRect(t, a, CellRef{9, 3})
			actHdr := image.Rect(act.Min.X, int(gy), act.Max.X, int(gy)+a.colHdrH())
			if n := countNear(img, actHdr, tokColor(t, dark, "xl-header-active-bg"), 2); n == 0 {
				t.Errorf("active column header lacks xl-header-active-bg")
			}
			grid := cellRect(t, a, CellRef{20, 8})
			if n := countNear(img, image.Rect(grid.Min.X-2, grid.Min.Y-2, grid.Max.X+2, grid.Min.Y+2), tokColor(t, dark, "xl-gridline"), 2); n == 0 {
				t.Errorf("no gridline pixels around an empty cell")
			}
			sel := tokColor(t, dark, "xl-selection-border")
			edge := image.Rect(act.Min.X-1, act.Min.Y+4, act.Min.X+3, act.Max.Y-4)
			if n := countNear(img, edge, sel, 2); n == 0 {
				t.Errorf("active cell border lacks xl-selection-border %v", sel)
			}
		})
	}
}

// TestThemePixelsRibbonTab checks the active ribbon tab paints differently
// from an inactive one, and that both differ across themes.
func TestThemePixelsRibbonTab(t *testing.T) {
	var prev color.Color
	for _, dark := range []bool{false, true} {
		a, _ := newTestApp(t)
		setTheme(t, a, dark)
		img := framePNG(t, a)
		var home, insert ownframe.Box
		for _, b := range a.page.Boxes() {
			switch b.Action {
			case "xl:tab:Home":
				home = b
			case "xl:tab:Insert":
				insert = b
			}
		}
		if home.W == 0 || insert.W == 0 {
			t.Fatalf("ribbon tab boxes missing: %+v %+v", home, insert)
		}
		// Sample just inside the tab's top-left corner, clear of the label.
		on := img.At(int(home.X+3), int(home.Y+home.H/2))
		off := img.At(int(insert.X+3), int(insert.Y+insert.H/2))
		if on == off {
			t.Errorf("dark=%v: active and inactive ribbon tab paint the same %v", dark, on)
		}
		if prev != nil && prev == on {
			t.Errorf("active ribbon tab unchanged between themes: %v", on)
		}
		prev = on
	}
}

// TestCondFormatPixels pins the cell fill from a rule and the data bar paint
// in both themes.
func TestCondFormatPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		a, _ := newTestApp(t)
		setTheme(t, a, dark)
		sh := a.sheet()
		sh.AddCondFormat(Range{Min: CellRef{1, 2}, Max: CellRef{4, 2}}, CondRule{Kind: "gt", A: "9000", Preset: "red"})
		sh.AddCondFormat(Range{Min: CellRef{1, 3}, Max: CellRef{4, 3}}, CondRule{Kind: "databar", Preset: "green"})
		a.redraw()
		img := framePNG(t, a)

		red := cellRect(t, a, CellRef{1, 2}).Inset(2) // C2 = 12500 > 9000
		fill := color.RGBA{0xFF, 0xC7, 0xCE, 255}
		if n := countNear(img, red, fill, 3); n < red.Dx()*red.Dy()/4 {
			t.Errorf("dark=%v: C2 rule fill painted %d/%d px", dark, n, red.Dx()*red.Dy())
		}
		if _, _, ok := sh.DataBar(1, 3); !ok {
			t.Fatalf("dark=%v: D2 has no data bar", dark)
		}
		bar := cellRect(t, a, CellRef{1, 3}).Inset(2)
		sheetBg := tokColor(t, dark, "xl-sheet-bg")
		if n := countNear(img, bar, sheetBg, 4); n > bar.Dx()*bar.Dy()*9/10 {
			t.Errorf("dark=%v: D2 data bar painted nothing (%d bg px of %d)", dark, n, bar.Dx()*bar.Dy())
		}
	}
}
