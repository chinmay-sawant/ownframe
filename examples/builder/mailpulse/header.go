package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) pageHead() *ui.Element {
	rangeBtn := chadcn.Button("📅 "+a.rangeLabel+" ▾", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "range:cycle")
	exportBtn := chadcn.Button("⤓ Export", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	left := ui.Div().Class("mp-pagehead-text").Children(
		chadcn.H1("Engagement"),
		chadcn.Muted("How readers open, click and reply to every send."),
	)
	actions := ui.Div().Class("mp-pagehead-actions").Children(rangeBtn, exportBtn)
	return ui.Div().Class("mp-pagehead").Children(left, actions)
}
