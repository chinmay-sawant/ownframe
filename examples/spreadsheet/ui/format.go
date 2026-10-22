package ui

import (
	"encoding/csv"
	"strconv"
	"strings"
)

// colName returns the column label for a zero-based index: A, B, ..., Z,
// AA, AB, and so on.
func colName(c int) string {
	if c < 0 {
		return ""
	}

	var b [8]byte
	i := len(b)
	for c >= 0 {
		i--
		b[i] = byte('A' + c%uiAlphaBase)
		c = c/uiAlphaBase - 1
	}

	return string(b[i:])
}

// ref names one cell, such as "B3".
func ref(r, c int) string { return colName(c) + strconv.Itoa(r+1) }

// parseRef reads a cell name into zero-based coordinates. The second result
// is false when the name is not a cell reference.
func parseRef(s string) (r, c int, ok bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	i := 0
	for i < len(s) && s[i] >= 'A' && s[i] <= 'Z' {
		c = c*uiAlphaBase + int(s[i]-'A'+1)
		i++
	}

	if i == 0 || i == len(s) {
		return 0, 0, false
	}

	n, err := strconv.Atoi(s[i:])
	if err != nil || n < 1 {
		return 0, 0, false
	}

	return n - 1, c - 1, true
}

// encodeTSV writes cells as tab-separated values, one line per row, using
// quoting for tabs, newlines, and quotes.
func encodeTSV(cells []Cell, w, h int) string {
	var sb strings.Builder
	rec := make([]string, w)
	out := csv.NewWriter(&sb)
	out.Comma = '\t'

	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			i := r*w + c
			if i < len(cells) {
				rec[c] = cells[i].Raw
			} else {
				rec[c] = ""
			}
		}

		_ = out.Write(rec)
	}

	out.Flush()

	return sb.String()
}

// decodeTSV reads tab-separated values. A line that fails to parse ends the
// input, so a malformed paste still applies its leading rows.
func decodeTSV(s string) [][]string {
	in := csv.NewReader(strings.NewReader(s))
	in.Comma = '\t'
	in.FieldsPerRecord = -1

	rows, err := in.ReadAll()
	if err != nil && len(rows) == 0 {
		return nil
	}

	return rows
}

func itoa(i int) string { return strconv.Itoa(i) }
