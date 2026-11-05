package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// exportToast is the bottom-right export confirmation, mirroring the
// loglens toast: it renders nothing while closed.
func (a *app) exportToast() *ui.Element {
	return chadcn.Toast(chadcn.ToastOptions{
		Open:        a.toastOpen,
		Title:       a.toastTitle,
		Description: a.toastBody,
		CloseAction: "toast:close",
	})
}

// poDialog builds the new-purchase-order modal. The host owns a.poOpen and the
// po:close/po:create actions; CloseAction covers the backdrop and the header
// close control. With no trigger supplied, Dialog returns nil while closed, so
// the caller can mount the result directly.
func (a *app) poDialog() *ui.Element {
	cancelBtn := chadcn.Button("Cancel", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "po:close")
	// Class replaces the attribute, so repeat the button classes alongside
	// inv-primary instead of appending to them.
	createBtn := chadcn.Button("Create PO", chadcn.ButtonOptions{
		Variant: chadcn.ButtonFilled,
		Size:    chadcn.ButtonSmall,
	}).Class("of-button of-button--filled of-button--small inv-primary").
		Attr("data-action", "po:create")
	return chadcn.Dialog(chadcn.DialogOptions{
		ID:          "po-dialog",
		Title:       "New purchase order",
		Description: "Create an inbound order — it joins the inbound list on arrival.",
		CloseAction: "po:close",
		Open:        a.poOpen,
		Footer:      []*ui.Element{cancelBtn, createBtn},
	},
		poDialogRow("Vendor", "Pacific Bean · PO-10483 (next)"),
		poDialogRow("Items", "1,680 units · Coffee & tea"),
		poDialogRow("ETA", "Today · Dock 3"),
		ui.Text("Draft — nothing is sent until it is confirmed.").Class("app-muted"),
	)
}

// poDialogRow is one label/value line in the purchase-order dialog body.
func poDialogRow(label, value string) *ui.Element {
	return ui.Div().Class("app-row").Gap(ui.Px(8)).Children(
		ui.Text(label).FontWeight(ui.FontWeightBold),
		ui.Text(value),
	)
}
