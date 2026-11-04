package main

import (
	"github.com/chinmay-sawant/ownframe/examples/builder/mailpulse/components"
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) kpiRow() *ui.Element {
	items := kpisFor(a.rangeDays())
	cards := make([]*ui.Element, 0, len(items))
	for _, item := range items {
		cards = append(cards, components.KPICard(item.label, item.value, item.delta, item.footnote, item.tint, item.up))
	}
	return ui.Div().Class("mp-kpis").Children(cards...)
}
