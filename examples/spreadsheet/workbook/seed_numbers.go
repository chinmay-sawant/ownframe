package workbook

import (
	"fmt"
)

// fillNumbers lays out text labels, signed numbers, formulas, a large
// value, and explicit zeroes that must stay distinct from blanks.
func fillNumbers(w *Workbook, s *Sheet) {
	for r := 0; r < DummyRows; r++ {
		row := r + 1

		w.put(s, Pos{r, 0}, Cell{Kind: Text, Text: fmt.Sprintf("row %d", row)})

		n := float64(r * seedTimes10)
		if r%2 == 1 {
			n = -n
		}

		w.put(s, Pos{r, 1}, Cell{Kind: Number, Number: n})
		w.put(s, Pos{r, 2}, formulaCell("B%d*2", row))
		w.put(s, Pos{r, 3}, formulaCell("SUM(B1:B%d)", row))

		if r%50 == 0 {
			w.put(s, Pos{r, 4}, Cell{Kind: Number, Number: seedBig})
		}

		if r%3 == 0 {
			w.put(s, Pos{r, 5}, Cell{Kind: Number, Number: 0})
		}

		w.put(s, Pos{r, 6}, formulaCell("C%d/2", row))
		w.put(s, Pos{r, 7}, formulaCell("B%d+C%d", row, row))
		w.put(s, Pos{r, 8}, formulaCell("AVERAGE(B1:C%d)", row))
		w.put(s, Pos{r, 9}, formulaCell("C%d-B%d", row, row))

		if r%10 == 0 {
			w.put(s, Pos{r, 10}, Cell{Kind: Number, Number: float64(r * seedTimes10)})
			w.put(s, Pos{r, 19}, Cell{Kind: Number, Number: float64(-r)})
		}
	}
}
