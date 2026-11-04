package main

import (
	"fmt"
	"strings"
)

// handleCatalog handles catalog: actions from the ChadCN catalog tab. The
// FilterBar pills and SidebarSection nav items report a preview in the notice
// text. The Checkbox, Switch, and Slider entries are live: they toggle or step
// builderApp catalog state and re-render. Tabs and Select are live through
// their OnSelect/OnOpenChange callbacks in screen_catalog.go, so they need no
// action here. Chart, AreaChart, HexGrid, StatCard, and StatusSegments are
// display-only and stay static.
func (a *builderApp) handleCatalog(action string) bool {
	name, ok := strings.CutPrefix(action, "catalog:")
	if !ok {
		return false
	}
	switch name {
	case "filter:daily", "filter:weekly", "filter:monthly":
		a.notice = "Catalog filter preview: " + strings.TrimPrefix(name, "filter:") + "."
		return true
	case "nav:dashboard", "nav:work-orders", "nav:maintenance":
		a.notice = "Catalog preview: " + strings.TrimPrefix(name, "nav:") + "."
		return true
	case "check":
		a.catalogCheck = !a.catalogCheck
		a.notice = fmt.Sprintf("Catalog checkbox %s.", toggledWord(a.catalogCheck))
		return true
	case "switch":
		a.catalogSwitch = !a.catalogSwitch
		a.notice = fmt.Sprintf("Catalog switch %s.", toggledWord(a.catalogSwitch))
		return true
	case "slider:inc":
		if a.catalogSlider+catalogSliderStep > catalogSliderMax {
			a.catalogSlider = catalogSliderMax
		} else {
			a.catalogSlider += catalogSliderStep
		}
		a.notice = fmt.Sprintf("Catalog slider %.0f.", a.catalogSlider)
		return true
	case "slider:dec":
		if a.catalogSlider-catalogSliderStep < 0 {
			a.catalogSlider = 0
		} else {
			a.catalogSlider -= catalogSliderStep
		}
		a.notice = fmt.Sprintf("Catalog slider %.0f.", a.catalogSlider)
		return true
	}
	return false
}
