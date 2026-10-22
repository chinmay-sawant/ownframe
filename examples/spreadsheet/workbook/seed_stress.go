package workbook

import (
	"strconv"
)

// demoLabels rotate through the dummy workbook's text sheet.
var demoLabels = [...]string{
	"Türkçe metin",
	"Ελληνικά",
	"日本語のテキスト",
	"مرحبا بالعالم",
	"Привет мир",
	"emoji 🙂 and ∑ symbols",
	"naïve café",
	"Ünïcödé",
}

// SeedStress builds a sparse workbook of rows rows and returns it with
// formula values computed. Every row stores one number, every seventh row a
// text note, and formulas appear every fifty rows.
func SeedStress(rows int) *Workbook {
	w := New(stressSheets, "Stress sparse")
	s := w.AddSheet("Sparse")

	for r := 0; r < rows; r++ {
		w.put(s, Pos{r, 0}, Cell{Kind: Number, Number: float64(r)})

		if r%7 == 0 {
			w.put(s, Pos{r, 1}, Cell{Kind: Text, Text: "note " + strconv.Itoa(r)})
		}

		if r%50 == 0 && r > 0 {
			w.put(s, Pos{r, 2}, formulaCell("A%d+1", r))
		}

		if r%500 == 0 && r >= 50 {
			w.put(s, Pos{r, 3}, formulaCell("SUM(A%d:A%d)", r-stressBack, r+1))
		}
	}

	w.RecalcAll()

	return w
}
