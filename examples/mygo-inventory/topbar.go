// Top bar and page header: brand, nav, search, alerts and account,
// period segment, category filter, and the New PO dialog.
package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// topbar is the inventory chrome: brand, product nav, search, alerts and
// account on one row.
func (s *heatApp) topbar(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(10).Padding(12).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(28, 28).Radius(8).Background(ui.Hex("#1d5c3f")).
			AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Text(c, "N").Bold().TextColor(ui.Hex("#fafaf7"))
		})
		ui.Text(c, "Northwind Goods").Bold()
		for _, item := range []string{"Inventory", "Purchase orders", "Receiving", "Picking", "Suppliers", "Reports"} {
			if item == "Inventory" {
				ui.Box(c).Radius(12).Background(ui.Hex("#e8f0e8")).Padding(6).Children(func() {
					ui.Text(c, item).FontSize(13).Bold().TextColor(ui.Hex("#1d5c3f"))
				})
				continue
			}
			ui.Text(c, item).FontSize(13).TextColor(t.TextMuted)
		}
		ui.Box(c).Grow(1)
		s.searchBox(c)
		s.alertsButton(c)
		s.accountButton(c)
	})
	ui.Divider(c)
}

// searchBox is the topbar search: a SearchField plus the filtered SKU and
// PO results, and the last picked item.
func (s *heatApp) searchBox(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(260).Children(func() {
		ui.SearchField(c, &s.query)
		q := strings.ToLower(strings.TrimSpace(s.query))
		if q == "" {
			if s.picked != "" {
				ui.Text(c, "Picked: "+s.picked).FontSize(11).TextColor(t.TextMuted)
			}
			return
		}
		shown := 0
		for _, p := range products {
			if shown >= 5 {
				break
			}
			if !strings.Contains(strings.ToLower(p.name+" "+p.code+" "+p.category), q) {
				continue
			}
			shown++
			p := p
			if ui.Button(c, p.name+" · "+p.status).Clicked() {
				s.picked = p.code
				s.query = ""
			}
		}
		for _, po := range purchaseOrders {
			if !strings.Contains(strings.ToLower(po.vendor+" "+po.poNo), q) {
				continue
			}
			po := po
			if ui.Button(c, po.poNo+" · "+po.vendor).Clicked() {
				s.picked = po.poNo
				s.query = ""
			}
		}
		if shown == 0 {
			ui.Text(c, "No matches").FontSize(12).TextColor(t.TextMuted)
		}
	})
}

// alertsButton is the stock-alert inbox: a button with the unread count
// opening a popover with the two zero-on-hand rows.
func (s *heatApp) alertsButton(c *ui.Context) {
	t := c.Theme()
	anchor := ui.Button(c, "Alerts (2)")
	if anchor.Clicked() {
		s.notifOpen = !s.notifOpen
	}
	ui.Popover(c, anchor, &s.notifOpen, func() {
		ui.Column(c).Gap(6).Padding(8).Children(func() {
			ui.Text(c, "Stock alerts").Bold()
			ui.Text(c, "Cold brew filter packs: 0 on hand · 1,200 inbound today").FontSize(12)
			ui.Text(c, "Ceremonial matcha: 0 on hand · reorder advised").FontSize(12)
			ui.Text(c, "Updated just now").FontSize(11).TextColor(t.TextMuted)
		})
	})
}

// accountButton is the session menu: Profile and a dangerous Sign out,
// both closing the popover.
func (s *heatApp) accountButton(c *ui.Context) {
	anchor := ui.Button(c, "AK")
	if anchor.Clicked() {
		s.userOpen = !s.userOpen
	}
	ui.Popover(c, anchor, &s.userOpen, func() {
		ui.Column(c).Gap(4).Padding(8).Children(func() {
			ui.Text(c, "Account").FontSize(12)
			if ui.Button(c, "Profile").Clicked() {
				s.userOpen = false
			}
			if ui.Button(c, "Sign out").Clicked() {
				s.userOpen = false
			}
		})
	})
}

// pageHeader is the "Inventory health" title row with the period segment,
// the category filter, and the New PO button.
func (s *heatApp) pageHeader(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
		ui.Column(c).Gap(2).Grow(1).Children(func() {
			ui.Text(c, "Inventory health").FontSize(20).Bold()
			ui.Text(c, "Warehouse 1 · Reno, NV · synced 4 min ago from NetSuite").FontSize(11).TextColor(t.TextMuted)
		})
		for _, period := range []string{"30 days", "90 days", "12 months"} {
			period := period
			b := ui.Button(c, period)
			if period == s.period {
				b = b.Background(ui.Hex("#1d5c3f")).TextColor(ui.Hex("#fafaf7"))
			}
			if b.Clicked() {
				s.period = period
			}
		}
		ui.Select(c, &s.category, []string{
			"All categories", "Coffee & tea", "Dairy alternatives", "Packaging", "Janitorial", "Syrups",
		})
		if ui.PrimaryButton(c, "+ New PO").Clicked() {
			s.poOpen = true
		}
	})
	ui.Modal(c, &s.poOpen, func() {
		ui.Column(c).Gap(8).Padding(16).Children(func() {
			ui.Text(c, "New purchase order").FontSize(16).Bold()
			ui.Text(c, "Create an inbound order — it joins the inbound list on arrival.").FontSize(12).TextColor(t.TextMuted)
			ui.Text(c, "Vendor: Pacific Bean · PO-10483 (next)").FontSize(13)
			ui.Text(c, "Items: 1,680 units · Coffee & tea").FontSize(13)
			ui.Text(c, "ETA: Today · Dock 3").FontSize(13)
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					s.poOpen = false
				}
				if ui.PrimaryButton(c, "Create PO").Clicked() {
					s.poOpen = false
				}
			})
		})
	})
}
