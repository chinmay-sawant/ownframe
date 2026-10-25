package main

import (
	"math"
	"strconv"
	"strings"
)

// The chart uses a 1000 by 400 design space expressed as percentages. Candles
// and their moving averages share the upper band, volume columns the lower
// band, and the summary, axis labels, and price tags are HTML around the same
// box. The HTML template paints the spans; Go only maps data to positions.
const (
	chartViewW = 1000.0
	chartViewH = 400.0

	chartCandleX0    = 48.0
	chartCandleX1    = 976.0
	chartPriceTop    = 24.0
	chartPriceBottom = 300.0
	chartVolumeTop   = 318.0
	chartVolumeBot   = 390.0

	chartPriceRows = 5

	chartPadShare     = 0.04
	chartBodyShare    = 0.62
	chartBodyMinWidth = 5.0
	chartBodyMaxWidth = 7.0
	chartBodyMinH     = 1.6
	chartBarCenter    = 0.5
	chartVolumeMinH   = 1.0
	chartVolumeLift   = 2.0
	chartMAWidth      = 1.4

	chartWindow1D  = 40
	chartWindow5D  = 55
	chartWindow1M  = 75
	chartWindow3M  = 100
	chartWindowMax = 140

	chartMA5   = 5
	chartMA8   = 8
	chartMA13  = 13
	chartMA21  = 21
	chartMA34  = 34
	chartMA50  = 50
	chartMA100 = 100

	chartXLabels        = 6
	chartXAxisStart     = 4.4
	chartXAxisStep      = 16.32
	chartCrossXShare    = 0.68
	chartCrossShare     = 0.55
	chartBitSize        = 64
	chartCrossIndexN    = 2
	chartCrossIndexD    = 3
	chartPercentBase    = 100.0
	chartPercentDigits  = 1
	chartPositionDigits = 2
	chartWholeDigits    = 0

	chartLineMinPts = 2
	chartToneStride = 2

	chartLabelBigPrice      = 1000.0
	chartLabelBigDecimals   = 0
	chartLabelSmallDecimals = 2
	chartGroupSize          = 3

	chartThousand = 1e3
	chartMillion  = 1e6
	chartBillion  = 1e9

	chartLogoRunes = 2
	chartTimeRunes = 5
)

// prepareChart turns generated candles into data-derived positions and labels
// for the CSS plot. The HTML template paints the spans; Go does not draw the
// chart or start its animation. The last N candles of the range feed the plot
// so switching timeframe never changes the card height.
func prepareChart(quote Quote, candles []Candle, timeframe string) Chart {
	window := chartWindow(candles, timeframe)
	chart := Chart{
		Summary:   chartSummary(quote, timeframe),
		LastPrice: quote.Price,
		CrossX:    strconv.FormatFloat(chartCrossXShare*chartPercentBase, 'f', chartWholeDigits, 64),
		CrossY:    "37",
		Up:        quote.Up,
	}

	if len(window) == 0 {
		chart.LastTop = chart.CrossY
		return chart
	}

	high, low := chartRange(window)
	if high <= low {
		base := window[len(window)-1].Close
		if parsed, ok := chartParsePrice(quote.Price); ok {
			base = parsed
		}
		if base == 0 {
			base = 1
		}
		high = base * (1 + chartPadShare)
		low = base * (1 - chartPadShare)
	}

	span := high - low
	if span <= 0 {
		span = high
		low = high - span
	}

	pad := span * chartPadShare
	maxPrice := high + pad
	minPrice := low - pad
	priceSpan := maxPrice - minPrice
	if priceSpan <= 0 {
		priceSpan = 1
	}

	y := func(price float64) float64 {
		return chartPriceTop + (maxPrice-price)/priceSpan*(chartPriceBottom-chartPriceTop)
	}

	if chart.Summary.Open == "" {
		chart.Summary.Open = chartLabel(window[0].Open)
	}
	if chart.Summary.High == "" {
		chart.Summary.High = chartLabel(high)
	}
	if chart.Summary.Low == "" {
		chart.Summary.Low = chartLabel(low)
	}
	if chart.Summary.Close == "" {
		chart.Summary.Close = chartLabel(window[len(window)-1].Close)
	}

	totalVolume := 0.0
	for _, candle := range window {
		totalVolume += candle.Volume
	}
	if chart.Summary.Volume == "" && totalVolume > 0 {
		chart.Summary.Volume = chartVolumeLabel(totalVolume)
	}

	lastClose := window[len(window)-1].Close
	chart.LastTop = chartPosition(y(lastClose))
	chart.HighTop = chartPosition(y(high))
	chart.HighLabel = chartLabel(high)
	chart.LowTop = chartPosition(y(low))
	chart.LowLabel = chartLabel(low)

	chart.Candles, chart.Volume = chartBars(window, y)
	step := (chartCandleX1 - chartCandleX0) / float64(len(window))
	chart.SlotLead = chartPercent(chartCandleX0-step/2, chartViewW)
	chart.SlotWidth = chartPercent(step, chartViewW)
	chart.Lines = chartLines(window, timeframe, y)
	chart.YLabels = chartYLabels(maxPrice, priceSpan)
	chart.XLabels = chartXLabelsFor(window, timeframe)

	crossPrice := low + (high-low)*chartCrossShare
	crossY := y(crossPrice)
	chart.CrossY = chartPosition(crossY)
	chart.CrossPrice = chartLabel(crossPrice)
	chart.CrossTime = chartCrossTime(window)
	if chart.CrossTime == "" {
		chart.CrossTime = chartXLabelAt(chart.XLabels, chartCrossXShare)
	}

	return chart
}

// ChartLogoText is the two-letter monogram in the quote header. It is derived
// from the selected symbol so the data type needs no extra field.
func (v View) ChartLogoText() string {
	symbol := strings.TrimSpace(v.Quote.Symbol)
	if symbol == "" {
		return "M"
	}

	runes := []rune(symbol)
	if len(runes) < chartLogoRunes {
		return strings.ToUpper(symbol)
	}

	return strings.ToUpper(string(runes[:chartLogoRunes]))
}

// chartWindow picks the trailing candles for one range.
func chartWindow(candles []Candle, timeframe string) []Candle {
	size := chartWindowSize(timeframe)
	if size > len(candles) {
		size = len(candles)
	}
	if size <= 0 {
		return nil
	}

	return candles[len(candles)-size:]
}

// chartWindowSize is the candle count each timeframe shows.
func chartWindowSize(timeframe string) int {
	switch timeframe {
	case "1D":
		return chartWindow1D
	case "5D":
		return chartWindow5D
	case "1M":
		return chartWindow1M
	case "3M":
		return chartWindow3M
	default:
		return chartWindowMax
	}
}

// chartRange returns the high and low of one candle window.
func chartRange(candles []Candle) (float64, float64) {
	high, low := candles[0].High, candles[0].Low

	for _, candle := range candles {
		if candle.High > high {
			high = candle.High
		}
		if candle.Low < low {
			low = candle.Low
		}
	}

	return high, low
}

// chartBars lays out the candle bodies, the volume columns, and the wick
// geometry in plot pixels.
func chartBars(candles []Candle, y func(float64) float64) ([]CandleBar, []VolumeBar) {
	count := len(candles)
	step := (chartCandleX1 - chartCandleX0) / float64(count)
	width := step * chartBodyShare
	if width < chartBodyMinWidth {
		width = chartBodyMinWidth
	}
	if width > chartBodyMaxWidth {
		width = chartBodyMaxWidth
	}
	width = math.Round(width)

	maxVolume := 0.0
	for _, candle := range candles {
		if candle.Volume > maxVolume {
			maxVolume = candle.Volume
		}
	}

	bars := make([]CandleBar, 0, count)
	volumes := make([]VolumeBar, 0, count)

	for i, candle := range candles {
		x := chartCandleX0 + step*(float64(i)+chartBarCenter)
		openY := y(candle.Open)
		closeY := y(candle.Close)
		bodyTop := math.Min(openY, closeY)
		bodyHeight := math.Abs(openY - closeY)
		if bodyHeight < chartBodyMinH {
			bodyHeight = chartBodyMinH
		}

		up := candle.Close >= candle.Open
		wickTop := y(candle.High)
		wickBottom := y(candle.Low)
		bars = append(bars, CandleBar{
			X:          x,
			Left:       x - width/2,
			Width:      width,
			WickTop:    wickTop,
			WickBottom: wickBottom,
			WickOffset: wickTop - bodyTop,
			WickHeight: wickBottom - wickTop,
			BodyTop:    bodyTop,
			Height:     bodyHeight,
			SlotWidth:  step,
			Up:         up,
		})

		volumeHeight := 0.0
		if maxVolume > 0 {
			volumeHeight = candle.Volume / maxVolume * (chartVolumeBot - chartVolumeTop)
		}
		if volumeHeight < chartVolumeMinH {
			volumeHeight = chartVolumeMinH
		}
		volumes = append(volumes, VolumeBar{
			X:      x - chartVolumeLift,
			Top:    chartVolumeBot - volumeHeight,
			Height: volumeHeight,
			Up:     up,
		})
	}

	return bars, volumes
}

// chartLines builds the moving averages over the visible window.
func chartLines(candles []Candle, timeframe string, y func(float64) float64) []ChartLine {
	periods := chartMAPeriods(timeframe)
	colors := []string{"#64a8ff", "#c084fc", "#f0b75a", "#8b95a5"}
	lines := make([]ChartLine, 0, len(periods))
	step := (chartCandleX1 - chartCandleX0) / float64(len(candles))

	for i, period := range periods {
		if period > len(candles) {
			continue
		}

		points := make([]string, 0, len(candles)-period+1)
		for end := period - 1; end < len(candles); end++ {
			sum := 0.0
			for start := end - period + 1; start <= end; start++ {
				sum += candles[start].Close
			}

			x := chartCandleX0 + step*(float64(end)+chartBarCenter)
			points = append(points, chartPoint(x, y(sum/float64(period))))
		}

		if len(points) < chartLineMinPts {
			continue
		}

		firstX := chartCandleX0 + step*(float64(period-1)+chartBarCenter)
		line := ChartLine{
			Points:   strings.Join(points, " "),
			Color:    colors[i%len(colors)],
			Width:    chartMAWidth,
			SlotLead: chartPercent(firstX-step/2, chartViewW),
		}
		line.Segments = chartSegments(line.Points, line.Color, line.Width)
		lines = append(lines, line)
	}

	return lines
}

// chartMAPeriods is the moving-average set for one range. Short ranges skip
// periods they cannot fill, so every line has at least two points.
func chartMAPeriods(timeframe string) []int {
	switch timeframe {
	case "1D":
		return []int{chartMA5, chartMA8, chartMA21}
	case "5D":
		return []int{chartMA5, chartMA13, chartMA21, chartMA34}
	case "1M":
		return []int{chartMA8, chartMA21, chartMA34, chartMA50}
	case "3M":
		return []int{chartMA8, chartMA21, chartMA50}
	default:
		return []int{chartMA21, chartMA50, chartMA100}
	}
}

// chartYLabels aligns the price labels with the five horizontal grid lines.
func chartYLabels(maxPrice, priceSpan float64) []ChartLabel {
	labels := make([]ChartLabel, 0, chartPriceRows)

	for i := 0; i < chartPriceRows; i++ {
		price := maxPrice - priceSpan*float64(i)/float64(chartPriceRows-1)
		lineY := chartPriceTop + (chartPriceBottom-chartPriceTop)*float64(i)/float64(chartPriceRows-1)

		tone := "dim"
		if i%chartToneStride == 0 {
			tone = "muted"
		}

		labels = append(labels, ChartLabel{
			Top:  chartPercent(lineY, chartViewH),
			Left: "0%",
			Text: chartLabel(price),
			Tone: tone,
		})
	}

	return labels
}

// chartXLabelsFor labels the bottom of the plot at six fixed positions.
func chartXLabelsFor(candles []Candle, timeframe string) []ChartLabel {
	fallback := chartXLabelFallback(timeframe)
	labels := make([]ChartLabel, 0, chartXLabels)
	count := len(candles)

	for i := 0; i < chartXLabels; i++ {
		index := 0
		if count > 1 {
			index = int(math.Round(float64(i) * float64(count-1) / float64(chartXLabels-1)))
		}

		text := strings.TrimSpace(candles[index].Time)
		if text == "" {
			text = fallback[i%len(fallback)]
		}
		runes := []rune(text)
		if len(runes) > chartTimeRunes {
			text = string(runes[:chartTimeRunes])
		}

		left := chartXAxisStart + chartXAxisStep*float64(i)
		labels = append(labels, ChartLabel{
			Top:  "94%",
			Left: strconv.FormatFloat(left, 'f', 1, 64) + "%",
			Text: text,
			Tone: "dim",
		})
	}

	return labels
}

// chartXLabelFallback is used when the generated candles carry no time.
func chartXLabelFallback(timeframe string) []string {
	switch timeframe {
	case "5D":
		return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	case "1M":
		return []string{"W1", "W2", "W3", "W4", "W5", "W6"}
	case "3M":
		return []string{"W1", "W2", "W3", "W4", "W5", "W6"}
	case "1Y", "ALL":
		return []string{"Feb", "Apr", "Jun", "Aug", "Oct", "Dec"}
	default:
		return []string{"09", "10", "11", "12", "13", "14"}
	}
}

// chartCrossTime is the timestamp under the static crosshair, near two thirds
// of the visible window.
func chartCrossTime(candles []Candle) string {
	index := int(math.Round(float64(len(candles)-1) * float64(chartCrossIndexN) / float64(chartCrossIndexD)))
	if index < 0 {
		index = 0
	}
	if index >= len(candles) {
		index = len(candles) - 1
	}

	return strings.TrimSpace(candles[index].Time)
}

// chartXLabelAt picks the x label nearest a horizontal share of the plot.
func chartXLabelAt(labels []ChartLabel, share float64) string {
	if len(labels) == 0 {
		return ""
	}

	index := int(math.Round(share * float64(len(labels)-1)))
	if index < 0 {
		index = 0
	}
	if index >= len(labels) {
		index = len(labels) - 1
	}

	return labels[index].Text
}

// chartSummary formats the strip above the plot.
func chartSummary(quote Quote, timeframe string) ChartSummary {
	return ChartSummary{
		Period: chartPeriodLabel(timeframe),
		Change: chartSigned(quote.Change, quote.Up),
		Pct:    chartSigned(quote.ChangePct, quote.Up),
		Open:   quote.Open,
		High:   quote.High,
		Low:    quote.Low,
		Close:  quote.Price,
		Volume: quote.Volume,
		Up:     quote.Up,
	}
}

// chartPeriodLabel spells out the selected range.
func chartPeriodLabel(timeframe string) string {
	switch timeframe {
	case "5D":
		return "5 day · 5D"
	case "1M":
		return "1 month · 1M"
	case "3M":
		return "3 month · 3M"
	case "1Y":
		return "1 year · 1Y"
	case "ALL":
		return "All time · ALL"
	default:
		return "Intraday · 1D"
	}
}

// chartSigned keeps an authored sign and adds one when the feed omitted it.
func chartSigned(text string, up bool) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "+") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "−") {
		return trimmed
	}

	sign := "+"
	if !up {
		sign = "-"
	}

	return sign + trimmed
}

// chartPoint formats one polyline vertex.
func chartPoint(x, y float64) string {
	return strconv.FormatFloat(x, 'f', 1, 64) + "," + strconv.FormatFloat(y, 'f', 1, 64)
}

// chartPosition is a vertical offset in percent of the 400 unit design space.
func chartPosition(y float64) string {
	return strconv.FormatFloat(y/chartViewH*chartPercentBase, 'f', chartPositionDigits, 64)
}

// chartPercent formats a share of total with a trailing percent sign.
func chartPercent(value, total float64) string {
	if total == 0 {
		total = 1
	}

	return strconv.FormatFloat(value/total*chartPercentBase, 'f', chartPercentDigits, 64) + "%"
}

// chartLabel formats a price for an axis or a tag.
func chartLabel(price float64) string {
	decimals := chartLabelSmallDecimals
	if math.Abs(price) >= chartLabelBigPrice {
		decimals = chartLabelBigDecimals
	}

	return chartGroup(strconv.FormatFloat(price, 'f', decimals, 64))
}

// chartPriceText strips the currency mark and separators from a display price.
func chartPriceText(text string) string {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimPrefix(cleaned, "HK$")
	cleaned = strings.TrimPrefix(cleaned, "$")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// chartParsePrice reads a display price such as "$1,204.50" as a number.
func chartParsePrice(text string) (float64, bool) {
	cleaned := chartPriceText(text)
	if cleaned == "" {
		return 0, false
	}

	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, false
	}

	return value, true
}

// chartGroup inserts thousands separators into a formatted number.
func chartGroup(text string) string {
	dot := strings.IndexByte(text, '.')
	integer, fraction := text, ""
	if dot >= 0 {
		integer, fraction = text[:dot], text[dot:]
	}

	negative := strings.HasPrefix(integer, "-")
	if negative {
		integer = integer[1:]
	}

	if len(integer) <= chartGroupSize {
		return text
	}

	var out strings.Builder
	first := len(integer) % chartGroupSize
	if first == 0 {
		first = chartGroupSize
	}

	out.WriteString(integer[:first])
	for i := first; i < len(integer); i += chartGroupSize {
		out.WriteByte(',')
		out.WriteString(integer[i : i+chartGroupSize])
	}

	grouped := out.String()
	if negative {
		grouped = "-" + grouped
	}

	return grouped + fraction
}

// chartVolumeLabel compresses a share count for the summary fallback.
func chartVolumeLabel(volume float64) string {
	switch {
	case volume >= chartBillion:
		return strconv.FormatFloat(volume/chartBillion, 'f', 1, 64) + "B"
	case volume >= chartMillion:
		return strconv.FormatFloat(volume/chartMillion, 'f', 1, 64) + "M"
	case volume >= chartThousand:
		return strconv.FormatFloat(volume/chartThousand, 'f', 1, 64) + "K"
	default:
		return strconv.FormatFloat(volume, 'f', 0, 64)
	}
}

// chartSegments splits an x,y point string into CSS line segments. The source
// points stay in the chart data; the HTML fragment only draws the spans.
func chartSegments(points, color string, width float64) []ChartSegment {
	fields := strings.Fields(points)
	if len(fields) < chartLineMinPts {
		return nil
	}

	out := make([]ChartSegment, 0, len(fields)-1)
	prevX, prevY, ok := chartParsePoint(fields[0])
	if !ok {
		return nil
	}

	for _, field := range fields[1:] {
		x, y, valid := chartParsePoint(field)
		if !valid {
			continue
		}

		dx := x - prevX
		dy := y - prevY
		length := math.Hypot(dx, dy)
		if length > 0 {
			out = append(out, ChartSegment{
				Left:   prevX,
				Top:    prevY,
				Length: length,
				Angle:  math.Atan2(dy, dx) * 180 / math.Pi,
				Color:  color,
				Width:  width,
			})
		}

		prevX, prevY = x, y
	}

	return out
}

// chartParsePoint parses one "x,y" pair.
func chartParsePoint(pair string) (float64, float64, bool) {
	left, right, ok := strings.Cut(pair, ",")
	if !ok {
		return 0, 0, false
	}

	x, errX := strconv.ParseFloat(left, 64)
	y, errY := strconv.ParseFloat(right, 64)
	if errX != nil || errY != nil {
		return 0, 0, false
	}

	return x, y, true
}

// chartPctText formats one design axis value as a percentage.
func chartPctText(value, total float64) string {
	if total == 0 {
		total = 1
	}

	return strconv.FormatFloat(value/total*chartPercentBase, 'f', chartPositionDigits, 64) + "%"
}

// chartGrow formats one design-space value as a flex-grow weight. Only the
// ratio of the weights on one flex line matters, so the value is used as is.
func chartGrow(value float64) string {
	return strconv.FormatFloat(value, 'f', chartPositionDigits, 64)
}

// GrowBodyTop is the space above the candle body.
func (c CandleBar) GrowBodyTop() string { return chartGrow(c.BodyTop) }

// GrowBody is the candle body height.
func (c CandleBar) GrowBody() string { return chartGrow(c.Height) }

// GrowBodyBottom is the space below the candle body.
func (c CandleBar) GrowBodyBottom() string {
	return chartGrow(chartViewH - c.BodyTop - c.Height)
}

// GrowWickTop is the space above the wick.
func (c CandleBar) GrowWickTop() string { return chartGrow(c.WickTop) }

// GrowWick is the wick height.
func (c CandleBar) GrowWick() string { return chartGrow(c.WickHeight) }

// GrowWickBottom is the space below the wick.
func (c CandleBar) GrowWickBottom() string {
	return chartGrow(chartViewH - c.WickTop - c.WickHeight)
}

// BodyShare is the body width as a percentage of one candle slot.
func (c CandleBar) BodyShare() string {
	if c.SlotWidth == 0 {
		return "60.00%"
	}

	return strconv.FormatFloat(c.Width/c.SlotWidth*chartPercentBase, 'f', chartPositionDigits, 64) + "%"
}

// GrowTop is the space above the volume column.
func (v VolumeBar) GrowTop() string { return chartGrow(v.Top) }

// GrowBar is the volume column height.
func (v VolumeBar) GrowBar() string { return chartGrow(v.Height) }

// GrowBottom is the space below the volume column.
func (v VolumeBar) GrowBottom() string {
	return chartGrow(chartViewH - v.Top - v.Height)
}

// GrowTop is the space above one moving-average point.
func (s ChartSegment) GrowTop() string { return chartGrow(s.Top) }

// GrowBottom is the space below one moving-average point.
func (s ChartSegment) GrowBottom() string { return chartGrow(chartViewH - s.Top) }

// RestTop returns the space below a top percentage as a flex weight, so an
// overlay row can sit at the price it marks.
func (c Chart) RestTop(top string) string {
	value, err := strconv.ParseFloat(strings.TrimSuffix(top, "%"), chartBitSize)
	if err != nil {
		return "50"
	}

	return chartGrow(chartPercentBase - value)
}

// RestLeft returns the space right of a left percentage as a flex weight.
func (c Chart) RestLeft(left string) string {
	value, err := strconv.ParseFloat(strings.TrimSuffix(left, "%"), chartBitSize)
	if err != nil {
		return "50"
	}

	return chartGrow(chartPercentBase - value)
}
