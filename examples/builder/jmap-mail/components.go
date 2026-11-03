package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) actionButton(label, action string, prominent bool) *ui.Element {
	variant := chadcn.ButtonTonal
	if prominent {
		variant = chadcn.ButtonFilled
	}
	options := chadcn.ButtonOptions{Variant: variant}
	if action == "compose" {
		options.Overrides = composeProperties()
	}
	return chadcn.Button(label, options).Attr("data-action", action)
}

func settingsButton() *ui.Element {
	return chadcn.Button("Settings", chadcn.ButtonOptions{Icon: settingsIcon()}).ID("settings-open").Attr("data-action", "settings")
}
