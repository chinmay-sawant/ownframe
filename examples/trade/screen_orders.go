package main

import "strings"

const (
	ordGroupUniverse = iota
	ordGroupPrice
	ordGroupTechnical
)

// ScreenerUniverseFilters returns the filter-builder chips for market, sector,
// and size filters.
func (v View) ScreenerUniverseFilters() []ScreenerFilter {
	groups := ordScreenerGroups(v.Screener.Filters)
	return groups[ordGroupUniverse]
}

// ScreenerPriceFilters returns the filter-builder chips for price and
// liquidity filters.
func (v View) ScreenerPriceFilters() []ScreenerFilter {
	groups := ordScreenerGroups(v.Screener.Filters)
	return groups[ordGroupPrice]
}

// ScreenerTechnicalFilters returns the filter-builder chips for trend and
// momentum filters.
func (v View) ScreenerTechnicalFilters() []ScreenerFilter {
	groups := ordScreenerGroups(v.Screener.Filters)
	return groups[ordGroupTechnical]
}

// ScreenerSignalClass returns the tone class for a screener signal tag.
func (v View) ScreenerSignalClass(signal string) string {
	s := strings.ToLower(signal)

	switch {
	case ordHasAny(s, "strong buy", "buy", "accumulate", "outperform", "overweight", "long", "bull"):
		return "up"
	case ordHasAny(s, "strong sell", "sell", "reduce", "underperform", "underweight", "avoid", "short", "bear"):
		return "down"
	case ordHasAny(s, "watch", "neutral", "hold", "mixed"):
		return "amber"
	}

	return "dim"
}

// ordScreenerGroups splits the flat filter list into the three builder cards.
// A filter with a recognisable key goes to its own card; any other filter fills
// the least populated card, so every chip shows exactly once.
func ordScreenerGroups(filters []ScreenerFilter) [3][]ScreenerFilter {
	groups := [3][]ScreenerFilter{}

	for _, filter := range filters {
		group := ordScreenerGroup(filter.Key)
		if group < 0 {
			group = 0
			for i := 1; i < len(groups); i++ {
				if len(groups[i]) < len(groups[group]) {
					group = i
				}
			}
		}

		groups[group] = append(groups[group], filter)
	}

	return groups
}

// ordScreenerGroup returns the builder card for a filter key, or -1 when the
// key names nothing it recognises.
func ordScreenerGroup(key string) int {
	key = strings.ToLower(key)

	switch key {
	case "market-cap", "sector", "region", "exchange", "industry", "country", "index":
		return ordGroupUniverse
	case "price-range", "avg-volume", "turnover", "spread", "liquidity":
		return ordGroupPrice
	case "rsi", "trend", "macd", "volatility", "momentum":
		return ordGroupTechnical
	}

	switch {
	case ordHasAny(key, "price", "volume", "liquidity", "turnover", "spread", "beta", "float"):
		return ordGroupPrice
	case ordHasAny(key, "rsi", "macd", "sma", "ema", "trend", "momentum", "signal", "technical", "volatility", "atr", "breakout", "gap"):
		return ordGroupTechnical
	case ordHasAny(key, "market", "exchange", "region", "sector", "industry", "country", "index", "universe", "asset", "cap", "valuation", "eps", "growth", "dividend"):
		return ordGroupUniverse
	}

	return -1
}

// ordHasAny reports whether s contains one of the fragments.
func ordHasAny(s string, fragments ...string) bool {
	for _, fragment := range fragments {
		if strings.Contains(s, fragment) {
			return true
		}
	}

	return false
}
