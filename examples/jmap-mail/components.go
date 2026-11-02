package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
)

func (a *app) actionButton(label, action string, prominent bool) *ui.Element {
	options := components.ButtonOptions{Primary: prominent}
	if action == "compose" {
		options.Overrides = composeProperties()
	}
	return components.Button(label, options).Attr("data-action", action)
}

func settingsButton() *ui.Element {
	return components.Button("Settings", components.ButtonOptions{Icon: settingsIcon()}).ID("settings-open").Attr("data-action", "settings")
}
