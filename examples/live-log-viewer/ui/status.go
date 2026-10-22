package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// statusText is the left footer: dataset size, active filters, or a
// transient note.
func (a *App) statusText(now time.Time) string {
	if a.note != "" && now.Before(a.noteUntil) {
		return a.note
	}

	parts := []string{
		strconv.Itoa(len(a.sources)) + " sources",
		strconv.Itoa(a.pager.Total) + " entries",
	}

	if a.activeSource != "" {
		parts = append(parts, "source "+a.activeSourceName())
	}

	if a.filters.Text != "" {
		parts = append(parts, "text \u201c"+truncate(a.filters.Text, uiTextLen)+"\u201d")
	}

	if a.minSev != "" {
		parts = append(parts, a.minSev)
	}

	return strings.Join(parts, " \u00b7 ")
}

// rightText is the right footer: follow state, anchor state, and perf.
func (a *App) rightText(time.Time) string {
	mode := "live"
	if a.pager.HWM != 0 {
		mode = "frozen at #" + strconv.FormatInt(a.pager.HWM, 10)
	}

	if a.follow.Paused {
		mode = "paused \u00b7 " + mode
	}

	if a.perf {
		mode += fmt.Sprintf(" \u00b7 redraw %.1f/%.1fms", ms(a.lastDraw), ms(a.maxDraw))

		if d := a.dropped.Load(); d > 0 {
			mode += fmt.Sprintf(" \u00b7 coalesced %d", d)
		}
	}

	return mode
}
