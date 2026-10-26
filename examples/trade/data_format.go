//nolint:mnd // The sample market fixture prints literal prices, sizes, and percentages; those are display content, not tunable constants.
package main

import (
	"math"
	"strconv"
	"strings"
)

const (
	dataSeed          = 20261025
	dataFakerSeed     = 8675309
	dataMarketUpdated = "10:42:18 ET"
	dataCandleCount   = 180
	dataBookDepth     = 5
	dataTradeCount    = 12
	dataOptionCount   = 7
)

// dataFmtPrice renders a value with a fixed number of decimals and no
// thousands separator. Prices use this shape.
func dataFmtPrice(v float64, decimals int) string {
	return strconv.FormatFloat(v, 'f', decimals, 64)
}

// dataFixed renders a non-price figure such as a donut total. It keeps the
// same plain shape as dataFmtPrice so callers can add their own unit.
func dataFixed(v float64, decimals int) string {
	return strconv.FormatFloat(v, 'f', decimals, 64)
}

// dataSignedNum renders a signed number with a leading plus for gains.
func dataSignedNum(v float64, decimals int) string {
	text := dataFmtPrice(math.Abs(v), decimals)
	if v < 0 {
		return "-" + text
	}

	return "+" + text
}

// dataSignedPct renders a signed percentage with two decimals.
func dataSignedPct(v float64) string {
	return dataSignedNum(v, 2) + "%"
}

// dataShortNum compresses a large value into K, M, B, or T units.
func dataShortNum(v float64) string {
	abs := math.Abs(v)

	switch {
	case abs >= 1e12:
		return dataFmtPrice(v/1e12, 2) + "T"
	case abs >= 1e9:
		return dataFmtPrice(v/1e9, 2) + "B"
	case abs >= 1e6:
		return dataFmtPrice(v/1e6, 2) + "M"
	case abs >= 1e3:
		return dataFmtPrice(v/1e3, 1) + "K"
	default:
		return dataFmtPrice(v, 0)
	}
}

// dataComma groups the integer part of a number with commas.
func dataComma(v float64, decimals int) string {
	text := strconv.FormatFloat(math.Abs(v), 'f', decimals, 64)

	whole, fraction, found := strings.Cut(text, ".")
	if found {
		fraction = "." + fraction
	}

	var b strings.Builder
	if v < 0 {
		b.WriteByte('-')
	}

	start := len(whole) % 3
	if start == 0 {
		start = 3
	}

	b.WriteString(whole[:start])
	for i := start; i < len(whole); i += 3 {
		b.WriteByte(',')
		b.WriteString(whole[i : i+3])
	}

	b.WriteString(fraction)

	return b.String()
}

// dataMoney renders a value with a currency mark and grouped digits.
func dataMoney(v float64, decimals int) string {
	return "$" + dataComma(v, decimals)
}

// dataSignedMoney renders a gain or loss as signed currency.
func dataSignedMoney(v float64, decimals int) string {
	sign := "+"
	if v < 0 {
		sign = "-"
	}

	return sign + "$" + dataComma(math.Abs(v), decimals)
}

// dataCurrencyMark is the display prefix for a traded currency.
func dataCurrencyMark(currency string) string {
	if currency == "HKD" {
		return "HK$"
	}

	return "$"
}

// dataPad2 renders a two-digit clock field.
func dataPad2(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}

	return strconv.Itoa(v)
}

// dataClockSeconds renders an absolute second-of-day as HH:MM:SS.
func dataClockSeconds(total int) string {
	hours := total / 3600
	minutes := total % 3600 / 60
	seconds := total % 60

	return dataPad2(hours) + ":" + dataPad2(minutes) + ":" + dataPad2(seconds)
}

// dataClockMinute renders a minute-of-day as HH:MM.
func dataClockMinute(total int) string {
	return dataPad2(total/60) + ":" + dataPad2(total%60)
}

// dataCandleClock labels one candle of an intraday session. The last bar
// carries the closing bell so the series ends at the quote timestamp.
func dataCandleClock(index, count int) string {
	if index == count-1 {
		return "16:00"
	}

	return dataClockMinute(9*60 + 30 + index*2)
}

// dataLogoClass turns a ticker into the lowercase CSS class that paints its
// monogram tile.
func dataLogoClass(symbol string) string {
	class := strings.ToLower(symbol)
	class = strings.ReplaceAll(class, ".", "-")

	return "logo-" + class
}

// dataParseFloat reads a formatted price back into a number. Fixture code
// only uses it where a display string is the source of truth.
