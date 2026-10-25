//nolint:mnd,gosec // The sample portfolio, news, screener, and settings fixtures are literal display content.
package main

import "math/rand/v2"

const (
	dataPerformancePoints = "0,212 138,198 275,184 413,190 551,158 688,144 826,120 963,98 1101,74 1239,48"
	dataBenchmarkPoints   = "0,226 138,218 275,212 413,204 551,196 688,188 826,174 963,162 1101,146 1239,132"
)

// dataBuildPortfolio assembles the account screen from the generated holdings.
func dataBuildPortfolio(holdings []Holding) PortfolioData {
	allocation := []AllocationSlice{
		{Label: "Semiconductors", Value: "$1.87M", Pct: "38.4%", Color: "#3ddc84", Weight: 38.4},
		{Label: "Software & cloud", Value: "$1.11M", Pct: "22.8%", Color: "#64a8ff", Weight: 22.8},
		{Label: "Consumer & EV", Value: "$0.69M", Pct: "14.2%", Color: "#c084fc", Weight: 14.2},
		{Label: "Financials", Value: "$0.52M", Pct: "10.6%", Color: "#f0b75a", Weight: 10.6},
		{Label: "Cash & equivalents", Value: "$0.39M", Pct: "8.0%", Color: "#5c6573", Weight: 8.0},
		{Label: "Other", Value: "$0.29M", Pct: "6.0%", Color: "#f0606a", Weight: 6.0},
	}

	return PortfolioData{
		TotalValue:  "$4,862,904.18",
		DayPL:       "+$38,204.12",
		DayPct:      "+0.79%",
		TotalPL:     "+$612,448.90",
		TotalPct:    "+14.41%",
		BuyingPower: "$1,204,800.00",
		Cash:        "$389,204.18",
		Invested:    "$4,473,700.00",
		Margin:      "$0.00",
		DayUp:       true,
		TotalUp:     true,
		Holdings:    holdings,
		Allocation:  allocation,
		Performance: PerformanceChart{
			Points:            dataPerformancePoints,
			BenchmarkPoints:   dataBenchmarkPoints,
			Segments:          chartSegments(dataPerformancePoints, "#3ddc84", 1.8),
			BenchmarkSegments: chartSegments(dataBenchmarkPoints, "#525c6b", 1.4),
			Grid: []GridLine{
				{X1: 0, Y1: 48, X2: 999, Y2: 48, Major: true},
				{X1: 0, Y1: 101, X2: 999, Y2: 101},
				{X1: 0, Y1: 154, X2: 999, Y2: 154, Major: true},
				{X1: 0, Y1: 207, X2: 999, Y2: 207},
			},
			YLabels: []ChartLabel{
				{Top: "8%", Left: "0%", Text: "+18%", Tone: "muted"},
				{Top: "34%", Left: "0%", Text: "+14%", Tone: "dim"},
				{Top: "58%", Left: "0%", Text: "+10%", Tone: "dim"},
				{Top: "84%", Left: "0%", Text: "+6%", Tone: "dim"},
			},
			XLabels: []ChartLabel{
				{Top: "94%", Left: "0%", Text: "Jan", Tone: "dim"},
				{Top: "94%", Left: "18%", Text: "Mar", Tone: "dim"},
				{Top: "94%", Left: "38%", Text: "May", Tone: "dim"},
				{Top: "94%", Left: "58%", Text: "Jul", Tone: "dim"},
				{Top: "94%", Left: "78%", Text: "Sep", Tone: "dim"},
				{Top: "94%", Left: "94%", Text: "Oct", Tone: "dim"},
			},
			Start:  "Jan 02",
			End:    "Oct 25",
			Return: "+14.41%",
			Up:     true,
		},
		Events: []PortfolioEvent{
			{Time: "09:31:04", Text: "Bought 100 NVDA at 182.94", Value: "-$18,294.00", Up: false},
			{Time: "09:36:41", Text: "Sold 33 AAPL at 246.51", Value: "+$8,134.83", Up: true},
			{Time: "10:02:18", Text: "Dividend · MSFT", Value: "+$714.00", Up: true},
			{Time: "10:18:55", Text: "Bought 10 ARM at 176.20", Value: "-$1,762.00", Up: false},
			{Time: "10:31:27", Text: "Realized P&L · TSLA", Value: "+$1,245.90", Up: true},
			{Time: "10:39:02", Text: "Fee · exchange", Value: "-$12.40", Up: false},
		},
		Stats: []StatRow{
			{Label: "Net liquidation", Value: "$4,862,904.18", Up: true},
			{Label: "Day trades left", Value: "3 of 4", Tone: "amber"},
			{Label: "Margin used", Value: "$0.00", Up: true},
			{Label: "Win rate", Value: "68.4%", Up: true},
			{Label: "Avg hold", Value: "11.2 days"},
			{Label: "Account beta", Value: "1.14"},
		},
	}
}

// dataBuildNews is the synthetic newsroom feed.
func dataBuildNews() NewsData {
	return NewsData{
		Items: []NewsItem{
			{Source: "Bloomberg", Time: "10:38", Title: "Nvidia extends gains as AI capex forecasts move higher", Summary: "Hyperscaler spending plans lifted semiconductor peers and kept the Nasdaq on course for a fourth straight advance.", Ticker: "NVDA", Tag: "tech", Hot: true, Up: true},
			{Source: "Reuters", Time: "10:24", Title: "Treasury yields ease after soft services print", Summary: "The 10-year moved lower as traders pared back the odds of another near-term hike.", Ticker: "TLT", Tag: "macro", Up: true},
			{Source: "CNBC", Time: "10:12", Title: "Apple suppliers point to a firmer phone build", Summary: "Channel checks suggest the December quarter is tracking ahead of the prior cycle.", Ticker: "AAPL", Tag: "tech", Hot: true, Up: true},
			{Source: "WSJ", Time: "09:58", Title: "Banks brace for a quieter fourth-quarter trading book", Summary: "Fixed-income desks expect volumes to normalise after a busy September.", Ticker: "JPM", Tag: "markets", Up: false},
			{Source: "FT", Time: "09:44", Title: "Oil slips as supply risk premium unwinds", Summary: "Brent gave back part of Monday's gain while products stayed firm.", Ticker: "CL", Tag: "macro", Up: false},
			{Source: "Bloomberg", Time: "09:31", Title: "Tesla deliveries draw a mixed read from analysts", Summary: "Price cuts supported volume but margin estimates were trimmed at two brokers.", Ticker: "TSLA", Tag: "markets", Hot: true, Up: false},
			{Source: "Reuters", Time: "09:18", Title: "Chip equipment orders stabilise in Asia", Summary: "The latest book-to-bill reading points to a shallower downturn than feared.", Ticker: "AMAT", Tag: "tech", Up: true},
			{Source: "CNBC", Time: "09:02", Title: "S&P 500 breadth improves for a third session", Summary: "More components closed above their 50-day average, a sign the rally is broadening.", Ticker: "SPY", Tag: "markets", Up: true},
			{Source: "WSJ", Time: "08:46", Title: "Retail sales beat leaves the consumer in focus", Summary: "Discretionary names led the pre-market move after a firmer headline number.", Ticker: "XLY", Tag: "macro", Up: true},
			{Source: "FT", Time: "08:22", Title: "Alibaba's cloud unit draws a fresh look", Summary: "Investors weighed a potential re-rating as margin commentary improved.", Ticker: "BABA", Tag: "tech", Up: true},
			{Source: "Bloomberg", Time: "08:04", Title: "Gold holds near a record as real yields fall", Summary: "Bullion funds recorded a second week of inflows.", Ticker: "GLD", Tag: "macro", Up: true},
			{Source: "Reuters", Time: "07:41", Title: "Futures point to a higher open on Wall Street", Summary: "Megacap technology led the pre-market tape while small caps lagged.", Ticker: "QQQ", Tag: "markets", Up: true},
		},
		Trending: []TrendingTicker{
			{Symbol: "NVDA", Name: "NVIDIA", Mentions: "18.4K", Pct: "+2.84%", Up: true},
			{Symbol: "TSLA", Name: "Tesla", Mentions: "12.1K", Pct: "-1.92%", Up: false},
			{Symbol: "AAPL", Name: "Apple", Mentions: "9.8K", Pct: "+0.71%", Up: true},
			{Symbol: "ARM", Name: "Arm Holdings", Mentions: "6.2K", Pct: "-2.35%", Up: false},
			{Symbol: "BABA", Name: "Alibaba", Mentions: "4.7K", Pct: "+1.08%", Up: true},
		},
		Sentiment: []SentimentSlice{
			{Label: "Bullish", Pct: "62%", Up: true},
			{Label: "Neutral", Pct: "23%", Up: true},
			{Label: "Bearish", Pct: "15%", Up: false},
		},
		Calendar: []CalendarEvent{
			{Time: "08:30", Country: "US", Event: "Retail sales m/m", Actual: "0.7%", Forecast: "0.3%", Prior: "0.1%", Impact: "High", High: true},
			{Time: "09:45", Country: "US", Event: "Services PMI", Actual: "54.2", Forecast: "53.8", Prior: "53.4", Impact: "Medium"},
			{Time: "10:00", Country: "EU", Event: "ECB comments", Actual: "—", Forecast: "—", Prior: "—", Impact: "Medium"},
			{Time: "11:30", Country: "US", Event: "10Y note auction", Actual: "—", Forecast: "—", Prior: "4.21%", Impact: "High", High: true},
			{Time: "14:00", Country: "US", Event: "Fed minutes", Actual: "—", Forecast: "—", Prior: "—", Impact: "High", High: true},
		},
		Filter:    "all",
		Gauge:     64,
		GaugeText: "Risk-on",
	}
}

// dataBuildScreener is the seeded result set for the screener screen.
func dataBuildScreener(symbols map[string]SymbolData) ScreenerData {
	rng := rand.New(rand.NewPCG(dataSeed+17, dataFakerSeed))
	names := []string{"NVDA", "AAPL", "TSLA", "MSFT", "AMD", "META", "AMZN", "GOOGL", "JPM", "GS", "ARM", "BABA"}
	sectors := []string{"Semiconductors", "Hardware", "Consumer & EV", "Software", "Semiconductors", "Internet", "Consumer", "Internet", "Financials", "Financials", "Semiconductors", "Internet"}
	signals := []string{"Strong Buy", "Buy", "Watch", "Buy", "Accumulate", "Buy", "Accumulate", "Buy", "Hold", "Buy", "Watch", "Accumulate"}
	pes := []string{"58.4", "34.9", "92.4", "36.8", "41.6", "28.4", "38.7", "25.3", "15.1", "17.4", "128.6", "18.4"}
	scores := []int{98, 91, 74, 93, 88, 84, 79, 87, 70, 82, 66, 77}
	rows := make([]ScreenerRow, 0, len(names))

	for index, symbol := range names {
		data, ok := symbols[symbol]
		if !ok {
			continue
		}

		rows = append(rows, ScreenerRow{
			Symbol:    symbol,
			Name:      data.Quote.Name,
			Sector:    sectors[index],
			Price:     data.Quote.Price,
			Change:    data.Quote.Change,
			Pct:       data.Quote.ChangePct,
			Volume:    data.Quote.Volume,
			MarketCap: data.Quote.MarketCap,
			PE:        pes[index],
			RSI:       dataFixed(48+float64(index)*1.9, 1),
			Signal:    signals[index],
			Up:        data.Quote.Up,
			Score:     scores[index],
			Spark:     dataSparkBars(rng, 10),
		})
	}

	return ScreenerData{
		Presets: []Chip{
			{Key: "momentum", Label: "Momentum leaders", Active: true},
			{Key: "quality", Label: "Quality growth", Active: false},
			{Key: "value", Label: "Value tilt", Active: false},
			{Key: "breakout", Label: "Breakouts", Active: false},
		},
		Filters: []ScreenerFilter{
			{Key: "market-cap", Label: "Large cap", Value: "$10B+", Active: true},
			{Key: "price-range", Label: "Price", Value: "$50 – $1,000", Active: true},
			{Key: "avg-volume", Label: "Avg volume", Value: "5M+", Active: true},
			{Key: "rsi", Label: "RSI", Value: "50 – 70", Active: true},
			{Key: "trend", Label: "Above 50DMA", Value: "Yes", Active: true},
			{Key: "sector", Label: "Sector", Value: "Any", Active: false},
			{Key: "region", Label: "Region", Value: "US", Active: true},
			{Key: "exchange", Label: "Exchange", Value: "Any", Active: false},
			{Key: "turnover", Label: "Turnover", Value: "$1M+", Active: true},
			{Key: "spread", Label: "Spread", Value: "< 5 bps", Active: false},
			{Key: "macd", Label: "MACD", Value: "Positive", Active: true},
			{Key: "volatility", Label: "Volatility", Value: "< 35", Active: true},
		},
		Rows:    rows,
		Matches: len(rows),
		Updated: "10:42:18 ET",
	}
}

// dataBuildSettings is the settings screen's generated preference set.
func dataBuildSettings() SettingsData {
	return SettingsData{
		Sections: []SettingSection{
			{
				Key:   "trading",
				Title: "Trading",
				Note:  "Order entry and confirmation behavior for this session.",
				Rows: []SettingRow{
					{Key: "confirm", Title: "Confirm before submitting", Detail: "Ask for a second click on market and limit orders.", Kind: "toggle", On: true},
					{Key: "flash", Title: "Flash price changes", Detail: "Highlight rows whose quote moved since the previous sample.", Kind: "toggle", On: true},
					{Key: "sound", Title: "Order sounds", Detail: "Play a short cue when an order reaches a terminal state.", Kind: "toggle", On: false},
					{Key: "extended", Title: "Extended hours", Detail: "Show pre-market and after-hours prints in Time & Sales.", Kind: "toggle", On: true},
				},
			},
			{
				Key:   "appearance",
				Title: "Appearance",
				Note:  "Density and display choices for the terminal shell.",
				Rows: []SettingRow{
					{Key: "density", Title: "Density", Detail: "Controls row height across watchlists and blotters.", Kind: "choice", Value: "Compact", On: true},
					{Key: "compact", Title: "Compact headers", Detail: "Reduce the chart and panel header height.", Kind: "toggle", On: true},
					{Key: "tape", Title: "Scrolling tape", Detail: "Show the market tape in the status bar.", Kind: "toggle", On: true},
					{Key: "glow", Title: "Live accents", Detail: "Animate live dots, scans, and price flashes.", Kind: "toggle", On: true},
				},
			},
			{
				Key:   "data",
				Title: "Data",
				Note:  "The demo runs entirely on the generated local dataset.",
				Rows: []SettingRow{
					{Key: "provider", Title: "Provider", Detail: "Simulated feed backed by gofakeit and a seeded price walk.", Kind: "choice", Value: "Simulated", On: true},
					{Key: "delayed", Title: "Delayed quotes", Detail: "Use the 15-minute synthetic snapshot instead of the live walk.", Kind: "toggle", On: false},
					{Key: "notifications", Title: "Price alerts", Detail: "Show the notification badge on the top bar.", Kind: "toggle", On: true},
					{Key: "export", Title: "Export history", Detail: "Allow CSV export of the simulated blotter.", Kind: "toggle", On: true},
				},
			},
			{
				Key:   "advanced",
				Title: "Advanced",
				Note:  "Developer and diagnostic switches.",
				Rows: []SettingRow{
					{Key: "telemetry", Title: "Local metrics", Detail: "Record frame timings for the devtools Frame tab.", Kind: "toggle", On: false},
					{Key: "devtools", Title: "Developer tools", Detail: "Open the inset dock for frame, element, and op inspection.", Kind: "toggle", On: true},
					{Key: "hardware", Title: "Hardware acceleration", Detail: "Use the host canvas for replay and retained layers.", Kind: "toggle", On: true},
					{Key: "cache", Title: "Warm paint caches", Detail: "Warm the raster, chunk, and text caches on first paint.", Kind: "toggle", On: true},
				},
			},
		},
		Shortcuts: []Shortcut{
			{Keys: "1 – 7", Action: "Switch rail screens"},
			{Keys: "/", Action: "Focus symbol search"},
			{Keys: "B / S", Action: "Buy or sell the selected symbol"},
			{Keys: "F12", Action: "Toggle developer tools"},
			{Keys: "Esc", Action: "Close the active panel"},
			{Keys: "↑ / ↓", Action: "Move through watchlist rows"},
			{Keys: "Enter", Action: "Open the selected symbol"},
			{Keys: "Cmd / Ctrl + K", Action: "Open the command palette"},
		},
		Theme:      "Midnight",
		Density:    "Compact",
		Currency:   "USD",
		Timezone:   "America/New_York",
		Build:      "v0.0.3 · local",
		DataSource: "Simulated · gofakeit",
	}
}
