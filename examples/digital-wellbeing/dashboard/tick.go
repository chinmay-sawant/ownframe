package dashboard

import (
	"context"
	"encoding/base64"
)

func (a *App) apply(ctx context.Context) error {
	a.mu.Lock()
	access, accessSet := a.access, a.accessSet
	apps, appsSet := a.apps, a.appsSet
	usage, usageSet := a.usage, a.usageSet
	restore, restoreSet := a.restore, a.restoreSet
	back := a.backRequest
	a.accessSet, a.appsSet, a.usageSet, a.restoreSet, a.backRequest = false, false, false, false, false
	a.mu.Unlock()
	changed := false
	if accessSet && a.view.Access != access {
		a.view.Access, changed = access, true
	}
	if appsSet {
		for i := range apps {
			if data, err := base64.StdEncoding.DecodeString(apps[i].Icon); err == nil && len(data) > 0 {
				apps[i].IconSource = "appicon-" + apps[i].Package
				a.page.SetImage(apps[i].IconSource, data)
			}
			apps[i].Icon = ""
		}
		a.view.Apps, changed = apps, true
	}
	if restoreSet && a.view.Selected == "" && restore != "" {
		a.view.Selected, changed = restore, true
		a.clearUsage()
		a.setSelected(restore)
	}
	if back && a.applyBack() {
		changed = true
	}
	if usageSet && usage.Package == a.view.Selected && usage.WeekOffset == a.WeekOffset() {
		a.applyUsage(usage)
		changed = true
	}
	if !changed {
		return nil
	}
	a.refreshView()
	a.page.SetData(&a.view)
	return a.page.Redraw(ctx)
}
