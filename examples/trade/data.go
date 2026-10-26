//nolint:mnd,gosec // The sample market fixture is literal display content: quotes, sizes, and percentages are generated fixture values.
package main

import (
	"math"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// Candle is one synthetic OHLC bar. Volume is in shares.
type Candle struct {
	Time                   string
	Open, High, Low, Close float64
	Volume                 float64
}

// SymbolData is the generated snapshot for one instrument.
type SymbolData struct {
	Quote       Quote
	Candles     []Candle
	Book        []BookLevel
	Trades      []TradeRow
	BBO         BBO
	Donut       Donut
	DonutLegend []LegendRow
	Stats       []StatRow
	Options     []OptionRow
	Financials  []StatRow
}

// Demo is the deterministic sample dataset shared by every screen.
type Demo struct {
	Symbols      map[string]SymbolData
	Watch        []WatchRow
	Orders       []OrderRow
	Holdings     []Holding
	Indices      []IndexRow
	Tickers      []TickerItem
	Markets      MarketsData
	Portfolio    PortfolioData
	News         NewsData
	Screener     ScreenerData
	Settings     SettingsData
	ProfileName  string
	Clock        string
	Date         string
	MarketStatus string
}

// dataQuoteNumbers carries the raw prices a formatted quote is built from.
// The candle and order-book builders share it so every visible price agrees.
type dataQuoteNumbers struct {
	Price, PrevClose, Open, High, Low float64
	Bid, Ask                          float64
	BidSize, AskSize                  int
	Volume                            float64
	Decimals                          int
}

// newDemo builds the seeded sample dataset. The faker instance and the price
// RNG both carry fixed seeds, so the market data is identical on every run.
// Only the display clock reads the wall clock.
func newDemo() *Demo {
	rng := rand.New(rand.NewPCG(dataSeed, dataSeed>>1))
	faker := gofakeit.New(dataFakerSeed)
	specs := dataWatchSpecs()

	demo := &Demo{
		Symbols: make(map[string]SymbolData, len(specs)),
		Watch:   make([]WatchRow, 0, len(specs)),
	}

	for index, spec := range specs {
		data := dataBuildSymbol(spec, rng)
		demo.Symbols[spec.Symbol] = data
		demo.Watch = append(demo.Watch, dataWatchRow(index, spec, data.Quote))
	}

	demo.Orders = dataBuildOrders(faker, specs, rng)
	demo.Holdings = dataBuildHoldings(specs, demo.Symbols)
	demo.Indices = dataBuildIndices()
	demo.Tickers = dataBuildTickers(specs, demo.Symbols)
	demo.Markets = dataBuildMarkets(rng)
	demo.Portfolio = dataBuildPortfolio(demo.Holdings)
	demo.News = dataBuildNews()
	demo.Screener = dataBuildScreener(demo.Symbols)
	demo.Settings = dataBuildSettings()

	now := time.Now()
	demo.ProfileName = "Alex Morgan"
	demo.Clock = now.Format("15:04:05")
	demo.Date = now.Format("Mon 02 Jan 2006")
	demo.MarketStatus = "Market open"

	return demo
}

// dataWatchRow turns one generated quote into its watchlist line.
func dataWatchRow(index int, spec dataWatchSpec, quote Quote) WatchRow {
	return WatchRow{
		Index:     index,
		Symbol:    quote.Symbol,
		Name:      quote.Name,
		Price:     quote.Price,
		Change:    quote.Change,
		ChangePct: quote.ChangePct,
		Market:    quote.Market,
		Exchange:  quote.Exchange,
		Up:        quote.Up,
		Active:    false,
		Starred:   spec.Starred,
	}
}

// dataBuildSymbol produces every panel for one instrument.
func dataBuildSymbol(spec dataWatchSpec, rng *rand.Rand) SymbolData {
	numbers := dataQuoteNumbersFor(spec, rng)
	candles := dataBuildCandles(spec, numbers, rng)

	numbers.High = candles[0].High
	numbers.Low = candles[0].Low

	for _, candle := range candles {
		numbers.High = math.Max(numbers.High, candle.High)
		numbers.Low = math.Min(numbers.Low, candle.Low)
	}

	numbers.High = math.Max(numbers.High, math.Max(numbers.Open, spec.Price))
	numbers.Low = math.Min(numbers.Low, math.Min(numbers.Open, spec.Price))

	quote := dataBuildQuote(spec, numbers, rng)
	stats := dataBuildQuoteStats(quote, numbers)

	return SymbolData{
		Quote:       quote,
		Candles:     candles,
		Book:        dataBuildBook(numbers, rng),
		Trades:      dataBuildTrades(numbers, rng),
		BBO:         dataBuildBBO(numbers),
		Donut:       dataBuildDonut(),
		DonutLegend: dataBuildDonutLegend(),
		Stats:       stats,
		Options:     dataBuildOptions(spec, rng),
		Financials:  dataBuildFinancials(spec, quote, rng),
	}
}

// dataQuoteNumbersFor picks the session's open, range, and top of book around
// the curated close.
func dataQuoteNumbersFor(spec dataWatchSpec, rng *rand.Rand) dataQuoteNumbers {
	prevClose := spec.Price / (1 + spec.ChangePct/100)
	gap := (rng.Float64() - 0.5) * 0.006
	open := prevClose * (1 + gap)
	high := math.Max(open, spec.Price) * (1 + 0.0015 + rng.Float64()*0.004)
	low := math.Min(open, spec.Price) * (1 - 0.0015 - rng.Float64()*0.004)
	tick := math.Pow(10, float64(-spec.Decimals))
	bid := spec.Price - tick*float64(1+rng.IntN(2))
	ask := bid + tick*float64(1+rng.IntN(2))

	return dataQuoteNumbers{
		Price:     spec.Price,
		PrevClose: prevClose,
		Open:      open,
		High:      high,
		Low:       low,
		Bid:       bid,
		Ask:       ask,
		BidSize:   200 + rng.IntN(1800),
		AskSize:   200 + rng.IntN(1800),
		Volume:    spec.Volume,
		Decimals:  spec.Decimals,
	}
}

// dataBuildQuote formats the instrument header and its risk fields.
func dataBuildQuote(spec dataWatchSpec, numbers dataQuoteNumbers, rng *rand.Rand) Quote {
	change := numbers.Price - numbers.PrevClose
	week52High := numbers.Price * (1.18 + rng.Float64()*0.35)
	week52Low := numbers.Price * (0.58 + rng.Float64()*0.22)
	sharesOut := spec.MarketCap / numbers.Price
	mark := dataCurrencyMark(spec.Currency)

	return Quote{
		Symbol:     spec.Symbol,
		Name:       spec.Name,
		Exchange:   spec.Exchange,
		Market:     spec.Market,
		Currency:   spec.Currency,
		Price:      dataFmtPrice(numbers.Price, numbers.Decimals),
		Change:     dataSignedNum(change, numbers.Decimals),
		ChangePct:  dataSignedPct(spec.ChangePct),
		Up:         change >= 0,
		Open:       dataFmtPrice(numbers.Open, numbers.Decimals),
		High:       dataFmtPrice(numbers.High, numbers.Decimals),
		Low:        dataFmtPrice(numbers.Low, numbers.Decimals),
		PrevClose:  dataFmtPrice(numbers.PrevClose, numbers.Decimals),
		Volume:     dataShortNum(numbers.Volume),
		Turnover:   dataShortNum(numbers.Volume * numbers.Price),
		MarketCap:  mark + dataShortNum(spec.MarketCap),
		SharesOut:  dataShortNum(sharesOut),
		Week52High: dataFmtPrice(week52High, numbers.Decimals),
		Week52Low:  dataFmtPrice(week52Low, numbers.Decimals),
		DayRange:   dataFmtPrice(numbers.Low, numbers.Decimals) + " - " + dataFmtPrice(numbers.High, numbers.Decimals),
		Bid:        dataFmtPrice(numbers.Bid, numbers.Decimals),
		BidSize:    dataComma(float64(numbers.BidSize), 0),
		Ask:        dataFmtPrice(numbers.Ask, numbers.Decimals),
		AskSize:    dataComma(float64(numbers.AskSize), 0),
		Status:     "Open",
		Session:    "Regular",
		Updated:    dataMarketUpdated,
		LogoClass:  dataLogoClass(spec.Symbol),
	}
}

// dataBuildCandles walks 180 intraday bars from the open to the quote price.
// A final bridge correction spreads the residual between the random walk and
// the curated close across the whole session, so the last candle does not
// carry a visible jump.
func dataBuildCandles(spec dataWatchSpec, numbers dataQuoteNumbers, rng *rand.Rand) []Candle {
	count := dataCandleCount
	drift := (spec.Price - numbers.Open) / float64(count)
	baseVolume := spec.Volume / float64(count)
	candles := make([]Candle, 0, count)
	last := numbers.Open

	for index := range count {
		open := last
		close := open + drift + rng.NormFloat64()*spec.Price*0.0022

		wick := math.Abs(close-open)*float64(0.15+rng.Float64()*0.55) + spec.Price*0.0003
		high := math.Max(open, close) + wick
		low := math.Min(open, close) - wick
		if low <= 0 {
			low = math.Min(open, close) * 0.99
		}

		volume := baseVolume * (0.55 + rng.Float64()*0.9)
		if volume <= 0 {
			volume = baseVolume
		}

		candles = append(candles, Candle{
			Time:   dataCandleClock(index, count),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
		last = close
	}

	// Spread the residual so the final close lands on the curated price while
	// the opening bar stays at the session open.
	residual := spec.Price - candles[len(candles)-1].Close
	for index := range candles {
		shift := residual * float64(index+1) / float64(count)
		candles[index].Open += shift
		candles[index].High += shift
		candles[index].Low += shift
		candles[index].Close += shift
	}

	// Rebuild the opens from the bridged closes and clamp the wicks around
	// them, keeping every bar internally consistent.
	for index := range candles {
		if index > 0 {
			candles[index].Open = candles[index-1].Close
		}

		if candles[index].High < math.Max(candles[index].Open, candles[index].Close) {
			candles[index].High = math.Max(candles[index].Open, candles[index].Close)
		}

		if candles[index].Low > math.Min(candles[index].Open, candles[index].Close) {
			candles[index].Low = math.Min(candles[index].Open, candles[index].Close)
		}
	}

	candles[len(candles)-1].Close = spec.Price

	return candles
}

// dataBuildBook stacks five bid and five ask levels. The totals accumulate
// away from the touch and the depth is a percentage of the deepest level.
func dataBuildBook(numbers dataQuoteNumbers, rng *rand.Rand) []BookLevel {
	tick := math.Pow(10, float64(-numbers.Decimals))
	sizes := make([]int, dataBookDepth*2)

	bids := make([]float64, dataBookDepth)
	asks := make([]float64, dataBookDepth)

	for level := range dataBookDepth {
		bids[level] = numbers.Bid - float64(level)*tick
		asks[level] = numbers.Ask + float64(level)*tick
		sizes[level] = numbers.BidSize
		sizes[dataBookDepth+level] = numbers.AskSize

		if level > 0 {
			sizes[level] += rng.IntN(900)
			sizes[dataBookDepth+level] += rng.IntN(900)
		}
	}

	totals := make([]int, dataBookDepth*2)
	bidTotal, askTotal := 0, 0
	maxTotal := 0

	for level := range dataBookDepth {
		bidTotal += sizes[level]
		askTotal += sizes[dataBookDepth+level]
		totals[level] = bidTotal
		totals[dataBookDepth+level] = askTotal
	}

	for _, total := range totals {
		maxTotal = max(maxTotal, total)
	}

	book := make([]BookLevel, 0, dataBookDepth*2)

	for level := range dataBookDepth {
		book = append(book, dataBookRow(bids[level], sizes[level], totals[level], maxTotal, numbers.Decimals, "bid", level == 0))
	}

	for level := range dataBookDepth {
		book = append(book, dataBookRow(asks[level], sizes[dataBookDepth+level], totals[dataBookDepth+level], maxTotal, numbers.Decimals, "ask", level == 0))
	}

	return book
}

// dataBookRow formats one depth level.
func dataBookRow(price float64, size, total, maxTotal, decimals int, side string, best bool) BookLevel {
	depth := 0
	if maxTotal > 0 {
		depth = total * 100 / maxTotal
	}

	return BookLevel{
		Price: dataFmtPrice(price, decimals),
		Size:  dataComma(float64(size), 0),
		Total: dataComma(float64(total), 0),
		Depth: strconv.Itoa(depth),
		Side:  side,
		Best:  best,
	}
}

// dataBuildBBO mirrors the best bid and offer in the book.
func dataBuildBBO(numbers dataQuoteNumbers) BBO {
	return BBO{
		BidPrice: dataFmtPrice(numbers.Bid, numbers.Decimals),
		BidSize:  dataComma(float64(numbers.BidSize), 0),
		AskPrice: dataFmtPrice(numbers.Ask, numbers.Decimals),
		AskSize:  dataComma(float64(numbers.AskSize), 0),
	}
}

// dataBuildTrades prints the last dozen transactions, newest first.
func dataBuildTrades(numbers dataQuoteNumbers, rng *rand.Rand) []TradeRow {
	tick := math.Pow(10, float64(-numbers.Decimals))
	base := 10*3600 + 41*60 + 58
	trades := make([]TradeRow, 0, dataTradeCount)

	for index := range dataTradeCount {
		buyer := rng.Float64() > 0.44
		price := numbers.Bid - tick*float64(rng.IntN(2))
		aggressor := "S"

		if buyer {
			price = numbers.Ask + tick*float64(rng.IntN(2))
			aggressor = "B"
		}

		size := 100 * (1 + rng.IntN(24))
		second := base - index*9 - rng.IntN(5)

		trades = append(trades, TradeRow{
			Time:      dataClockSeconds(second),
			Price:     dataFmtPrice(price, numbers.Decimals),
			Size:      dataComma(float64(size), 0),
			Up:        buyer,
			Aggressor: aggressor,
		})
	}

	return trades
}

// dataDonutValues is the fixed transaction mix, in thousands, that the ring
// and its legend share. The net is -2012.31K.
func dataDonutValues() (inflow []float64, outflow []float64) {
	return []float64{4286.40, 1912.75}, []float64{3845.20, 2097.44, 1458.62, 810.20}
}

// dataDonutColors are the two inflow greens followed by the four outflow
// reds, in the same order as the values.
func dataDonutColors() []string {
	return []string{"#3ddc84", "#2a9d5c", "#f0606a", "#c84a54", "#8f3942", "#5f2a30"}
}

// dataBuildDonut lays the six transaction arcs onto the 44 unit circle.
func dataBuildDonut() Donut {
	inflow, outflow := dataDonutValues()
	values := append(append([]float64{}, inflow...), outflow...)
	colors := dataDonutColors()
	circumference := 2 * math.Pi * 44
	total := 0.0
	inflowTotal := 0.0
	outflowTotal := 0.0

	for _, value := range inflow {
		inflowTotal += value
	}

	for _, value := range outflow {
		outflowTotal += value
	}

	total = inflowTotal + outflowTotal
	segments := make([]DonutSegment, 0, len(values))
	offset := 0.0

	for index, value := range values {
		dash := value / total * circumference
		start := offset / circumference
		end := (offset + dash) / circumference
		segments = append(segments, DonutSegment{
			Color:  colors[index],
			Dash:   dash,
			Offset: -offset,
			Path:   dataDonutPath(start, end),
		})
		offset += dash
	}

	net := inflowTotal - outflowTotal

	return Donut{
		Segments: segments,
		Inflow:   dataFixed(inflowTotal, 2),
		Outflow:  dataFixed(outflowTotal, 2),
		Net:      dataFixed(net, 2),
		NetPct:   dataFixed(net/total*100, 2) + "%",
		NetUp:    net >= 0,
		Unit:     "K",
	}
}

// dataDonutPath draws one arc in the 120 unit donut square. The start and end
// are fractions of the circle beginning at twelve o'clock.
func dataDonutPath(start, end float64) string {
	if end <= start {
		return ""
	}

	const (
		donutCenter = 60.0
		donutRadius = 44.0
	)

	begin := -math.Pi/2 + start*2*math.Pi
	finish := -math.Pi/2 + end*2*math.Pi
	x1 := donutCenter + donutRadius*math.Cos(begin)
	y1 := donutCenter + donutRadius*math.Sin(begin)
	x2 := donutCenter + donutRadius*math.Cos(finish)
	y2 := donutCenter + donutRadius*math.Sin(finish)
	large := 0
	if end-start > 0.5 {
		large = 1
	}

	return "M " + strconv.FormatFloat(x1, 'f', 2, 64) + " " + strconv.FormatFloat(y1, 'f', 2, 64) +
		" A " + strconv.FormatFloat(donutRadius, 'f', 2, 64) + " " + strconv.FormatFloat(donutRadius, 'f', 2, 64) +
		" 0 " + strconv.Itoa(large) + " 1 " + strconv.FormatFloat(x2, 'f', 2, 64) + " " + strconv.FormatFloat(y2, 'f', 2, 64)
}

// dataBuildDonutLegend labels each arc and prints its share of turnover.
func dataBuildDonutLegend() []LegendRow {
	inflow, outflow := dataDonutValues()
	values := append(append([]float64{}, inflow...), outflow...)
	colors := dataDonutColors()
	labels := []string{"Block inflow", "Sweep inflow", "Block outflow", "Sweep outflow", "Dark pool", "Retail flow"}
	total := 0.0

	for _, value := range values {
		total += value
	}

	out := make([]LegendRow, 0, len(values))

	for index, value := range values {
		out = append(out, LegendRow{
			Label: labels[index],
			Value: dataFmtPrice(value/total*100, 1) + "%",
			Color: colors[index],
			Up:    index < len(inflow),
		})
	}

	return out
}

// dataBuildQuoteStats is the eight-row session grid beside the quote.
func dataBuildQuoteStats(quote Quote, numbers dataQuoteNumbers) []StatRow {
	return []StatRow{
		{Label: "Open", Value: quote.Open, Up: numbers.Open >= numbers.PrevClose},
		{Label: "High", Value: quote.High, Up: true},
		{Label: "Low", Value: quote.Low},
		{Label: "Prev Close", Value: quote.PrevClose, Up: quote.Up},
		{Label: "Volume", Value: quote.Volume, Up: true},
		{Label: "Turnover", Value: quote.Turnover, Up: quote.Up},
		{Label: "Market Cap", Value: quote.MarketCap, Up: true},
		{Label: "52W range", Value: quote.Week52Low + " - " + quote.Week52High, Up: quote.Up},
	}
}

// dataStrikeStep picks a plausible call strike increment for the price.
func dataStrikeStep(price float64) float64 {
	switch {
	case price < 25:
		return 1
	case price < 100:
		return 2.5
	case price < 250:
		return 5
	case price < 500:
		return 10
	default:
		return 20
	}
}

// dataBuildOptions lists seven calls centred near the spot price.
func dataBuildOptions(spec dataWatchSpec, rng *rand.Rand) []OptionRow {
	step := dataStrikeStep(spec.Price)
	centre := math.Round(spec.Price/step) * step
	rows := make([]OptionRow, 0, dataOptionCount)

	for offset := -3; offset <= 3; offset++ {
		strike := centre + float64(offset)*step
		if strike <= 0 {
			strike = step
		}

		intrinsic := math.Max(spec.Price-strike, 0)
		extrinsic := spec.Price*0.012*(1-math.Abs(float64(offset))/4) + step*0.08
		last := intrinsic + extrinsic
		bid := math.Max(last-step*0.35, 0.01)
		ask := last + step*0.35
		delta := 0.5 - float64(offset)*0.12
		delta = math.Min(math.Max(delta, 0.08), 0.92)
		iv := 24 + rng.Float64()*38 + math.Abs(float64(offset))*1.8

		rows = append(rows, OptionRow{
			Strike: dataFmtPrice(strike, 2),
			Last:   dataFmtPrice(last, 2),
			Bid:    dataFmtPrice(bid, 2),
			Ask:    dataFmtPrice(ask, 2),
			Volume: dataComma(float64(100*(1+rng.IntN(48))), 0),
			OI:     dataComma(float64(200+rng.IntN(12000)), 0),
			IV:     dataFmtPrice(iv, 1) + "%",
			Delta:  dataFmtPrice(delta, 2),
			ITM:    strike < spec.Price,
		})
	}

	return rows
}

// dataBuildFinancials prints the ten fundamentals rows.
func dataBuildFinancials(spec dataWatchSpec, quote Quote, rng *rand.Rand) []StatRow {
	pe := spec.PE
	if pe <= 0 {
		pe = 20
	}

	mark := dataCurrencyMark(spec.Currency)
	eps := spec.Price / pe
	gross := 38 + rng.Float64()*42
	net := math.Min(8+rng.Float64()*24, gross*0.55)
	roe := 14 + rng.Float64()*38
	debt := 0.12 + rng.Float64()*1.2
	beta := spec.Beta
	if beta <= 0 {
		beta = 1
	}

	return []StatRow{
		{Label: "Market cap", Value: quote.MarketCap},
		{Label: "P/E", Value: dataFmtPrice(pe, 1)},
		{Label: "EPS", Value: mark + dataFmtPrice(eps, 2)},
		{Label: "Revenue", Value: mark + dataShortNum(spec.Revenue)},
		{Label: "Gross margin", Value: dataFmtPrice(gross, 1) + "%", Up: true},
		{Label: "Net margin", Value: dataFmtPrice(net, 1) + "%", Up: true},
		{Label: "ROE", Value: dataFmtPrice(roe, 1) + "%", Up: true},
		{Label: "Debt/Equity", Value: dataFmtPrice(debt, 2)},
		{Label: "Beta", Value: dataFmtPrice(beta, 2)},
		{Label: "Dividend yield", Value: dataFmtPrice(spec.Dividend, 2) + "%"},
	}
}
