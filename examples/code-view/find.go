package main

import (
	"context"
	"strings"
	"unicode/utf8"
)

// findHit is one match in source-line rune coordinates.
type findHit struct {
	Line  int
	Start int
	End   int
}

// findCurrent is the 1-based visible match number.
func (a *App) findCurrent() int {
	if !a.findOpen || len(a.findHits) == 0 {
		return 0
	}

	return a.findIndex + 1
}

// reindexFind rebuilds the match list for the active document.
func (a *App) reindexFind() {
	a.findHits = nil
	a.findIndex = 0

	doc := a.activeDoc()
	if doc == nil || doc.Err != "" || doc.Binary {
		return
	}

	query := strings.TrimSpace(a.findQuery)
	if query == "" {
		return
	}

	if !a.findCase {
		query = strings.ToLower(query)
	}

	for lineNo, line := range doc.Lines {
		hay := line
		if !a.findCase {
			hay = strings.ToLower(line)
		}

		for from := 0; from < len(hay); {
			idx := strings.Index(hay[from:], query)
			if idx < 0 {
				break
			}

			start := utf8.RuneCountInString(hay[:from+idx])
			end := start + utf8.RuneCountInString(query)
			a.findHits = append(a.findHits, findHit{Line: lineNo, Start: start, End: end})
			from += idx + len(query)
		}
	}
}

// openFind shows the find bar and focuses its field.
func (a *App) openFind(ctx context.Context) error {
	if a.mode != "explorer" {
		a.showToast("info", "Find", "Find searches the file view. Return to Explorer.")
		a.refresh()

		return nil
	}

	if a.activeDoc() == nil {
		return nil
	}

	if !a.findOpen {
		a.findReturn = a.page.Form().FocusID()
		a.findOpen = true
	}

	a.reindexFind()
	a.refresh()

	if err := a.page.Redraw(ctx); err != nil {
		return err
	}

	if err := a.page.Form().Focus(ctx, "find"); err != nil {
		return err
	}

	return a.page.Edit().SelectAll(ctx)
}

// closeFind hides the bar and restores the previous focus when possible.
func (a *App) closeFind(ctx context.Context) error {
	previous := a.findReturn
	a.findOpen = false
	a.findQuery = ""
	a.findCase = false
	a.findHits = nil
	a.findIndex = 0
	a.findReturn = ""
	a.refresh()

	if err := a.page.Redraw(ctx); err != nil {
		return err
	}

	if previous != "" && a.page.Form().FocusID() == "" {
		return a.page.Form().Focus(ctx, previous)
	}

	return nil
}

// findStep moves to the next or previous match and reveals it.
func (a *App) findStep(delta int) {
	if len(a.findHits) == 0 {
		a.refresh()

		return
	}

	a.findIndex = (a.findIndex + delta + len(a.findHits)) % len(a.findHits)
	hit := a.findHits[a.findIndex]
	a.refresh()
	a.scrollToSourceLine(hit.Line)
}

// toggleFindCase flips case sensitivity and rebuilds the matches.
func (a *App) toggleFindCase() {
	a.findCase = !a.findCase
	a.reindexFind()
	a.refresh()
}

// markFindRows adds the match classes to a copy of the visible rows.
func (a *App) markFindRows(rows []codeRow) {
	if !a.findOpen || len(a.findHits) == 0 {
		return
	}

	for i := range rows {
		row := &rows[i]
		if row.Segments == nil {
			continue
		}

		start := row.Start
		length := 0
		for _, seg := range row.Segments {
			length += utf8.RuneCountInString(seg.Text)
		}

		end := start + length

		for hitIndex, hit := range a.findHits {
			if hit.Line != row.Line {
				continue
			}

			from := max(hit.Start, start)
			to := min(hit.End, end)
			if from >= to {
				continue
			}

			extra := "find-hit"
			if hitIndex == a.findIndex {
				extra += " find-cur"
			}

			row.Segments = markRange(row.Segments, from-start, to-start, extra)
		}
	}
}

// copyFindRows copies the visible rows before a find mark mutates them.
func (a *App) copyFindRows(rows []codeRow) []codeRow {
	if !a.findOpen || len(a.findHits) == 0 || len(rows) == 0 {
		return rows
	}

	out := append([]codeRow(nil), rows...)
	a.markFindRows(out)

	return out
}
