package ui

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// onClick routes one click. The page redraws after this returns.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	switch action := box.Action; {
	case action == "detail-back":
		a.closeDetail()
	case action == "older":
		a.loadPage(intentOlder)
	case action == "newer":
		a.loadPage(intentNewer)
	case action == "follow":
		a.toggleFollow()
	case action == "pause":
		a.togglePause()
	case action == "newest":
		a.jumpNewest()
	case action == "theme":
		a.toggleTheme()
	case action == "export":
		a.startExport()
	case action == "sev-clear":
		a.clearFilters()
	case strings.HasPrefix(action, "sev-"):
		a.pickSev(strings.TrimPrefix(action, "sev-"))
	case strings.HasPrefix(action, "source-"):
		a.pickSource(strings.TrimPrefix(action, "source-"))
	case strings.HasPrefix(action, "row-"):
		a.pickRow(strings.TrimPrefix(action, "row-"))
	}

	return nil
}

// onChange debounces search edits.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	if box.ID != "search" {
		return nil
	}

	a.filters.Edit(a.page.Form().FormValue("search"), time.Now(), SearchDelay)

	return nil
}

// pickRow selects a row; clicking the selected row opens its detail.
func (a *App) pickRow(text string) {
	id, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return
	}

	if id == a.selected && !a.detailOpen {
		a.openDetail(id)

		return
	}

	a.selected = id
	a.selAnchor = id
	a.setWindow(a.winStart, a.winEnd)
}

// openDetail requests the full message and shows the detail pane.
func (a *App) openDetail(id int64) {
	a.selected = id
	a.selAnchor = id
	a.requestDetail(id)
}

// closeDetail returns to the list at the reading anchor.
func (a *App) closeDetail() {
	a.detailOpen = false
	a.detailID = 0
	a.scrollTo(a.pager.AnchorOffset(HeaderH))
}
