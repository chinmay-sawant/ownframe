package main

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// app owns the MailPulse engagement dashboard state. Panels are pure
// functions of this state and are rebuilt through View.Update after every
// action.
type app struct {
	view         *ui.View
	page         *ownframe.Page
	tab          string
	query        string
	subQuery     string
	rangeLabel   string
	toast        string
	toastOpen    bool
	selectedKind string
	selectedID   string
	heatDay      int
	heatCol      int
}

func newApp() (*app, error) {
	a := &app{
		tab:        "engagement",
		rangeLabel: "Last 30 Days",
		heatDay:    -1,
		heatCol:    -1,
	}
	view, err := ui.Mount(ownframe.Config{
		Title:       "MailPulse · Engagement",
		Width:       1440,
		Height:      960,
		MinWidth:    1180,
		MinHeight:   800,
		Theme:       chadcn.Light + chadcn.Stylesheet + appCSS + hoverCSS,
		DevTools:    false,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	view.Handle(ownframe.Handlers{
		Click:  a.handleClick,
		Change: a.handleChange,
		Hover:  a.handleHover,
	})
	return a, nil
}

func (a *app) refresh(ctx context.Context) error {
	return a.view.Update(ctx, a.screen())
}

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	switch {
	case strings.HasPrefix(box.Action, "nav:") || strings.HasPrefix(box.Action, "nav-"):
		tab := tabFromAction(box.Action)
		if tab == "" {
			return nil
		}
		a.tab = tab
		a.selectedKind = ""
		a.selectedID = ""
		if a.page != nil {
			a.page.ScrollTo(0, 0)
		}
		return a.refresh(ctx)
	case box.Action == "range:cycle":
		switch a.rangeLabel {
		case "Last 7 Days":
			a.rangeLabel = "Last 30 Days"
		case "Last 30 Days":
			a.rangeLabel = "Last 90 Days"
		default:
			a.rangeLabel = "Last 7 Days"
		}
		return a.refresh(ctx)
	case box.Action == "export":
		a.toast = "Export queued — CSV will land in your inbox."
		a.toastOpen = true
		return a.refresh(ctx)
	case box.Action == "toast:close":
		a.toastOpen = false
		return a.refresh(ctx)
	case box.Action == "search:clear":
		a.query = ""
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "select:"):
		rest := strings.TrimPrefix(box.Action, "select:")
		parts := strings.SplitN(rest, ":", 3)
		if len(parts) != 3 {
			return nil
		}
		return a.selectToggle(ctx, parts[0], parts[1], parts[2])
	default:
		return nil
	}
}

func (a *app) isSelected(kind, id string) bool {
	return a.selectedKind == kind && a.selectedID == id
}

func (a *app) selectToggle(ctx context.Context, kind, id, label string) error {
	if kind == "heat" {
		if day, col, ok := parseHeatKey(id); ok {
			a.heatDay = day
			a.heatCol = col
		}
	}
	if a.isSelected(kind, id) {
		a.selectedKind = ""
		a.selectedID = ""
		a.toast = "Cleared selection"
		a.toastOpen = true
		return a.refresh(ctx)
	}
	a.selectedKind = kind
	a.selectedID = id
	a.toast = "Selected " + label
	a.toastOpen = true
	return a.refresh(ctx)
}

// parseHeatKey recovers the weekday and column from a "day-col" key.
func parseHeatKey(key string) (int, int, bool) {
	day, col, ok := strings.Cut(key, "-")
	if !ok {
		return 0, 0, false
	}
	d, err := strconv.Atoi(day)
	if err != nil {
		return 0, 0, false
	}
	c, err := strconv.Atoi(col)
	if err != nil {
		return 0, 0, false
	}
	if d < 0 || d > 6 || c < 0 || c > 17 {
		return 0, 0, false
	}
	return d, c, true
}

// handleChange keeps the topbar and subscriber search fields live. The page
// form owns the keystrokes; the host mirrors them into filter state and
// re-renders so the subscriber table filters as you type.
func (a *app) handleChange(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "sub-search":
		a.subQuery = a.page.Form().FormValue(box.ID)
		return a.refresh(ctx)
	case "mp-search":
		a.query = a.page.Form().FormValue(box.ID)
		return a.refresh(ctx)
	default:
		return nil
	}
}

// handleHover updates the heatmap readout without a full re-render. Other
// elements use CSS :hover only, so they report no change.
func (a *app) handleHover(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	day, col, ok := parseHeatCellID(next.ID)
	if !ok {
		if _, _, was := parseHeatCellID(previous.ID); was {
			label := a.heatReadout()
			if err := a.patchReadout(ctx, label); err != nil {
				return false, err
			}
		}
		return false, nil
	}
	a.heatDay = day
	a.heatCol = col
	label := heatCellLabel(day, col)
	if err := a.patchReadout(ctx, label); err != nil {
		return false, err
	}
	return false, nil
}

func (a *app) patchReadout(ctx context.Context, label string) error {
	err := a.page.PatchElement(ctx, "mp-heat-readout", ownframe.ElementEdit{Text: &label})
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return a.refresh(ctx)
	}
	return err
}

func (a *app) themeSheet() string {
	return chadcn.Light + chadcn.Stylesheet + appCSS + hoverCSS
}
