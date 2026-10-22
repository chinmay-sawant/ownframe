package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// pageLabel is the loaded-window label in the footer.
func (a *App) pageLabel() string {
	if a.pager.Total > 0 {
		return fmt.Sprintf("%d rows of %d", a.pager.Len(), a.pager.Total)
	}

	return fmt.Sprintf("%d rows", a.pager.Len())
}

// detailView builds the detail pane data, splitting the full message when
// the feed did not already do it.
func (a *App) detailView() DetailView {
	d := a.detail
	lines := d.Lines
	if lines == nil {
		lines = splitLines(d.Text)
	}

	parts := []string{
		"#" + strconv.FormatInt(d.ID, 10),
		formatTime(d.Entry),
		d.Source,
		sevLabel(d.Severity),
		bytesText(int64(d.Bytes)),
	}

	return DetailView{
		Meta:      strings.Join(parts, " \u00b7 "),
		Lines:     lines,
		Truncated: d.Truncated,
		Partial:   d.Partial,
	}
}

// activeSourceName returns the display name of the active source.
func (a *App) activeSourceName() string {
	for _, s := range a.sources {
		if s.Key == a.activeSource {
			return s.Name
		}
	}

	return a.activeSource
}

// ms converts a duration to fractional milliseconds.
func ms(d time.Duration) float64 {
	return float64(d.Microseconds()) / uiMsDiv
}
