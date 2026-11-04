package main

import (
	"context"
	"errors"
	"strings"
	"time"

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
	toastUntil   time.Time
	notifOpen    bool
	userOpen     bool
	notifs       []notifItem
	selectedKind string
	selectedID   string
	heatDay      int
	heatCol      int
	growthHover  int
	// enterBars renders the List growth bars at scaleY(0) for the load
	// entrance; enterSlow carries the mp-bar-slow class for the 3s run.
	// enterSeen marks that the zero-scale bars committed on their own frame,
	// so the flip lands in a later frame (the window batches one frame's
	// pointer-to-tick refreshes into a single commit, which would otherwise
	// swallow the zero state and leave the run with no start paint).
	enterBars bool
	enterSlow bool
	enterSeen bool
}

func newApp() (*app, error) {
	a := &app{
		tab:         "engagement",
		rangeLabel:  "Last 30 Days",
		heatDay:     -1,
		heatCol:     -1,
		growthHover: -1,
		notifs:      defaultNotifs(),
		enterBars:   true,
		enterSlow:   true,
	}
	view, err := ui.Mount(ownframe.Config{
		Title:       "MailPulse · Engagement",
		Width:       1440,
		Height:      960,
		MinWidth:    1000,
		MinHeight:   800,
		Theme:       chadcn.Light + chadcn.Stylesheet + appCSS + hoverCSS + topbarFixCSS + cardsFixCSS + layoutFixCSS + dataFixCSS,
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
		Click:   a.handleClick,
		Change:  a.handleChange,
		Hover:   a.handleHover,
		Submit:  a.handleSubmit,
		KeyDown: a.handleKey,
	})
	a.escortEnter()
	return a, nil
}

func (a *app) refresh(ctx context.Context) error {
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return err
	}
	a.ensureToastTick()
	return nil
}

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	panelsChanged := false
	if a.notifOpen || a.userOpen {
		if !strings.HasPrefix(box.Action, "notif:") && !strings.HasPrefix(box.Action, "user:") {
			a.notifOpen = false
			a.userOpen = false
			panelsChanged = true
		}
	}
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
		if tab == "overview" {
			a.enterBars = true
			a.enterSlow = true
			a.enterSeen = false
			a.escortEnter()
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
		a.escortMotion(1500 * time.Millisecond)
		return a.refresh(ctx)
	case box.Action == "export":
		a.toast = "Export queued — CSV will land in your inbox."
		a.toastOpen = true
		return a.refresh(ctx)
	case box.Action == "toast:close":
		a.toastOpen = false
		a.toastUntil = time.Time{}
		if a.page != nil {
			a.page.SetTick(nil)
		}
		return a.refresh(ctx)
	case box.Action == "search:clear":
		a.query = ""
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "notif:"):
		if strings.HasPrefix(box.Action, "notif:dismiss:") {
			a.dismissNotif(strings.TrimPrefix(box.Action, "notif:dismiss:"))
		}
		return a.refresh(ctx)
	case box.Action == "user:profile":
		a.userOpen = false
		a.toast = "Profile opening soon"
		a.toastOpen = true
		return a.refresh(ctx)
	case box.Action == "user:signout":
		a.userOpen = false
		a.toast = "Signed out (demo)"
		a.toastOpen = true
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "select:"):
		rest := strings.TrimPrefix(box.Action, "select:")
		parts := strings.SplitN(rest, ":", 3)
		if len(parts) != 3 {
			return nil
		}
		return a.selectToggle(ctx, parts[0], parts[1], parts[2])
	default:
		if panelsChanged {
			return a.refresh(ctx)
		}
		return nil
	}
}

func (a *app) isSelected(kind, id string) bool {
	return a.selectedKind == kind && a.selectedID == id
}

func (a *app) selectToggle(ctx context.Context, kind, id, label string) error {
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

// handleHover updates the heatmap and List growth readouts without a full
// re-render. Other elements use CSS :hover only, so they report no change.
func (a *app) handleHover(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	if i, ok := parseGrowthBarID(next.ID); ok {
		a.growthHover = i
		if err := a.patchElementText(ctx, "mp-growth-readout", a.growthBarLabelFor(i)); err != nil {
			return false, err
		}
		if _, _, wasHeat := parseHeatCellID(previous.ID); wasHeat {
			if err := a.patchReadout(ctx, a.heatReadout()); err != nil {
				return false, err
			}
		}
		return false, nil
	}
	if _, ok := parseGrowthBarID(previous.ID); ok {
		a.growthHover = -1
		if err := a.patchElementText(ctx, "mp-growth-readout", a.growthReadoutDefaultFor()); err != nil {
			return false, err
		}
	}
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
	return a.patchElementText(ctx, "mp-heat-readout", label)
}

func (a *app) patchElementText(ctx context.Context, id, label string) error {
	err := a.page.PatchElement(ctx, id, ownframe.ElementEdit{Text: &label})
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return a.refresh(ctx)
	}
	return err
}

func (a *app) themeSheet() string {
	return chadcn.Light + chadcn.Stylesheet + appCSS + hoverCSS + topbarFixCSS + cardsFixCSS + layoutFixCSS + dataFixCSS
}
