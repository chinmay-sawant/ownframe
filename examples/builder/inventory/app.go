package main

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	inventoryViewportID = "inv-body-viewport"
	inventoryContentID  = "inv-body-content"

	inventoryWindowWidth     = 1440
	inventoryWindowHeight    = 960
	inventoryWindowMinWidth  = 900
	inventoryWindowMinHeight = 600
)

type app struct {
	view     *ui.View
	page     *ownframe.Page
	period   string
	category string

	poOpen       bool
	searchOpen   bool
	searchQuery  string
	searchPicked string
	searchOffset int
	statusFilter string
	selectedPO   string
	toastOpen    bool
	toastTitle   string
	toastBody    string
	hoverBin     string
	notifOpen    bool
	userMenuOpen bool

	bodyOffset int
	bodyFrac   float64
}

func newApp() (*app, error) {
	a := &app{
		period:   "30 days",
		category: "All categories",
	}
	view, err := ui.Mount(ownframe.Config{
		Title:       "Northwind Goods · Inventory health",
		Width:       inventoryWindowWidth,
		Height:      inventoryWindowHeight,
		MinWidth:    inventoryWindowMinWidth,
		MinHeight:   inventoryWindowMinHeight,
		LockView:    true,
		Theme:       chadcn.Light + chadcn.Stylesheet + appCSS,
		DevTools:    false,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	a.page.Scroll().SetAllowScroll(false)
	view.Handle(ownframe.Handlers{
		Click:   a.handleClick,
		Wheel:   a.handleWheel,
		Hover:   a.handleHover,
		Change:  a.handleChange,
		KeyDown: a.handleKey,
	})
	return a, nil
}

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	// Click-away dismisses the popovers, mirroring loglens: a click whose
	// box is outside the popover subtree closes it. Trigger clicks never
	// reach here — the component wires them to OnOpenChange instead.
	dismissed := false
	if a.notifOpen && !strings.HasPrefix(box.ID, "notifications-popover") {
		a.notifOpen = false
		dismissed = true
	}
	if a.userMenuOpen && !strings.HasPrefix(box.ID, "user-menu") {
		a.userMenuOpen = false
		dismissed = true
	}
	switch {
	case box.Action == "win-min":
		a.page.Window().Minimize()
		return nil
	case box.Action == "win-max":
		a.page.Window().ToggleMaximize()
		return nil
	case box.Action == "win-close":
		a.page.Window().Close()
		return nil
	case strings.HasPrefix(box.Action, "period:"):
		a.period = strings.TrimPrefix(box.Action, "period:")
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "category:"):
		a.category = strings.TrimPrefix(box.Action, "category:")
		return a.refresh(ctx)
	case box.Action == "po:new":
		a.poOpen = true
		return a.refresh(ctx)
	case box.Action == "po:close" || box.Action == "po:create":
		a.poOpen = false
		return a.refresh(ctx)
	case box.Action == "search:open":
		a.searchOpen = !a.searchOpen
		a.searchOffset = 0
		return a.refresh(ctx)
	case box.Action == "search:close":
		if !a.searchOpen {
			return nil
		}
		a.searchOpen = false
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "search:pick:"):
		a.searchPicked = strings.TrimPrefix(box.Action, "search:pick:")
		a.searchOpen = false
		a.searchOffset = 0
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "status:"):
		a.statusFilter = strings.TrimPrefix(box.Action, "status:")
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "po:view:"):
		a.selectedPO = strings.TrimPrefix(box.Action, "po:view:")
		return a.refresh(ctx)
	case box.Action == "po:viewall":
		a.selectedPO = ""
		return a.refresh(ctx)
	case box.Action == "export:run":
		a.toastOpen = true
		a.toastTitle = "Export queued"
		a.toastBody = "The stock-health CSV will land in your inbox."
		return a.refresh(ctx)
	case box.Action == "toast:close":
		a.toastOpen = false
		return a.refresh(ctx)
	}
	if dismissed {
		return a.refresh(ctx)
	}
	return nil
}

func (a *app) handleKey(ctx context.Context, key string) error {
	if key != "escape" {
		return nil
	}
	if !a.poOpen && !a.searchOpen && !a.notifOpen && !a.userMenuOpen && !a.toastOpen {
		return nil
	}
	a.poOpen = false
	a.searchOpen = false
	a.notifOpen = false
	a.userMenuOpen = false
	a.toastOpen = false
	return a.refresh(ctx)
}

// overlaysOpen reports whether any overlay surface is mounted. The window
// drag guard (main.go) uses it to disable the draggable strip while an
// overlay owns its presses.
func (a *app) overlaysOpen() bool {
	return a.searchOpen || a.notifOpen || a.userMenuOpen || a.poOpen || a.toastOpen
}

func (a *app) refresh(ctx context.Context) error {
	return a.view.Update(ctx, a.screen())
}

func (a *app) screen() *ui.Element {
	return ui.Div().Class("inv-shell").
		BackgroundColor(ui.ColorVar("--of-background", ui.Hex("#f8fafc"))).
		Children(
			a.searchBackdrop(),
			a.topbar(),
			ui.Div().ID(inventoryViewportID).Class("inv-body-viewport").Children(
				ui.Div().ID(inventoryContentID).Class("inv-body-content").Top(ui.Px(-float64(a.bodyOffset))).Children(
					a.pageHeader(),
					a.kpiPanel(),
					ui.Div().Class("inv-grid").Children(
						ui.Div().Class("inv-left").Children(a.stockPanel(), a.categoryPanel()),
						ui.Div().Class("inv-right").Children(a.binsPanel(), a.inboundPanel()),
					),
				),
			),
			a.poDialog(),
			a.exportToast(),
		)
}
