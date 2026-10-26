package main

import "strings"

const (
	diffWrapMinCols = 2
	diffWrapTailDiv = 3
)

// wrapDiffRows expands the diff into fixed-height visual rows when wrap is on.
// With wrap off it returns a clipped copy, so the existing single-row layout
// stays as it was.
func wrapDiffRows(rows []diffRow, unifiedCols, splitCols int, wrap bool) []diffRow {
	if len(rows) == 0 {
		return nil
	}

	if !wrap {
		out := append([]diffRow(nil), rows...)
		clipDiffRows(out, unifiedCols, splitCols)

		return out
	}

	if unifiedCols < diffWrapMinCols {
		unifiedCols = diffWrapMinCols
	}

	if splitCols < diffWrapMinCols {
		splitCols = diffWrapMinCols
	}

	out := make([]diffRow, 0, len(rows))

	for _, row := range rows {
		if row.Kind == "split" {
			out = append(out, wrapSplitDiffRow(row, splitCols)...)

			continue
		}

		out = append(out, wrapUnifiedDiffRow(row, unifiedCols)...)
	}

	return out
}

// wrapUnifiedDiffRow wraps the single cell of a unified or metadata row.
func wrapUnifiedDiffRow(row diffRow, cols int) []diffRow {
	if row.Hunk != "" {
		return wrapHunkDiffRow(row, cols)
	}

	chunks := wrapDiffSegments(row.Left, cols)
	if len(chunks) == 0 {
		chunks = [][]segment{nil}
	}

	out := make([]diffRow, 0, len(chunks))

	for i, chunk := range chunks {
		vis := row
		vis.Left = chunk
		vis.Right = nil

		if i > 0 {
			vis.OldNo = ""
			vis.NewNo = ""
			vis.Cont = true
		}

		out = append(out, vis)
	}

	return out
}

// wrapHunkDiffRow wraps the label of a hunk or meta row. Only the text moves;
// the row keeps its kind on every chunk.
func wrapHunkDiffRow(row diffRow, cols int) []diffRow {
	chunks := wrapDiffSegments([]segment{{Text: row.Hunk}}, cols)
	if len(chunks) == 0 {
		chunks = [][]segment{nil}
	}

	out := make([]diffRow, 0, len(chunks))

	for i, chunk := range chunks {
		vis := row
		vis.Hunk = diffSegmentText(chunk)
		vis.Left = nil
		vis.Right = nil
		vis.LeftPad = false
		vis.RightPad = false

		if i > 0 {
			vis.OldNo = ""
			vis.NewNo = ""
			vis.Cont = true
		}

		out = append(out, vis)
	}

	return out
}

// wrapSplitDiffRow wraps the left and right cells independently. The taller
// side decides the visual row count; the shorter side keeps its kind and is
// padded once its text runs out.
func wrapSplitDiffRow(row diffRow, cols int) []diffRow {
	left := wrapDiffSegments(row.Left, cols)
	right := wrapDiffSegments(row.Right, cols)

	height := max(len(left), len(right))
	if height == 0 {
		height = 1
	}

	out := make([]diffRow, 0, height)

	for i := range height {
		vis := row
		vis.Left = nil
		vis.Right = nil
		vis.LeftPad = row.LeftPad
		vis.RightPad = row.RightPad
		vis.LeftCont = false
		vis.RightCont = false

		if i < len(left) {
			vis.Left = left[i]
			vis.LeftPad = false
			vis.LeftCont = i > 0

			if i > 0 {
				vis.LeftNo = ""
			}
		} else if i > 0 {
			vis.LeftNo = ""
			vis.LeftPad = true
		}

		if i < len(right) {
			vis.Right = right[i]
			vis.RightPad = false
			vis.RightCont = i > 0

			if i > 0 {
				vis.RightNo = ""
			}
		} else if i > 0 {
			vis.RightNo = ""
			vis.RightPad = true
		}

		out = append(out, vis)
	}

	return out
}

// wrapDiffSegments cuts one styled cell into chunks of at most cols runes. It
// prefers the last space in the final third of the chunk and hard breaks when
// there is no usable space.
func wrapDiffSegments(segs []segment, cols int) [][]segment {
	if cols < diffWrapMinCols {
		cols = diffWrapMinCols
	}

	cells := make([]wrapCell, 0, len(segs)*wrapCellEstimate)

	for _, seg := range segs {
		for _, r := range seg.Text {
			cells = append(cells, wrapCell{r: r, class: seg.Class})
		}
	}

	var chunks [][]segment

	for len(cells) > 0 {
		cut := min(cols, len(cells))

		if cut < len(cells) {
			floor := cut - max(cols/diffWrapTailDiv, 1)
			if floor < 1 {
				floor = 1
			}

			for i := cut - 1; i >= floor; i-- {
				if cells[i].r == ' ' {
					cut = i + 1

					break
				}
			}
		}

		chunks = append(chunks, cellsToSegments(cells[:cut]))
		cells = cells[cut:]
	}

	return chunks
}

// diffSegmentText joins one wrapped chunk back into plain text.
func diffSegmentText(segs []segment) string {
	var b strings.Builder

	for _, seg := range segs {
		b.WriteString(seg.Text)
	}

	return b.String()
}
