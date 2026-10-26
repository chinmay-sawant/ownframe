package main

// View is the data the terminal template prints. The fields are plain values
// and preformatted strings so the page sends one SetData call per interaction.
type View struct {
	Screen          string
	Clock           string
	Date            string
	Search          string
	UpdateAvailable bool
	MarketOpen      bool
	MarketStatus    string
	SessionText     string
	Latency         string
	Notifications   int
	ProfileName     string
	ProfileInitials string
	AvatarClass     string

	Quote        Quote
	Watch        []WatchRow
	WatchFilter  string
	WatchFilters []FilterChip
	WatchCount   int
	WatchTotal   int

	Chart      Chart
	ChartTab   string
	ChartTabs  []Tab
	Timeframes []Chip
	Timeframe  string
	Indicators []Chip
	Options    []OptionRow
	Financials []StatRow
	ChartNews  []NewsItem
	RangeStats []StatRow
	KeyStats   []StatRow

	OrderTab     string
	OrderTabs    []Tab
	Orders       []OrderRow
	Holdings     []Holding
	OrderFilters []SelectView
	OrderStats   OrderStats
	Ticket       Ticket

	Book        []BookLevel
	Trades      []TradeRow
	BBO         BBO
	Donut       Donut
	DonutLegend []LegendRow
	QuoteStats  []StatRow
	DepthRatio  string

	Indices []IndexRow
	Tickers []TickerItem

	Markets   MarketsData
	Portfolio PortfolioData
	News      NewsData
	Screener  ScreenerData
	Settings  SettingsData

	Toast Toast
}

// Quote is the selected instrument and its session snapshot.
type Quote struct {
	Symbol, Name, Exchange, Market, Currency string
	Price, Change, ChangePct                 string
	Up                                       bool
	Open, High, Low, PrevClose               string
	Volume, Turnover, MarketCap, SharesOut   string
	Week52High, Week52Low, DayRange          string
	Bid, BidSize, Ask, AskSize               string
	Status, Session, Updated                 string
	LogoClass                                string
}

// WatchRow is one instrument in the watchlist. Price and the change fields are
// already formatted for display.
type WatchRow struct {
	Index                                  int
	Symbol, Name, Price, Change, ChangePct string
	Market, Exchange                       string
	Up, Active, Starred                    bool
}

// FilterChip is one watchlist filter or preset button.
type FilterChip struct {
	Key, Label string
	Active     bool
}

// Chip is a timeframe, indicator, or preset button.
type Chip struct {
	Key, Label string
	Active     bool
}

// Tab is one tab above the chart or the order blotter.
type Tab struct {
	Key, Label string
	Active     bool
	Count      int
}

// SelectView is one read-only filter control in the blotter header.
type SelectView struct {
	Label, Value string
}

// StatRow is one label/value pair in a stat grid.
type StatRow struct {
	Label, Value string
	Up           bool
	Tone         string
}

// OptionRow is one strike in the options chain.
type OptionRow struct {
	Strike, Last, Bid, Ask, Volume, OI, IV, Delta string
	ITM                                           bool
}

// OrderRow is one row of the order blotter and the full order history.
type OrderRow struct {
	Index                                                          int
	ID, Time, Symbol, Name, Side, Status, Type, Qty, Filled, Limit string
	Done, Pending, Cancelled, Sell                                 bool
}

// Holding is one position row and the portfolio's allocation source.
type Holding struct {
	Symbol, Name, Qty, Avg, Last, Value, PNL, Pct, Weight string
	Up                                                    bool
	Bar                                                   float64
}

// OrderStats is the small summary strip above the blotter.
type OrderStats struct {
	Filled, Working, Cancelled, Notional string
}

// Ticket is the order ticket state on the orders screen and the buy/sell panel.
type Ticket struct {
	Symbol, Side, Type, Qty, Limit, TIF, Estimated, BuyingPower, Status string
	SideBuy                                                             bool
}

// BookLevel is one price level in the order book.
type BookLevel struct {
	Price, Size, Total, Depth string
	Side                      string
	Best                      bool
}

// TradeRow is one time-and-sales print.
type TradeRow struct {
	Time, Price, Size string
	Up                bool
	Aggressor         string
}

// BBO is the best bid and offer strip.
type BBO struct {
	BidPrice, BidSize, AskPrice, AskSize string
}

// Donut is the transaction-statistics ring.
type Donut struct {
	Segments                     []DonutSegment
	Inflow, Outflow, Net, NetPct string
	NetUp                        bool
	Unit                         string
}

// DonutSegment is one arc of the inline SVG transaction ring. Path is the arc
// in a 120 unit square; Dash and Offset remain for small previews.
type DonutSegment struct {
	Color        string
	Dash, Offset float64
	Path         string
}

// LegendRow is one label in the transaction-statistics legend.
type LegendRow struct {
	Label, Value, Color string
	Up                  bool
}

// IndexRow is one market index in the footer.
type IndexRow struct {
	Name, Value, Change, ChangePct string
	Up                             bool
}

// TickerItem is one instrument in the scrolling tape.
type TickerItem struct {
	Symbol, Price, Change, ChangePct string
	Up                               bool
}

// Chart is the prepared candlestick chart. The CSS plot uses a 1000 by 400
// design box; the labels are HTML overlays positioned in percent.
type Chart struct {
	Candles    []CandleBar
	Lines      []ChartLine
	YLabels    []ChartLabel
	XLabels    []ChartLabel
	Volume     []VolumeBar
	Summary    ChartSummary
	LastPrice  string
	LastTop    string
	HighLabel  string
	HighTop    string
	LowLabel   string
	LowTop     string
	CrossX     string
	CrossY     string
	CrossPrice string
	CrossTime  string
	SlotLead   string
	SlotWidth  string
	Up         bool
}

// ChartSummary is the line above the plot.
type ChartSummary struct {
	Period, Change, Pct, Open, High, Low, Close, Volume string
	Up                                                  bool
}

// CandleBar is one candle in the CSS plot.
type CandleBar struct {
	X, Left, Width         float64
	WickTop, WickBottom    float64
	BodyTop, Height        float64
	WickOffset, WickHeight float64
	SlotWidth              float64
	Up                     bool
}

// ChartLine is one moving average. Points keeps the source data; Segments is
// the HTML/CSS line split into rotated spans.
type ChartLine struct {
	Points   string
	Color    string
	Width    float64
	SlotLead string
	Segments []ChartSegment
}

// ChartSegment is one painted line segment in the CSS plot.
type ChartSegment struct {
	Left, Top, Length, Angle float64
	Color                    string
	Width                    float64
}

// GridLine is a horizontal or vertical rule in the plot.
type GridLine struct {
	X1, Y1, X2, Y2 float64
	Major          bool
}

// ChartLabel is an HTML label over the CSS plot. Top and Left are percentages.
type ChartLabel struct {
	Top, Left, Text, Tone string
}

// VolumeBar is one volume column under the candles.
type VolumeBar struct {
	X, Top, Height float64
	Up             bool
}

// MarketsData is the market-overview screen.
type MarketsData struct {
	Breadth  []BreadthStat
	Sessions []SessionRow
	Sectors  []SectorRow
	Gainers  []MoverRow
	Losers   []MoverRow
	Calendar []CalendarEvent
	Indices  []IndexCard
	Flows    []FlowRow
	Updated  string
}

// BreadthStat is one advance/decline or new-high statistic.
type BreadthStat struct {
	Label, Value, Note string
	Pct                float64
	Up                 bool
}

// SessionRow is one regional trading session.
type SessionRow struct {
	Name, Region, Status, Change, Pct string
	Up, Open                          bool
}

// SectorRow is one heatmap tile and sector row.
type SectorRow struct {
	Name, Pct string
	Up        bool
	Weight    float64
	Delay     string
}

// MoverRow is one gainer or loser.
type MoverRow struct {
	Symbol, Name, Price, Pct, Volume string
	Up                               bool
}

// CalendarEvent is one economic calendar row.
type CalendarEvent struct {
	Time, Country, Event, Actual, Forecast, Prior, Impact string
	High                                                  bool
}

// IndexCard is one large index card on the markets screen.
type IndexCard struct {
	Name, Symbol, Value, Change, Pct string
	Up                               bool
	Spark                            []SparkBar
}

// SparkBar is one bar in a mini sparkline.
type SparkBar struct {
	Height float64
	Up     bool
}

// FlowRow is one sector or asset-class flow row.
type FlowRow struct {
	Label, Value string
	Up           bool
	Pct          float64
}

// PortfolioData is the portfolio screen.
type PortfolioData struct {
	TotalValue, DayPL, DayPct, TotalPL, TotalPct string
	BuyingPower, Cash, Invested, Margin          string
	DayUp, TotalUp                               bool
	Holdings                                     []Holding
	Allocation                                   []AllocationSlice
	Performance                                  PerformanceChart
	Events                                       []PortfolioEvent
	Stats                                        []StatRow
}

// AllocationSlice is one row of the allocation legend. Weight is the share
// of the portfolio in percent and drives the stacked bar width.
type AllocationSlice struct {
	Label, Value, Pct, Color string
	Weight                   float64
}

// PerformanceChart is the portfolio value line.
type PerformanceChart struct {
	Points, BenchmarkPoints string
	Segments                []ChartSegment
	BenchmarkSegments       []ChartSegment
	Grid                    []GridLine
	YLabels, XLabels        []ChartLabel
	Start, End, Return      string
	Up                      bool
}

// PortfolioEvent is one account activity row.
type PortfolioEvent struct {
	Time, Text, Value string
	Up                bool
}

// NewsData is the news screen.
type NewsData struct {
	Items     []NewsItem
	Trending  []TrendingTicker
	Sentiment []SentimentSlice
	Sources   []Chip
	Calendar  []CalendarEvent
	Filter    string
	Gauge     int
	GaugeText string
}

// NewsItem is one headline row.
type NewsItem struct {
	Source, Time, Title, Summary, Ticker, Tag string
	Hot, Up                                   bool
}

// TrendingTicker is one trending symbol in the news rail.
type TrendingTicker struct {
	Symbol, Name, Mentions, Pct string
	Up                          bool
}

// SentimentSlice is one row of the sentiment breakdown.
type SentimentSlice struct {
	Label, Pct string
	Up         bool
}

// ScreenerData is the screener screen.
type ScreenerData struct {
	Presets []Chip
	Filters []ScreenerFilter
	Rows    []ScreenerRow
	Matches int
	Updated string
}

// ScreenerFilter is one filter chip in the screener builder.
type ScreenerFilter struct {
	Key, Label, Value string
	Active            bool
}

// ScreenerRow is one result in the screener table.
type ScreenerRow struct {
	Symbol, Name, Sector, Price, Change, Pct string
	Volume, MarketCap, PE, RSI, Signal       string
	Up                                       bool
	Score                                    int
	Spark                                    []SparkBar
}

// SettingsData is the settings screen.
type SettingsData struct {
	Sections   []SettingSection
	Shortcuts  []Shortcut
	Theme      string
	Density    string
	Currency   string
	Timezone   string
	Build      string
	DataSource string
}

// SettingSection is one settings group.
type SettingSection struct {
	Key, Title, Note string
	Rows             []SettingRow
}

// SettingRow is one toggle or choice.
type SettingRow struct {
	Key, Title, Detail, Kind, Value string
	On                              bool
}

// Shortcut is one keyboard binding row.
type Shortcut struct {
	Keys, Action string
}

// Toast is the transient confirmation at the lower right.
type Toast struct {
	Show, Up bool
	Kind     string
	Title    string
	Message  string
}
