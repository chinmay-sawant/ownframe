package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// orderIDBase keeps new simulated order ids in the same display range as the
// generated seed orders.
const orderIDBase = 1000

// onClick routes every terminal control. All state changes end in one
// refresh, which publishes a new data snapshot; no animation is driven here.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	action := box.Action

	switch {
	case action == "" || action == "toast-close":
		a.clearToast()
	case action == "nav-back":
		a.moveHistory(-1)
	case action == "nav-forward":
		a.moveHistory(1)
	case strings.HasPrefix(action, "nav-"):
		a.setScreen(strings.TrimPrefix(action, "nav-"))
	case strings.HasPrefix(action, "watch-filter-"):
		a.watchFilter = strings.TrimPrefix(action, "watch-filter-")
		a.clearToast()
	case action == "watch-add":
		a.showToast("info", "Watchlist", "Symbol search is not part of this demo. Pick a row to open its chart.")
	case action == "watch-more":
		a.showToast("info", "Watchlist", "Columns are fixed to the reference layout in this example.")
	case strings.HasPrefix(action, "star-"):
		a.toggleStar(strings.TrimPrefix(action, "star-"))
	case strings.HasPrefix(action, "watch-"):
		a.selectSymbol(strings.TrimPrefix(action, "watch-"))
	case strings.HasPrefix(action, "tf-"):
		a.setTimeframe(strings.TrimPrefix(action, "tf-"))
	case strings.HasPrefix(action, "chart-tab-"):
		a.setChartTab(strings.TrimPrefix(action, "chart-tab-"))
	case strings.HasPrefix(action, "orders-tab-"):
		a.setOrderTab(strings.TrimPrefix(action, "orders-tab-"))
	case action == "order-submit":
		a.placeOrder()
	case action == "order-reset":
		a.ticket = defaultTicket(a.demo, a.symbol)
		a.showToast("info", "Ticket reset", "The order ticket is back to its default state.")
	case action == "order-side-buy":
		a.ticket.Side = "Buy"
		a.ticket.SideBuy = true
	case action == "order-side-sell":
		a.ticket.Side = "Sell"
		a.ticket.SideBuy = false
	case action == "order-type-market":
		a.ticket.Type = "Market"
	case action == "order-type-limit":
		a.ticket.Type = "Limit"
	case strings.HasPrefix(action, "order-qty-"):
		a.setQuantity(strings.TrimPrefix(action, "order-qty-"))
	case action == "order-tif-day":
		a.ticket.TIF = "Day"
	case action == "order-tif-gtc":
		a.ticket.TIF = "GTC"
	case action == "orders-export" || action == "order-filter":
		a.showToast("info", "Blotter", "The blotter is a local preview; no export file is written.")
	case strings.HasPrefix(action, "order-"):
		a.showToast("info", "Order detail", "Order "+strings.TrimPrefix(action, "order-")+" is simulated in this demo.")
	case strings.HasPrefix(action, "markets-index-"):
		a.openSymbol(strings.TrimPrefix(action, "markets-index-"), "Index snapshot")
	case strings.HasPrefix(action, "markets-sector"):
		a.showToast("info", "Sector", "Sector drill-down is not wired; the heatmap uses the generated sample.")
	case strings.HasPrefix(action, "markets-"):
		a.showToast("info", "Markets", "This market card is a preview. Switch to Watchlist to open a chart.")
	case strings.HasPrefix(action, "mover-"):
		a.openSymbol(strings.TrimPrefix(action, "mover-"), "Opened mover")
	case strings.HasPrefix(action, "holding-"):
		a.openSymbol(strings.TrimPrefix(action, "holding-"), "Opened position")
	case strings.HasPrefix(action, "trending-"):
		a.openSymbol(strings.TrimPrefix(action, "trending-"), "Opened trending symbol")
	case strings.HasPrefix(action, "screener-row-"):
		a.openSymbol(strings.TrimPrefix(action, "screener-row-"), "Opened screen result")
	case strings.HasPrefix(action, "screener-preset-"):
		a.applyScreenerPreset(strings.TrimPrefix(action, "screener-preset-"))
	case strings.HasPrefix(action, "screener-filter-"):
		a.toggleScreenerFilter(strings.TrimPrefix(action, "screener-filter-"))
	case action == "screener-run":
		a.demo.Screener.Updated = a.demo.Clock
		a.showToast("success", "Screen complete", fmt.Sprintf("%d simulated matches returned.", a.demo.Screener.Matches))
	case action == "screener-save":
		a.showToast("info", "Screen saved", "Saved screen presets are local to this demo.")
	case strings.HasPrefix(action, "news-filter-"):
		a.newsFilter = strings.TrimPrefix(action, "news-filter-")
		if a.newsFilter == "all" || a.newsFilter == "markets" || a.newsFilter == "tech" || a.newsFilter == "macro" {
			a.clearToast()
		}
	case action == "news-open":
		a.showToast("info", "Headline", "Article reading is outside this example; the feed is synthetic.")
	case action == "news-alerts":
		a.showToast("info", "Alerts", "News alerts are simulated and never leave the machine.")
	case strings.HasPrefix(action, "toggle-setting-"):
		a.toggleSetting(strings.TrimPrefix(action, "toggle-setting-"))
	case strings.HasPrefix(action, "setting-choice-"):
		a.showToast("info", "Preference", "This choice is applied to the preview only.")
	case action == "settings-reset":
		a.resetSettings()
	case action == "settings-reset-ui":
		a.resetSettings()
	case strings.HasPrefix(action, "settings-cat-"):
		a.clearToast()
	case action == "settings-docs":
		a.showToast("info", "Documentation", "The example README covers the terminal, the data, and the CSS motion.")
	case action == "bell":
		a.notifications = 0
		a.showToast("success", "Inbox clear", "Three simulated price alerts were marked read.")
	case action == "refresh-data":
		a.demo.Screener.Updated = a.demo.Clock
		a.showToast("success", "Feed refreshed", "Simulated quotes, candles, and prints are current.")
	case action == "update":
		a.showToast("info", "Update available", "Meridian v0.0.3 is ready. This preview stays local.")
	case action == "search":
		a.showToast("info", "Search", "Use the watchlist, screener, or trending rail in this demo.")
	case action == "profile":
		a.showToast("info", a.demo.ProfileName, "Paper account · simulated balances and executions.")
	case action == "shortcuts":
		a.showToast("info", "Shortcuts", "F12 opens devtools. The full key map is on the Settings screen.")
	case action == "help":
		a.showToast("info", "Help", "Every screen is sample data. Nothing is sent or persisted.")
	default:
		a.showToast("info", "Preview", "This control is visual only in the terminal demo.")
	}

	a.refresh()

	return nil
}

// toggleStar flips one watchlist star.
func (a *App) toggleStar(symbol string) {
	if symbol == "" {
		return
	}

	a.starred[symbol] = !a.starred[symbol]
}

// setTimeframe selects a chart range.
func (a *App) setTimeframe(key string) {
	switch key {
	case "1D", "5D", "1M", "3M", "1Y", "ALL":
		a.timeframe = key
	}
}

// setChartTab selects the content under the quote header.
func (a *App) setChartTab(key string) {
	switch key {
	case "chart", "options", "news", "financial":
		a.chartTab = key
	}
}

// setOrderTab selects the blotter view.
func (a *App) setOrderTab(key string) {
	switch key {
	case "today", "past", "holdings":
		a.orderTab = key
	}
}

// setQuantity updates the ticket quantity and its estimate.
func (a *App) setQuantity(text string) {
	n, err := strconv.Atoi(text)
	if err != nil || n <= 0 {
		return
	}

	a.ticket.Qty = strconv.Itoa(n)
	a.updateEstimate()
}

// updateEstimate recomputes the ticket notional from the live quote.
func (a *App) updateEstimate() {
	n, err := strconv.Atoi(a.ticket.Qty)
	if err != nil || n <= 0 {
		return
	}

	price := a.ticketPrice()
	a.ticket.Estimated = "$" + strconv.FormatFloat(float64(n)*price, 'f', 2, 64)
}

// ticketPrice reads the ticket limit or the current quote price.
func (a *App) ticketPrice() float64 {
	data, ok := a.demo.Symbols[a.symbol]
	if !ok {
		return 0
	}

	raw := a.ticket.Limit
	if a.ticket.Type == "Market" || strings.TrimSpace(raw) == "" {
		raw = data.Quote.Price
	}

	raw = strings.ReplaceAll(raw, "$", "")
	raw = strings.ReplaceAll(raw, ",", "")
	price, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}

	return price
}

// placeOrder adds a simulated working order to the blotter.
func (a *App) placeOrder() {
	qty, err := strconv.Atoi(a.ticket.Qty)
	if err != nil || qty <= 0 {
		a.showToast("danger", "Order rejected", "Choose a quantity before submitting.")
		return
	}

	data, ok := a.demo.Symbols[a.symbol]
	if !ok {
		a.showToast("danger", "Order rejected", "The selected symbol has no live quote.")
		return
	}

	a.nextOrder++

	row := OrderRow{
		Index:   0,
		ID:      fmt.Sprintf("ORD-%04d", orderIDBase+a.nextOrder),
		Time:    a.demo.Clock,
		Symbol:  data.Quote.Symbol,
		Name:    data.Quote.Name,
		Side:    a.ticket.Side,
		Status:  "Working",
		Type:    a.ticket.Type,
		Qty:     a.ticket.Qty,
		Filled:  "0",
		Limit:   a.ticket.Limit,
		Pending: true,
		Sell:    a.ticket.Side == "Sell",
	}

	if row.Type == "Market" {
		row.Limit = data.Quote.Price
	}

	a.demo.Orders = append([]OrderRow{row}, a.demo.Orders...)
	for i := range a.demo.Orders {
		a.demo.Orders[i].Index = i
	}

	a.ticket.Status = "Submitted"
	a.orderTab = "today"
	a.showToast("success", "Order submitted", fmt.Sprintf("%s %s %s at %s · working.", a.ticket.Side, a.ticket.Qty, data.Quote.Symbol, row.Limit))
}

// openSymbol selects a symbol and moves to the watchlist.
func (a *App) openSymbol(symbol, title string) {
	if _, ok := a.demo.Symbols[symbol]; !ok {
		a.showToast("info", title, "That symbol has no generated quote in this demo.")
		return
	}

	a.selectSymbol(symbol)
	a.setScreen("watchlist")
}

// applyScreenerPreset marks one preset active.
func (a *App) applyScreenerPreset(key string) {
	for i := range a.demo.Screener.Presets {
		a.demo.Screener.Presets[i].Active = a.demo.Screener.Presets[i].Key == key
	}

	a.demo.Screener.Updated = a.demo.Clock
	a.showToast("info", "Preset applied", "The result set is a seeded sample; filters stay editable.")
}

// toggleScreenerFilter flips one filter chip.
func (a *App) toggleScreenerFilter(key string) {
	for i := range a.demo.Screener.Filters {
		if a.demo.Screener.Filters[i].Key == key {
			a.demo.Screener.Filters[i].Active = !a.demo.Screener.Filters[i].Active
		}
	}
}

// toggleSetting flips one settings row and applies the devtools switch.
func (a *App) toggleSetting(key string) {
	if key == "" {
		return
	}

	on := !a.settings[key]
	a.settings[key] = on

	switch key {
	case "devtools":
		a.page.Window().SetDevTools(on)
		a.showToast("info", "Developer tools", "The dock is now "+map[bool]string{true: "open", false: "closed"}[on]+".")
	default:
		a.showToast("success", "Preference saved", "This demo applies the setting for the current session.")
	}
}

// resetSettings restores the generated defaults.
func (a *App) resetSettings() {
	for _, section := range a.demo.Settings.Sections {
		for _, row := range section.Rows {
			a.settings[row.Key] = row.On
		}
	}

	a.showToast("info", "Settings reset", "Preferences are back to the demo defaults.")
}
