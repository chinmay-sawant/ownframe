package ui

import "strconv"

// paintChrome marks the header, the detail card, or the footer. The footer
// stays on its own frame: it sits below the history, and a union with a
// queue row would repaint the whole page.
func (a *App) paintChrome(quiet bool) {
	sums := a.paintSums()
	detail := a.paintDetail()
	foot := setText(a.bindings.foot, a.footText())

	if !quiet {
		a.markUpper(sums, detail || a.detailWait)
		a.detailWait = false

		if foot {
			a.footWait = true
		}

		return
	}

	if foot || a.footWait {
		a.page.Invalidate("foot")
		a.footWait = false
		a.noteWait(sums, detail)

		return
	}

	a.markUpper(sums || a.sumWait, detail || a.detailWait)
	a.sumWait = false
	a.detailWait = false
}

// noteWait remembers header work that must not share the footer's frame.
func (a *App) noteWait(sums, detail bool) {
	if sums {
		a.sumWait = true
	}

	if detail {
		a.detailWait = true
	}
}

// markUpper marks the header chips and the detail card.
func (a *App) markUpper(sums, detail bool) {
	if sums {
		a.page.Invalidate("sums")
	}

	if detail {
		a.page.Invalidate("detail-body")
	}
}

// paintSums writes the four header counts.
func (a *App) paintSums() bool {
	vals := [4]string{
		strconv.Itoa(a.view.Summary.Active()),
		strconv.Itoa(a.view.Summary.Running),
		strconv.Itoa(a.view.Summary.Completed),
		strconv.Itoa(a.view.Summary.Failed),
	}
	changed := false

	for i, val := range vals {
		if setText(a.bindings.sum[i], val) {
			changed = true
		}
	}

	return changed
}

// paintDetail writes the open job's live fields.
func (a *App) paintDetail() bool {
	detail := a.view.Detail
	if detail == nil {
		return false
	}

	changed := setText(a.bindings.detail[0], detail.StateText())

	if setText(a.bindings.detail[1], detail.SizeText()) {
		changed = true
	}

	if setText(a.bindings.detail[2], detail.SpeedText()) {
		changed = true
	}

	return changed
}
