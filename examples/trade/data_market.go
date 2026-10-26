//nolint:mnd // The sample market fixture is curated display content: tickers, prices, and sizes are fixture values.
package main

import (
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

// dataWatchSpec is the curated seed for one watchlist instrument. The close
// and session change are fixed; the generator derives the rest around them.
type dataWatchSpec struct {
	Symbol, Name, Exchange, Market, Currency string
	Price, ChangePct                         float64
	Decimals                                 int
	Starred                                  bool
	Volume, MarketCap, PE, Revenue           float64
	Beta, Dividend                           float64
}

// dataWatchSpecs is the sixteen-row watchlist. Market is US, HK, or EV;
// electric-vehicle names carry EV so the watchlist filter can find them.
func dataWatchSpecs() []dataWatchSpec {
	return []dataWatchSpec{
		{Symbol: "NVDA", Name: "NVIDIA Corporation", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 183.26, ChangePct: 2.84, Decimals: 2, Starred: true, Volume: 168.4e6, MarketCap: 4.48e12, PE: 58.4, Revenue: 130.5e9, Beta: 1.68, Dividend: 0.03},
		{Symbol: "AAPL", Name: "Apple Inc.", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 246.38, ChangePct: 0.71, Decimals: 2, Volume: 52.1e6, MarketCap: 3.71e12, PE: 34.9, Revenue: 391.0e9, Beta: 1.12, Dividend: 0.44},
		{Symbol: "TSLA", Name: "Tesla, Inc.", Exchange: "NASDAQ", Market: "EV", Currency: "USD", Price: 428.62, ChangePct: -1.92, Decimals: 2, Volume: 96.8e6, MarketCap: 1.38e12, PE: 92.4, Revenue: 97.7e9, Beta: 2.31},
		{Symbol: "MSFT", Name: "Microsoft Corporation", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 512.94, ChangePct: 0.42, Decimals: 2, Volume: 21.4e6, MarketCap: 3.82e12, PE: 36.8, Revenue: 245.1e9, Beta: 0.91, Dividend: 0.68},
		{Symbol: "AMD", Name: "Advanced Micro Devices", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 204.18, ChangePct: 1.63, Decimals: 2, Volume: 48.2e6, MarketCap: 331.2e9, PE: 41.6, Revenue: 25.8e9, Beta: 1.74},
		{Symbol: "META", Name: "Meta Platforms", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 741.55, ChangePct: -0.58, Decimals: 2, Volume: 14.8e6, MarketCap: 1.88e12, PE: 28.4, Revenue: 164.5e9, Beta: 1.22, Dividend: 0.28},
		{Symbol: "AMZN", Name: "Amazon.com, Inc.", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 231.07, ChangePct: 0.96, Decimals: 2, Volume: 38.6e6, MarketCap: 2.46e12, PE: 38.7, Revenue: 638.0e9, Beta: 1.15},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 197.84, ChangePct: 0.33, Decimals: 2, Volume: 27.9e6, MarketCap: 2.42e12, PE: 25.3, Revenue: 350.0e9, Beta: 1.04, Dividend: 0.42},
		{Symbol: "JPM", Name: "JPMorgan Chase & Co.", Exchange: "NYSE", Market: "US", Currency: "USD", Price: 289.41, ChangePct: -0.24, Decimals: 2, Volume: 9.6e6, MarketCap: 812.4e9, PE: 15.1, Revenue: 178.0e9, Beta: 1.08, Dividend: 1.96},
		{Symbol: "GS", Name: "Goldman Sachs Group", Exchange: "NYSE", Market: "US", Currency: "USD", Price: 612.77, ChangePct: 1.12, Decimals: 2, Volume: 2.4e6, MarketCap: 198.6e9, PE: 17.4, Revenue: 53.5e9, Beta: 1.31, Dividend: 1.82},
		{Symbol: "SPY", Name: "SPDR S&P 500 ETF Trust", Exchange: "NYSE", Market: "US", Currency: "USD", Price: 604.32, ChangePct: 0.26, Decimals: 2, Volume: 71.3e6, MarketCap: 642.8e9, PE: 27.6, Revenue: 12.6e9, Beta: 1, Dividend: 1.16},
		{Symbol: "QQQ", Name: "Invesco QQQ Trust", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 531.58, ChangePct: 0.51, Decimals: 2, Volume: 32.5e6, MarketCap: 388.4e9, PE: 31.8, Revenue: 9.8e9, Beta: 1.11, Dividend: 0.52},
		{Symbol: "ARM", Name: "Arm Holdings plc", Exchange: "NASDAQ", Market: "US", Currency: "USD", Price: 176.92, ChangePct: -2.35, Decimals: 2, Starred: true, Volume: 8.7e6, MarketCap: 185.2e9, PE: 128.6, Revenue: 3.8e9, Beta: 1.42},
		{Symbol: "BABA", Name: "Alibaba Group Holding", Exchange: "NYSE", Market: "US", Currency: "USD", Price: 118.64, ChangePct: 1.08, Decimals: 2, Volume: 18.2e6, MarketCap: 289.6e9, PE: 18.4, Revenue: 137.0e9, Beta: 0.86, Dividend: 0.92},
		{Symbol: "9988.HK", Name: "Alibaba Group (HK)", Exchange: "HKEX", Market: "HK", Currency: "HKD", Price: 116.300, ChangePct: -0.77, Decimals: 3, Volume: 24.6e6, MarketCap: 2.26e12, PE: 19.2, Revenue: 1.02e12, Beta: 0.84, Dividend: 1.04},
		{Symbol: "RIVN", Name: "Rivian Automotive", Exchange: "NASDAQ", Market: "EV", Currency: "USD", Price: 14.86, ChangePct: 3.41, Decimals: 2, Volume: 41.8e6, MarketCap: 16.4e9, PE: 20, Revenue: 5.2e9, Beta: 2.08},
	}
}

// dataOrderSeed is one row of the order history. An empty Limit means a market
// order. Filled is the executed quantity; Done orders fill completely.
type dataOrderSeed struct {
	Symbol, Side, Type, Status, Time string
	Qty, Filled                      int
	Limit                            float64
}

// dataBuildOrders lays out today's blotter and the older history. The first
// ten rows are today's orders; the remaining eight carry past dates.
func dataBuildOrders(faker *gofakeit.Faker, specs []dataWatchSpec, _ *rand.Rand) []OrderRow {
	seeds := []dataOrderSeed{
		{Symbol: "NVDA", Side: "Buy", Type: "Limit", Status: "Fully Done", Qty: 100, Filled: 100, Limit: 183.10},
		{Symbol: "AMD", Side: "Buy", Type: "Market", Status: "Fully Done", Qty: 200, Filled: 200},
		{Symbol: "TSLA", Side: "Buy", Type: "Limit", Status: "Working", Qty: 33, Limit: 421.50},
		{Symbol: "AAPL", Side: "Buy", Type: "Limit", Status: "Fully Done", Qty: 150, Filled: 150, Limit: 245.20},
		{Symbol: "MSFT", Side: "Sell", Type: "Limit", Status: "Fully Done", Qty: 40, Filled: 40, Limit: 516.00},
		{Symbol: "META", Side: "Buy", Type: "Limit", Status: "Working", Qty: 10, Filled: 4, Limit: 728.00},
		{Symbol: "AMZN", Side: "Buy", Type: "Market", Status: "Fully Done", Qty: 100, Filled: 100},
		{Symbol: "GOOGL", Side: "Buy", Type: "Limit", Status: "Withdrawn", Qty: 33, Limit: 194.50},
		{Symbol: "JPM", Side: "Buy", Type: "Limit", Status: "Fully Done", Qty: 25, Filled: 25, Limit: 288.10},
		{Symbol: "GS", Side: "Buy", Type: "Limit", Status: "Working", Qty: 10, Limit: 608.00},
		{Symbol: "SPY", Side: "Buy", Type: "Limit", Status: "Fully Done", Qty: 50, Filled: 50, Limit: 601.20},
		{Symbol: "QQQ", Side: "Buy", Type: "Limit", Status: "Cancelled", Qty: 100, Limit: 548.00},
		{Symbol: "ARM", Side: "Sell", Type: "Limit", Status: "Fully Done", Qty: 200, Filled: 200, Limit: 182.40},
		{Symbol: "BABA", Side: "Buy", Type: "Market", Status: "Fully Done", Qty: 500, Filled: 500},
		{Symbol: "9988.HK", Side: "Buy", Type: "Limit", Status: "Working", Qty: 400, Filled: 120, Limit: 114.80},
		{Symbol: "RIVN", Side: "Buy", Type: "Limit", Status: "Fully Done", Qty: 1000, Filled: 1000, Limit: 13.90},
		{Symbol: "NVDA", Side: "Buy", Type: "Limit", Status: "Withdrawn", Qty: 50, Limit: 176.00},
		{Symbol: "TSLA", Side: "Buy", Type: "Limit", Status: "Fully Done", Qty: 20, Filled: 20, Limit: 402.10},
	}
	pastTimes := []string{"23 Oct 15:12", "23 Oct 14:48", "23 Oct 11:06", "22 Oct 15:32", "22 Oct 10:18", "21 Oct 14:02", "21 Oct 09:47", "20 Oct 13:24"}
	specBySymbol := make(map[string]dataWatchSpec, len(specs))

	for _, spec := range specs {
		specBySymbol[spec.Symbol] = spec
	}

	rows := make([]OrderRow, 0, len(seeds))

	for index, seed := range seeds {
		spec := specBySymbol[seed.Symbol]
		limit := "MKT"
		if seed.Limit > 0 {
			limit = dataFmtPrice(seed.Limit, spec.Decimals)
		}

		when := dataClockSeconds(9*3600 + 31*60 + 7 + index*7*60)
		if index >= 10 {
			when = pastTimes[(index-10)%len(pastTimes)]
		}

		row := OrderRow{
			Index:  index,
			ID:     strings.ToUpper(faker.LetterN(2)) + faker.DigitN(4),
			Time:   when,
			Symbol: seed.Symbol,
			Name:   spec.Name,
			Side:   seed.Side,
			Status: seed.Status,
			Type:   seed.Type,
			Qty:    dataComma(float64(seed.Qty), 0),
			Filled: dataComma(float64(seed.Filled), 0),
			Limit:  limit,
			Sell:   seed.Side == "Sell",
		}

		switch seed.Status {
		case "Fully Done":
			row.Done = true
		case "Working":
			row.Pending = true
		default:
			row.Cancelled = true
		}

		rows = append(rows, row)
	}

	return rows
}

// dataHoldingSeed is one curated position. The average cost is chosen so the
// gains add up to roughly the portfolio's lifetime P&L.
type dataHoldingSeed struct {
	Symbol   string
	Qty, Avg float64
}

// dataBuildHoldings turns the curated quantities into position rows. Bar is
// the position's value as a share of the largest holding.
func dataBuildHoldings(specs []dataWatchSpec, symbols map[string]SymbolData) []Holding {
	seeds := []dataHoldingSeed{
		{Symbol: "NVDA", Qty: 4200, Avg: 148.66},
		{Symbol: "AAPL", Qty: 3100, Avg: 215.24},
		{Symbol: "TSLA", Qty: 1250, Avg: 397.62},
		{Symbol: "MSFT", Qty: 1050, Avg: 445.14},
		{Symbol: "AMD", Qty: 2400, Avg: 172.98},
		{Symbol: "META", Qty: 620, Avg: 634.42},
		{Symbol: "AMZN", Qty: 1600, Avg: 204.27},
		{Symbol: "ARM", Qty: 3500, Avg: 155.07},
	}
	const invested = 4546557.0
	specBySymbol := make(map[string]dataWatchSpec, len(specs))

	for _, spec := range specs {
		specBySymbol[spec.Symbol] = spec
	}

	values := make([]float64, len(seeds))
	maxValue := 0.0

	for index, seed := range seeds {
		values[index] = seed.Qty * specBySymbol[seed.Symbol].Price
		maxValue = max(maxValue, values[index])
	}

	rows := make([]Holding, 0, len(seeds))

	for index, seed := range seeds {
		spec := specBySymbol[seed.Symbol]
		value := values[index]
		cost := seed.Qty * seed.Avg
		profit := value - cost
		weight := value / invested * 100
		bar := weight
		if maxValue > 0 {
			bar = value / maxValue * 100
		}

		rows = append(rows, Holding{
			Symbol: seed.Symbol,
			Name:   spec.Name,
			Qty:    dataComma(seed.Qty, 0),
			Avg:    dataFmtPrice(seed.Avg, spec.Decimals),
			Last:   symbols[seed.Symbol].Quote.Price,
			Value:  dataMoney(value, 2),
			PNL:    dataSignedMoney(profit, 2),
			Pct:    dataSignedPct(profit / cost * 100),
			Weight: dataFmtPrice(weight, 1) + "%",
			Up:     profit >= 0,
			Bar:    bar,
		})
	}

	return rows
}

// dataBuildIndices is the footer index strip.
func dataBuildIndices() []IndexRow {
	return []IndexRow{
		{Name: "DOW", Value: dataComma(46204.40, 2), Change: dataSignedNum(174.38, 2), ChangePct: dataSignedPct(0.38), Up: true},
		{Name: "NASDAQ", Value: dataComma(22418.62, 2), Change: dataSignedNum(230.14, 2), ChangePct: dataSignedPct(1.04), Up: true},
		{Name: "S&P 500", Value: dataComma(6742.18, 2), Change: dataSignedNum(52.84, 2), ChangePct: dataSignedPct(0.79), Up: true},
		{Name: "Russell 2000", Value: dataComma(2486.12, 2), Change: dataSignedNum(-5.48, 2), ChangePct: dataSignedPct(-0.22), Up: false},
		{Name: "VIX", Value: dataFmtPrice(14.82, 2), Change: dataSignedNum(-0.46, 2), ChangePct: dataSignedPct(-3.01), Up: false},
	}
}

// dataBuildTickers is the scrolling tape. It reads the same quotes the
// watchlist prints.
func dataBuildTickers(specs []dataWatchSpec, symbols map[string]SymbolData) []TickerItem {
	symbolsOrder := []string{"NVDA", "AAPL", "TSLA", "MSFT", "AMD", "META", "AMZN", "GOOGL", "JPM", "BABA", "9988.HK", "SPY"}
	_ = specs
	out := make([]TickerItem, 0, len(symbolsOrder))

	for _, symbol := range symbolsOrder {
		quote := symbols[symbol].Quote
		out = append(out, TickerItem{
			Symbol:    symbol,
			Price:     quote.Price,
			Change:    quote.Change,
			ChangePct: quote.ChangePct,
			Up:        quote.Up,
		})
	}

	return out
}

// dataDelay cycles the six entrance-delay classes across a heatmap.
func dataDelay(index int) string {
	return "delay-" + strconv.Itoa(index%6+1)
}

// dataSparkBars draws a deterministic mini sparkline. The bars rise and fall
// but stay inside the 20-100 percent band the component expects.
func dataSparkBars(rng *rand.Rand, count int) []SparkBar {
	out := make([]SparkBar, 0, count)
	previous := 45.0

	for range count {
		height := 20 + rng.Float64()*80
		out = append(out, SparkBar{Height: height, Up: height >= previous})
		previous = height
	}

	return out
}

// dataBuildMarkets assembles the market-overview screen.
func dataBuildMarkets(rng *rand.Rand) MarketsData {
	return MarketsData{
		Breadth: []BreadthStat{
			{Label: "Advance/decline", Value: "1.11", Note: "2,148 up · 1,932 down", Pct: 62, Up: true},
			{Label: "New highs", Value: "184", Note: "62 new lows", Pct: 74, Up: true},
			{Label: "Above 50DMA", Value: "68.4%", Note: "of S&P 500 members", Pct: 68, Up: true},
			{Label: "Up volume", Value: "61.2%", Note: "of NYSE volume", Pct: 61, Up: true},
		},
		Sessions: []SessionRow{
			{Name: "Asia", Region: "Tokyo · Hong Kong", Status: "Closed", Change: "+142.60", Pct: "+0.62%", Up: true, Open: false},
			{Name: "Europe", Region: "London · Frankfurt", Status: "Open", Change: "-18.40", Pct: "-0.14%", Up: false, Open: true},
			{Name: "United States", Region: "New York", Status: "Open", Change: "+154.82", Pct: "+0.79%", Up: true, Open: true},
		},
		Sectors: []SectorRow{
			{Name: "Technology", Pct: "+1.42%", Up: true, Weight: 31.2, Delay: dataDelay(0)},
			{Name: "Communication", Pct: "+0.86%", Up: true, Weight: 13.4, Delay: dataDelay(1)},
			{Name: "Consumer Disc.", Pct: "+0.24%", Up: true, Weight: 10.1, Delay: dataDelay(2)},
			{Name: "Financials", Pct: "-0.18%", Up: false, Weight: 13.2, Delay: dataDelay(3)},
			{Name: "Industrials", Pct: "+0.55%", Up: true, Weight: 8.4, Delay: dataDelay(4)},
			{Name: "Health Care", Pct: "-0.42%", Up: false, Weight: 11.6, Delay: dataDelay(5)},
			{Name: "Consumer Staples", Pct: "+0.11%", Up: true, Weight: 5.8, Delay: dataDelay(6)},
			{Name: "Energy", Pct: "-1.24%", Up: false, Weight: 3.1, Delay: dataDelay(7)},
			{Name: "Utilities", Pct: "+0.34%", Up: true, Weight: 2.4, Delay: dataDelay(8)},
			{Name: "Real Estate", Pct: "-0.66%", Up: false, Weight: 2.1, Delay: dataDelay(9)},
			{Name: "Materials", Pct: "-0.08%", Up: false, Weight: 1.9, Delay: dataDelay(10)},
		},
		Gainers: []MoverRow{
			{Symbol: "PLTR", Name: "Palantir Technologies", Price: "84.62", Pct: "+7.84%", Volume: "48.2M", Up: true},
			{Symbol: "SMCI", Name: "Super Micro Computer", Price: "42.18", Pct: "+6.21%", Volume: "31.4M", Up: true},
			{Symbol: "U", Name: "Unity Software", Price: "28.42", Pct: "+5.66%", Volume: "19.8M", Up: true},
			{Symbol: "RIVN", Name: "Rivian Automotive", Price: "14.86", Pct: "+3.41%", Volume: "41.8M", Up: true},
			{Symbol: "COIN", Name: "Coinbase Global", Price: "312.44", Pct: "+3.18%", Volume: "12.6M", Up: true},
		},
		Losers: []MoverRow{
			{Symbol: "ENPH", Name: "Enphase Energy", Price: "38.24", Pct: "-5.12%", Volume: "9.4M", Up: false},
			{Symbol: "LCID", Name: "Lucid Group", Price: "2.84", Pct: "-4.38%", Volume: "28.6M", Up: false},
			{Symbol: "NIO", Name: "NIO Inc.", Price: "5.12", Pct: "-3.62%", Volume: "36.2M", Up: false},
			{Symbol: "MSTR", Name: "MicroStrategy", Price: "302.18", Pct: "-2.94%", Volume: "7.8M", Up: false},
			{Symbol: "TSLA", Name: "Tesla, Inc.", Price: "428.62", Pct: "-1.92%", Volume: "96.8M", Up: false},
		},
		Calendar: []CalendarEvent{
			{Time: "08:30", Country: "US", Event: "CPI m/m", Actual: "+0.2%", Forecast: "+0.3%", Prior: "+0.4%", Impact: "High", High: true},
			{Time: "08:30", Country: "US", Event: "Initial jobless claims", Actual: "218K", Forecast: "224K", Prior: "221K", Impact: "Medium"},
			{Time: "10:00", Country: "US", Event: "ISM services PMI", Actual: "54.2", Forecast: "53.6", Prior: "52.8", Impact: "Medium"},
			{Time: "11:00", Country: "EU", Event: "ECB President speech", Actual: "—", Forecast: "—", Prior: "—", Impact: "High", High: true},
			{Time: "14:00", Country: "US", Event: "FOMC minutes", Actual: "—", Forecast: "—", Prior: "—", Impact: "Low"},
			{Time: "19:50", Country: "CN", Event: "Trade balance", Actual: "$82.4B", Forecast: "$78.9B", Prior: "$75.2B", Impact: "Low"},
		},
		Indices: []IndexCard{
			{Name: "S&P 500", Symbol: "SPX", Value: dataComma(6742.18, 2), Change: dataSignedNum(52.84, 2), Pct: dataSignedPct(0.79), Up: true, Spark: dataSparkBars(rng, 12)},
			{Name: "NASDAQ", Symbol: "IXIC", Value: dataComma(22418.62, 2), Change: dataSignedNum(230.14, 2), Pct: dataSignedPct(1.04), Up: true, Spark: dataSparkBars(rng, 12)},
			{Name: "Dow Jones", Symbol: "DJI", Value: dataComma(46204.40, 2), Change: dataSignedNum(174.38, 2), Pct: dataSignedPct(0.38), Up: true, Spark: dataSparkBars(rng, 12)},
			{Name: "Russell 2000", Symbol: "RUT", Value: dataComma(2486.12, 2), Change: dataSignedNum(-5.48, 2), Pct: dataSignedPct(-0.22), Up: false, Spark: dataSparkBars(rng, 12)},
		},
		Flows: []FlowRow{
			{Label: "Technology", Value: "+$1.82B", Up: true, Pct: 84},
			{Label: "Financials", Value: "+$642M", Up: true, Pct: 61},
			{Label: "Energy", Value: "-$318M", Up: false, Pct: 44},
			{Label: "Health care", Value: "+$214M", Up: true, Pct: 37},
			{Label: "Consumer", Value: "-$186M", Up: false, Pct: 31},
			{Label: "Utilities", Value: "-$94M", Up: false, Pct: 22},
		},
		Updated: dataMarketUpdated,
	}
}
