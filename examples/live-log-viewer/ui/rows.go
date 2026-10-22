package ui

import "strings"

// SourceRow is one sidebar source.
type SourceRow struct {
	ID          string
	Name        string
	Status      string
	StatusClass string
	Count       int
	Active      bool
}

// sourceRows builds the sidebar list; activeKey is "" for all sources.
func sourceRows(srcs []SourceInfo, activeKey string) []SourceRow {
	out := make([]SourceRow, 0, len(srcs)+1)
	out = append(out, SourceRow{
		ID: "all", Name: "All sources", Status: "all", StatusClass: "idle",
		Active: activeKey == "",
	})

	for _, s := range srcs {
		out = append(out, SourceRow{
			ID: s.Key, Name: truncate(s.Name, uiNameLen), Status: s.State,
			StatusClass: sourceClass(s.State), Count: s.Count, Active: s.Key == activeKey,
		})
	}

	return out
}

// sourceClass maps a source state to its CSS class suffix.
func sourceClass(state string) string {
	switch strings.ToLower(state) {
	case "live":
		return "live"
	case "missing", "error":
		return "missing"
	}

	return "idle"
}

// buildRows turns entries into display rows for a page width.
func buildRows(entries []Entry, width int, selected int64) []Row {
	budget := textBudget(width)
	rows := make([]Row, len(entries))

	for i, e := range entries {
		text := firstLine(strings.ReplaceAll(e.Text, "\t", "    "))
		rows[i] = Row{
			ID:        e.ID,
			Time:      formatTime(e),
			Source:    truncate(e.Source, uiSrcLen),
			Severity:  sevLabel(e.Severity),
			SevClass:  sevClass(e.Severity),
			Text:      truncate(text, budget),
			Lines:     e.Lines,
			Truncated: e.Truncated,
			Partial:   e.Partial,
			Lost:      e.Lost,
			Selected:  e.ID != 0 && e.ID == selected,
			Alt:       i%2 == 1,
		}
	}

	return rows
}
