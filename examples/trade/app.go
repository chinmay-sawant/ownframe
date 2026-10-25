package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

const (
	defaultWidth         = 1600
	defaultHeight        = 900
	minWidth             = 1180
	minHeight            = 700
	defaultSymbol        = "NVDA"
	defaultNotifications = 3
	blotterPageRows      = 10
)

// App owns the terminal page and the UI state that is not part of the sample
// dataset. One SetData call publishes a change; animations are CSS only.
type App struct {
	page *ownframe.Page
	demo *Demo
	view View

	screen      string
	symbol      string
	timeframe   string
	chartTab    string
	orderTab    string
	watchFilter string
	newsFilter  string

	starred       map[string]bool
	settings      map[string]bool
	ticket        Ticket
	toast         Toast
	notifications int

	navHistory []string
	navIndex   int
	nextOrder  int
}

// New builds the terminal, starts on the watchlist, and registers the click
// handler. Developer tools open by default for this example.
func New() (*App, error) {
	demo := newDemo()
	app := &App{
		demo:          demo,
		screen:        "watchlist",
		symbol:        defaultSymbol,
		timeframe:     "1D",
		chartTab:      "chart",
		orderTab:      "today",
		watchFilter:   "all",
		newsFilter:    "all",
		starred:       map[string]bool{},
		settings:      map[string]bool{},
		notifications: defaultNotifications,
		navHistory:    []string{"watchlist"},
	}

	for _, row := range demo.Watch {
		if row.Starred {
			app.starred[row.Symbol] = true
		}
	}

	for _, section := range demo.Settings.Sections {
		for _, row := range section.Rows {
			app.settings[row.Key] = row.On
		}
	}

	app.ticket = defaultTicket(demo, defaultSymbol)

	page, err := ownframe.New(ownframe.Config{
		DevTools:    true,
		RetainState: true,
		Title:       "Meridian — Trading Terminal",
		HTML:        buildHTML(),
		Width:       defaultWidth,
		Height:      defaultHeight,
		MinWidth:    minWidth,
		MinHeight:   minHeight,
	})
	if err != nil {
		return nil, err
	}

	app.page = page
	page.Handle(ownframe.Handlers{Click: app.onClick})
	app.refresh()

	return app, nil
}

// Page is the ownframe page the example opens.
func (a *App) Page() *ownframe.Page { return a.page }

// View returns the data the current frame prints. Tests use it.
func (a *App) View() View { return a.view }

// refresh publishes the current state to the page. The framework redraws after
// the click handler returns.
func (a *App) refresh() {
	a.view = a.buildView()
	a.page.SetData(a.view)
}

// buildView assembles the screen data from the generated dataset and the UI
// state. It never starts a timer or a tick callback.
func (a *App) buildView() View {
	demo := a.demo
	quote, ok := demo.Symbols[a.symbol]
	if !ok {
		for symbol, data := range demo.Symbols {
			a.symbol = symbol
			quote = data
			break
		}
	}

	view := View{
		Screen:          a.screen,
		Clock:           demo.Clock,
		Date:            demo.Date,
		UpdateAvailable: true,
		MarketOpen:      demo.MarketStatus == "Market open",
		MarketStatus:    demo.MarketStatus,
		SessionText:     "US session",
		Latency:         "12 ms",
		Notifications:   a.notifications,
		ProfileName:     demo.ProfileName,
		ProfileInitials: initials(demo.ProfileName),
		AvatarClass:     "avatar-a",
		Quote:           quote.Quote,
		ChartTab:        a.chartTab,
		ChartTabs:       []Tab{tab("chart", "Chart", a.chartTab), tab("options", "Options", a.chartTab), tab("news", "News", a.chartTab), tab("financial", "Financial", a.chartTab)},
		Timeframes:      timeframes(a.timeframe),
		Timeframe:       a.timeframe,
		Indicators:      indicators(),
		Options:         quote.Options,
		Financials:      quote.Financials,
		ChartNews:       demo.News.Items,
		RangeStats:      quote.Stats,
		KeyStats:        quote.Stats,
		OrderTab:        a.orderTab,
		OrderTabs:       orderTabs(demo, a.orderTab),
		OrderFilters: []SelectView{
			{Label: "Symbol type", Value: "All"},
			{Label: "Market", Value: "All markets"},
			{Label: "Order status", Value: "All"},
			{Label: "Order time", Value: "Today"},
		},
		Book:        quote.Book,
		Trades:      quote.Trades,
		BBO:         quote.BBO,
		Donut:       quote.Donut,
		DonutLegend: quote.DonutLegend,
		QuoteStats:  quote.Stats,
		DepthRatio:  "1.42",
		Indices:     demo.Indices,
		Tickers:     demo.Tickers,
		Markets:     demo.Markets,
		Portfolio:   demo.Portfolio,
		News:        demo.News,
		Screener:    demo.Screener,
		Settings:    a.settingsView(),
		Toast:       a.toast,
	}

	view.Watch = a.watchRows()
	view.WatchFilters = watchFilters(a.watchFilter)
	view.WatchCount = len(view.Watch)
	view.WatchTotal = len(demo.Watch)
	view.WatchFilter = a.watchFilter

	view.Chart = prepareChart(quote.Quote, quote.Candles, a.timeframe)
	view.Orders = a.orderRows()
	view.Holdings = demo.Holdings
	view.OrderStats = orderStats(demo.Orders)
	view.Ticket = a.ticket
	view.Ticket.Symbol = quote.Quote.Symbol
	view.Ticket.BuyingPower = demo.Portfolio.BuyingPower

	if a.ticket.Estimated == "" {
		view.Ticket.Estimated = quote.Quote.Price
	}

	view.News = a.newsView()
	view.Portfolio.DayUp = !strings.HasPrefix(view.Portfolio.DayPL, "-")
	view.Portfolio.TotalUp = !strings.HasPrefix(view.Portfolio.TotalPL, "-")

	return view
}

// watchRows applies the active watchlist filter and the star state.
func (a *App) watchRows() []WatchRow {
	rows := make([]WatchRow, 0, len(a.demo.Watch))

	for _, row := range a.demo.Watch {
		row.Starred = a.starred[row.Symbol]
		row.Active = row.Symbol == a.symbol

		if !watchMatch(row, a.watchFilter) {
			continue
		}

		row.Index = len(rows)
		rows = append(rows, row)
	}

	return rows
}

// orderRows returns the rows the active blotter tab prints.
func (a *App) orderRows() []OrderRow {
	orders := a.demo.Orders

	switch a.orderTab {
	case "past":
		if len(orders) > blotterPageRows {
			return orders[blotterPageRows:]
		}
		return orders
	case "holdings":
		return nil
	default:
		return orders
	}
}

// settingsView copies the settings sections and applies the live toggles.
func (a *App) settingsView() SettingsData {
	out := a.demo.Settings
	out.Sections = make([]SettingSection, len(a.demo.Settings.Sections))

	for i, section := range a.demo.Settings.Sections {
		rows := make([]SettingRow, len(section.Rows))
		copy(rows, section.Rows)

		for j := range rows {
			if on, ok := a.settings[rows[j].Key]; ok {
				rows[j].On = on
			}
		}

		section.Rows = rows
		out.Sections[i] = section
	}

	return out
}

// newsView applies the source filter and the source chips.
func (a *App) newsView() NewsData {
	out := a.demo.News
	out.Sources = newsSources(a.newsFilter)

	if a.newsFilter == "all" {
		return out
	}

	items := make([]NewsItem, 0, len(out.Items))

	for _, item := range out.Items {
		if item.Tag == a.newsFilter {
			items = append(items, item)
		}
	}

	out.Items = items

	return out
}

// showToast publishes a confirmation and keeps it until the user closes it.
func (a *App) showToast(kind, title, message string) {
	a.toast = Toast{Show: true, Up: kind != "danger", Kind: kind, Title: title, Message: message}
}

// clearToast hides the confirmation.
func (a *App) clearToast() {
	a.toast = Toast{}
}

// selectSymbol switches the quote, chart, book, and ticket to one instrument.
func (a *App) selectSymbol(symbol string) {
	if symbol == "" {
		return
	}

	if _, ok := a.demo.Symbols[symbol]; !ok {
		return
	}

	a.symbol = symbol
	a.ticket = defaultTicket(a.demo, symbol)
}

// setScreen switches the rail destination and records it for Back/Forward.
func (a *App) setScreen(screen string) {
	if !knownScreen(screen) {
		return
	}

	if screen != a.screen {
		if a.navIndex < len(a.navHistory)-1 {
			a.navHistory = a.navHistory[:a.navIndex+1]
		}

		a.navHistory = append(a.navHistory, screen)
		a.navIndex = len(a.navHistory) - 1
	}

	a.screen = screen
	a.clearToast()
}

// moveHistory walks the rail history. It reports false at either end.
func (a *App) moveHistory(delta int) bool {
	next := a.navIndex + delta
	if next < 0 || next >= len(a.navHistory) {
		return false
	}

	a.navIndex = next
	a.screen = a.navHistory[next]
	a.clearToast()

	return true
}

// knownScreen reports whether a rail destination exists.
func knownScreen(screen string) bool {
	switch screen {
	case "watchlist", "markets", "portfolio", "orders", "news", "screener", "settings":
		return true
	default:
		return false
	}
}

// defaultTicket builds the ticket state for a new symbol.
func defaultTicket(demo *Demo, symbol string) Ticket {
	data, ok := demo.Symbols[symbol]
	if !ok {
		return Ticket{Side: "Buy", Type: "Market", Qty: "100", TIF: "Day", SideBuy: true}
	}

	return Ticket{
		Symbol:  symbol,
		Side:    "Buy",
		Type:    "Market",
		Qty:     "100",
		Limit:   data.Quote.Price,
		TIF:     "Day",
		Status:  "Ready",
		SideBuy: true,
	}
}

// tab returns one tab with its active state.
func tab(key, label, active string) Tab {
	return Tab{Key: key, Label: label, Active: key == active}
}

// timeframes is the chart range strip.
func timeframes(active string) []Chip {
	keys := []string{"1D", "5D", "1M", "3M", "1Y", "ALL"}
	labels := []string{"1D", "5D", "1M", "3M", "1Y", "All"}
	out := make([]Chip, 0, len(keys))

	for i, key := range keys {
		out = append(out, Chip{Key: key, Label: labels[i], Active: key == active})
	}

	return out
}

// indicators is the chart indicator strip.
func indicators() []Chip {
	return []Chip{
		{Key: "sma", Label: "SMA", Active: true},
		{Key: "ema", Label: "EMA", Active: true},
		{Key: "vwap", Label: "VWAP", Active: true},
		{Key: "vol", Label: "Volume", Active: true},
		{Key: "rsi", Label: "RSI 54.2", Active: false},
	}
}

// watchFilters is the watchlist filter strip.
func watchFilters(active string) []FilterChip {
	keys := []string{"all", "us", "hk", "ev", "starred"}
	labels := []string{"ALL", "US", "HK", "EV", "STARRED"}
	out := make([]FilterChip, 0, len(keys))

	for i, key := range keys {
		out = append(out, FilterChip{Key: key, Label: labels[i], Active: key == active})
	}

	return out
}

// watchMatch reports whether a row passes the watchlist filter.
func watchMatch(row WatchRow, filter string) bool {
	switch filter {
	case "us":
		return row.Market == "US"
	case "hk":
		return row.Market == "HK"
	case "ev":
		return row.Market == "EV"
	case "starred":
		return row.Starred
	default:
		return true
	}
}

// newsSources is the news source strip.
func newsSources(active string) []Chip {
	keys := []string{"all", "markets", "tech", "macro"}
	labels := []string{"All", "Markets", "Tech", "Macro"}
	out := make([]Chip, 0, len(keys))

	for i, key := range keys {
		out = append(out, Chip{Key: key, Label: labels[i], Active: key == active})
	}

	return out
}

// orderTabs is the blotter tab strip with live counts.
func orderTabs(demo *Demo, active string) []Tab {
	working := 0
	cancelled := 0

	for _, order := range demo.Orders {
		switch {
		case order.Pending:
			working++
		case order.Cancelled:
			cancelled++
		}
	}

	return []Tab{
		{Key: "today", Label: "Today's Orders", Active: active == "today", Count: len(demo.Orders)},
		{Key: "past", Label: "Past Orders", Active: active == "past", Count: cancelled + working},
		{Key: "holdings", Label: "Holdings", Active: active == "holdings", Count: len(demo.Holdings)},
	}
}

// orderStats summarises the generated order book.
func orderStats(orders []OrderRow) OrderStats {
	filled, working, cancelled := 0, 0, 0

	for _, order := range orders {
		switch {
		case order.Pending:
			working++
		case order.Cancelled:
			cancelled++
		case order.Done:
			filled++
		}
	}

	return OrderStats{
		Filled:    strconv.Itoa(filled),
		Working:   strconv.Itoa(working),
		Cancelled: strconv.Itoa(cancelled),
		Notional:  "$2.84M",
	}
}

// initials returns up to two letters for the avatar.
func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "M"
	}

	first := []rune(parts[0])
	out := string(first[:1])

	if len(parts) > 1 {
		second := []rune(parts[1])
		out += string(second[:1])
	}

	return strings.ToUpper(out)
}
