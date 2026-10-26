package main

import (
	"reflect"
	"testing"
)

// TestDemoDeterministic checks the seeded generators produce the same market
// on two runs independent of the display clock.
func TestDemoDeterministic(t *testing.T) {
	first := newDemo()
	second := newDemo()

	if !reflect.DeepEqual(first.Symbols, second.Symbols) {
		t.Fatal("symbol snapshots differ between runs")
	}

	if !reflect.DeepEqual(first.Watch, second.Watch) {
		t.Fatal("watchlist differs between runs")
	}

	if !reflect.DeepEqual(first.Orders, second.Orders) {
		t.Fatal("orders differ between runs")
	}

	if len(first.Watch) != 16 {
		t.Fatalf("watchlist = %d, want 16", len(first.Watch))
	}
}

// TestCandleAndBookInvariants checks every generated quote is internally
// consistent enough for the chart and depth table.
func TestCandleAndBookInvariants(t *testing.T) {
	demo := newDemo()

	for symbol, data := range demo.Symbols {
		if data.Quote.Symbol != symbol {
			t.Fatalf("%s quote symbol = %q", symbol, data.Quote.Symbol)
		}

		if len(data.Candles) != dataCandleCount {
			t.Fatalf("%s candles = %d, want %d", symbol, len(data.Candles), dataCandleCount)
		}

		for index, candle := range data.Candles {
			if candle.High < candle.Open || candle.High < candle.Close {
				t.Fatalf("%s candle %d high %.2f below open/close", symbol, index, candle.High)
			}

			if candle.Low > candle.Open || candle.Low > candle.Close {
				t.Fatalf("%s candle %d low %.2f above open/close", symbol, index, candle.Low)
			}

			if candle.Volume < 0 {
				t.Fatalf("%s candle %d volume %.0f", symbol, index, candle.Volume)
			}
		}

		if len(data.Book) != dataBookDepth*2 {
			t.Fatalf("%s book = %d, want %d", symbol, len(data.Book), dataBookDepth*2)
		}

		for index := 0; index < dataBookDepth; index++ {
			if data.Book[index].Side != "bid" || data.Book[index+dataBookDepth].Side != "ask" {
				t.Fatalf("%s book sides are not bid then ask", symbol)
			}
		}
	}
}

// TestDonutGeometry checks the transaction ring has one arc per legend row.
func TestDonutGeometry(t *testing.T) {
	demo := newDemo()

	for symbol, data := range demo.Symbols {
		if len(data.Donut.Segments) != 6 || len(data.DonutLegend) != 6 {
			t.Fatalf("%s donut segments/legend = %d/%d", symbol, len(data.Donut.Segments), len(data.DonutLegend))
		}

		for index, segment := range data.Donut.Segments {
			if segment.Path == "" {
				t.Fatalf("%s donut segment %d has no path", symbol, index)
			}
		}
	}
}

// TestOrdersAndScreenData checks the blotter, holdings, and screen fixtures
// are populated and consistent.
func TestOrdersAndScreenData(t *testing.T) {
	demo := newDemo()

	if len(demo.Orders) != 18 {
		t.Fatalf("orders = %d, want 18", len(demo.Orders))
	}

	for _, order := range demo.Orders {
		if order.Status == "" || order.Symbol == "" {
			t.Fatalf("incomplete order: %+v", order)
		}

		if order.Done && order.Pending {
			t.Fatalf("order %s is both done and working", order.ID)
		}
	}

	if len(demo.Holdings) != len(demo.Portfolio.Holdings) {
		t.Fatalf("holdings disagree: %d vs %d", len(demo.Holdings), len(demo.Portfolio.Holdings))
	}

	if len(demo.Markets.Sectors) < 9 {
		t.Fatalf("sector heatmap = %d tiles", len(demo.Markets.Sectors))
	}

	if len(demo.Screener.Rows) == 0 || len(demo.Screener.Filters) < 9 {
		t.Fatalf("screener rows/filters = %d/%d", len(demo.Screener.Rows), len(demo.Screener.Filters))
	}

	if len(demo.News.Items) == 0 || len(demo.Settings.Sections) == 0 {
		t.Fatal("news or settings fixtures are empty")
	}
}
