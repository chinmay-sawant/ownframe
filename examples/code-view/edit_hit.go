package main

// visualRowAt maps a click inside the code body to the zero-based visual row
// under it. bodyTop is the fixed viewport top of the body and scrollY is the
// page scroll offset. A click above the body clamps to the first row, and a
// row height of zero has no rows to hit.
func visualRowAt(y, bodyTop, scrollY, rowH int) int {
	if rowH <= 0 {
		return 0
	}

	if scrollY < 0 {
		scrollY = 0
	}

	local := y - bodyTop + scrollY
	if local < 0 {
		return 0
	}

	return local / rowH
}

// sourceLineAt returns the zero-based source line that owns one visual row,
// using the document's cached wrap plan. A row past the plan clamps to the
// last source line, and a nil document has no line to report.
func sourceLineAt(doc *fileDoc, cols, visualRow int) int {
	if doc == nil {
		return 0
	}

	if visualRow < 0 {
		visualRow = 0
	}

	return doc.wrapped(cols).lineAt(visualRow)
}

// editScrollForLine returns the page scroll offset that puts one source line
// at the top of the code body. The line's first visual row comes from the
// document's cached wrap plan, so a wrapped line lands on its first row.
func editScrollForLine(doc *fileDoc, cols, line int) int {
	if doc == nil {
		return 0
	}

	return doc.wrapped(cols).firstRow(line) * rowH
}
