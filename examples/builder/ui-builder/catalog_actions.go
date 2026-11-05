package main

import (
	"fmt"
	"strings"
)

// handleCatalog handles catalog: actions from the ChadCN catalog tab. Every
// card carries at least one data-action tap target so clicks always report in
// the notice text. Controlled overlays (dialog, popover, date picker,
// combobox, dropdown, sheet, drawer, alert dialog, menubar, navigation menu)
// toggle through On* callbacks in screen_catalog.go; this handler covers the
// data-action buttons, including toggles for hover, tooltip, context, sidebar,
// toast, choice, and the generic display-card taps.
func (a *builderApp) handleCatalog(action string) bool {
	name, ok := strings.CutPrefix(action, "catalog:")
	if !ok {
		return false
	}
	switch name {
	case "filter:daily", "filter:weekly", "filter:monthly":
		a.catalogFilter = strings.TrimPrefix(name, "filter:")
		a.notice = "Catalog filter preview: " + a.catalogFilter + "."
		return true
	case "nav:dashboard", "nav:work-orders", "nav:maintenance":
		a.notice = "Catalog preview: " + strings.TrimPrefix(name, "nav:") + "."
		return true
	case "crumb:home", "crumb:library":
		a.notice = "Catalog crumb: " + strings.TrimPrefix(name, "crumb:") + "."
		return true
	case "check":
		a.catalogCheck = !a.catalogCheck
		a.notice = fmt.Sprintf("Catalog checkbox %s.", toggledWord(a.catalogCheck))
		return true
	case "switch":
		a.catalogSwitch = !a.catalogSwitch
		a.notice = fmt.Sprintf("Catalog switch %s.", toggledWord(a.catalogSwitch))
		return true
	case "choice:toggle":
		a.catalogChoiceSelected = !a.catalogChoiceSelected
		a.notice = fmt.Sprintf("Catalog choice %s.", toggledWord(a.catalogChoiceSelected))
		return true
	case "toggle:tap":
		a.catalogTogglePressed = !a.catalogTogglePressed
		a.notice = fmt.Sprintf("Catalog toggle %s.", toggledWord(a.catalogTogglePressed))
		return true
	case "hover:toggle":
		a.catalogHoverOpen = !a.catalogHoverOpen
		a.notice = fmt.Sprintf("Catalog hover open=%v.", a.catalogHoverOpen)
		return true
	case "tooltip:toggle":
		a.catalogTooltipOpen = !a.catalogTooltipOpen
		a.notice = fmt.Sprintf("Catalog tooltip open=%v.", a.catalogTooltipOpen)
		return true
	case "context:toggle":
		a.catalogContextOpen = !a.catalogContextOpen
		if a.catalogContextOpen {
			a.notice = "Catalog context menu open."
		} else {
			a.notice = "Catalog context menu closed."
		}
		return true
	case "sidebar:toggle":
		a.catalogSidebarOpen = !a.catalogSidebarOpen
		a.notice = fmt.Sprintf("Catalog sidebar collapsed=%v.", !a.catalogSidebarOpen)
		return true
	case "toast:show":
		a.catalogToastOpen = true
		a.notice = "Catalog toast open."
		return true
	case "toast:close", "toast:undo":
		a.catalogToastOpen = false
		if name == "toast:undo" {
			a.notice = "Catalog toast undone."
		} else {
			a.notice = "Catalog toast closed."
		}
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
	case "progress:inc":
		if a.catalogProgress+catalogProgressStep > catalogProgressMax {
			a.catalogProgress = catalogProgressMax
		} else {
			a.catalogProgress += catalogProgressStep
		}
		a.notice = fmt.Sprintf("Catalog progress %d%%.", a.catalogProgress)
		return true
	case "progress:dec":
		if a.catalogProgress-catalogProgressStep < 0 {
			a.catalogProgress = 0
		} else {
			a.catalogProgress -= catalogProgressStep
		}
		a.notice = fmt.Sprintf("Catalog progress %d%%.", a.catalogProgress)
		return true
	case "alert:view":
		a.notice = "Catalog alert viewed."
		return true
	case "attachment:dismiss":
		a.notice = "Catalog attachment dismissed."
		return true
	case "button:tap":
		a.notice = "Catalog button tapped."
		return true
	case "button-group:left", "button-group:right":
		a.notice = "Catalog button group: " + strings.TrimPrefix(name, "button-group:") + "."
		return true
	case "card:open":
		a.notice = "Catalog card opened."
		return true
	case "code:tap":
		a.notice = "Catalog code copied."
		return true
	case "empty:clear":
		a.notice = "Catalog empty cleared."
		return true
	case "item:open":
		a.notice = "Catalog item opened."
		return true
	case "message:reply":
		a.notice = "Catalog message reply."
		return true
	case "messages:load":
		a.notice = "Catalog messages loaded."
		return true
	case "resizable:reset":
		a.notice = "Catalog split reset."
		return true
	case "scroll:top":
		a.notice = "Catalog scroll top."
		return true
	case "scrollbar:nudge":
		a.notice = "Catalog scrollbar nudged."
		return true
	}
	if rest, ok := strings.CutPrefix(name, "tap:"); ok {
		a.notice = "Catalog preview: " + rest + "."
		return true
	}
	// Generic display-card taps: chart, area-chart, aspect, avatar, badge,
	// bubble, direction, hex, kbd, label, marker, notice, separator, skeleton,
	// spinner, stat, segments, table, trace, type.
	if strings.HasSuffix(name, ":tap") {
		a.notice = "Catalog preview: " + strings.TrimSuffix(name, ":tap") + "."
		return true
	}
	return false
}
