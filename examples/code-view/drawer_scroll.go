package main

const (
	// drawerResizeHalfW centers the 6 px drawer resize handle on the
	// drawer boundary: the handle's left edge sits half its width left
	// of the boundary.
	drawerResizeHalfW = 3
)

// drawerWidth returns the file/source-control drawer width in 1x pixels:
// sideWidth plus drawerExtra, clamped to [sideWidth, drawerMaxW] (268..348,
// about 30% wider), like the VS Code sidebar resize.
func (a *App) drawerWidth() int {
	extra := a.drawerExtra
	if extra < 0 {
		extra = 0
	}

	if extra > drawerMaxExtra {
		extra = drawerMaxExtra
	}

	return sideWidth + extra
}

// clampDrawerExtra pins a restored drawer widening to [0, drawerMaxExtra].
func clampDrawerExtra(extra int) int {
	if extra < 0 {
		return 0
	}

	if extra > drawerMaxExtra {
		return drawerMaxExtra
	}

	return extra
}

// drawerVisiblePx is the drawer width in page pixels.
func (a *App) drawerVisiblePx() int {
	return a.zx(a.drawerWidth())
}

// chromeLeftPx is the editor's left edge: activity bar plus drawer.
func (a *App) chromeLeftPx() int {
	return a.zx(activityW) + a.drawerVisiblePx()
}

// paneRailLeftPx is the vertical rail's left edge at the drawer right edge.
func (a *App) paneRailLeftPx() int {
	left := a.chromeLeftPx() - a.zx(paneScrollW)
	if left < 0 {
		return 0
	}

	return left
}

// setDrawerExtra pins the drawer widening to [0, drawerMaxExtra].
func (a *App) setDrawerExtra(extra int) {
	if extra < 0 {
		extra = 0
	}

	if extra > drawerMaxExtra {
		extra = drawerMaxExtra
	}

	if extra == a.drawerExtra {
		return
	}

	a.drawerExtra = extra
	a.refresh()
}

// fillDrawer publishes the drawer width and the resize handle on its
// boundary. The tree, changes, and history lists paint at the visible
// width, so widening the drawer shows more of each file row and card.
func (a *App) fillDrawer(v *View) {
	visible := a.drawerVisiblePx()
	v.SideW = visible
	v.DrawerResizeLeft = a.zx(activityW) + visible - a.zx(drawerResizeHalfW)
	v.DrawerResizeTop = a.zx(chromeTop)
	v.DrawerResizeH = a.viewportH - a.zx(chromeTop) - a.zx(chromeBottom)
}
