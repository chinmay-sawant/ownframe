package workbook

// fillMixed mixes types and ships deliberate formula errors: division by
// zero, an unknown function, an out-of-range reference, a text operand, and
// a mutual cycle.
func fillMixed(w *Workbook, s *Sheet) {
	for r := 0; r < DummyRows; r++ {
		row := r + 1

		w.put(s, Pos{r, 0}, Cell{Kind: Text, Text: "mix " + itoa(row)})
		w.put(s, Pos{r, 1}, Cell{Kind: Number, Number: float64(r - seedOffset)})

		if r == 5 || r == 15 || r == 25 {
			w.put(s, Pos{r, 2}, formulaCell("B%d/0", row))
		} else {
			w.put(s, Pos{r, 2}, formulaCell("B%d*3", row))
		}

		if r == seedRow10 {
			w.put(s, Pos{r, 4}, Cell{Kind: Formula, Source: "NOSUCH1(1)"})
		}

		if r == seedRow12 {
			w.put(s, Pos{r, 4}, Cell{Kind: Formula, Source: "Z99999999"})
		}

		if r == seedRow40 {
			w.put(s, Pos{r, 5}, Cell{Kind: Formula, Source: "F42"})
		}

		if r == seedRow41 {
			w.put(s, Pos{r, 5}, Cell{Kind: Formula, Source: "F41"})
		}

		if r%20 == 0 {
			w.put(s, Pos{r, 6}, formulaCell("B%d+H%d", row, row))
			w.put(s, Pos{r, 7}, Cell{Kind: Text, Text: "not a number"})
		}

		if r%25 == 0 && r > 0 {
			w.put(s, Pos{r, 8}, formulaCell("SUM(B1:B%d)", row))
		}

		if r >= 50 && r < 60 {
			w.put(s, Pos{r, 9}, Cell{Kind: Number, Number: float64(r)})
		}
	}
}

// itoa keeps the seed deterministic without touching the fmt package.
func itoa(n int) string {
	const digits = "0123456789"

	if n == 0 {
		return "0"
	}

	var b [20]byte

	i := len(b)

	for n > 0 {
		i--
		b[i] = digits[n%10]
		n /= 10
	}

	return string(b[i:])
}
