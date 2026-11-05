package main

import "fmt"

type sku struct {
	Name       string
	Code       string
	Category   string
	Status     string
	OnHand     int
	Inbound    int
	DaysLeft   int
	ReorderPct float64
	StockPct   float64
}

//nolint:gochecknoglobals // static inventory fixture
var sampleSKUs = []sku{
	{"Cold brew filter packs, 50 ct", "NW-CF-2210", "Coffee & tea", "Stockout", 0, 1200, 0, 0.05, 0.05},
	{"Ceremonial matcha, 1 kg", "NW-CF-3302", "Coffee & tea", "Stockout", 0, 200, 0, 0.12, 0.12},
	{"Oat milk barista, 32 oz x 12", "NW-DA-0418", "Dairy alternatives", "Reorder", 86, 600, 3, 0.35, 0.3},
	{"Compostable cups 12 oz, case 1,000", "NW-PK-1120", "Packaging", "Reorder", 412, -1, 9, 0.55, 0.3},
	{"House espresso blend, 5 lb", "NW-CF-1001", "Coffee & tea", "Healthy", 640, 480, 24, 0.6, 0.55},
	{"Nitrile gloves, large, box 100", "NW-JS-5510", "Janitorial", "Healthy", 1120, -1, 33, 0.65, 0.6},
	{"Paper straws 7.75 in, case 5,000", "NW-PK-0907", "Packaging", "Healthy", 1940, -1, 41, 0.85, 0.65},
	{"Vanilla syrup, 750 ml", "NW-SY-0031", "Syrups", "Overstock", 2380, -1, 122, 0.85, 0.85},
}

type inboundPO struct {
	Vendor string
	PONo   string
	Detail string
	ETA    string
	Amount string
	Badge  string
}

//nolint:gochecknoglobals // static inventory fixture
var samplePOs = []inboundPO{
	{"Pacific Bean", "PO-10482", "Dock 3 \u00b7 1680 units", "Today 2:00 PM", "$48,600", "Arriving today"},
	{"Oat Co. Foodservice", "PO-10477", "truck 62% of route \u00b7 600 cases", "Sat Nov 1", "$21,300", "In transit"},
	{"Uji Matcha Trading", "PO-10469", "Port of Oakland \u00b7 200 kg", "Thu Nov 6", "$38,920", "On the water"},
	{"GreenWare Packaging", "PO-10491", "rail Chicago \u00b7 2400 cases", "Mon Nov 10", "$77,600", "In transit"},
}

type binRow struct {
	Level string
	Cells [10]int
}

//nolint:gochecknoglobals // static inventory fixture
var sampleBins = []binRow{
	{"L5", [10]int{42, 55, 61, 38, 70, 66, 48, 35, 52, 28}},
	{"L4", [10]int{68, 74, 80, 62, 88, 85, 71, 59, 73, 47}},
	{"L3", [10]int{81, 86, 92, 77, 95, 91, 83, 72, 84, 63}},
	{"L2", [10]int{88, 93, 97, 85, 99, 96, 90, 81, 89, 70}},
	{"L1", [10]int{94, 98, 100, 91, 100, 99, 95, 87, 93, 76}},
}

type catSplit struct {
	Label  string
	Amount string
	Pct    string
	Share  float64
}

//nolint:gochecknoglobals // static inventory fixture
var sampleCats = []catSplit{
	{"Coffee & tea", "$842k", "34.9%", 0.349},
	{"Packaging", "$518k", "21.5%", 0.215},
	{"Dairy alt.", "$331k", "13.7%", 0.137},
	{"Syrups", "$289k", "12.0%", 0.12},
	{"Snacks", "$216k", "9.0%", 0.09},
	{"Janitorial", "$214k", "8.9%", 0.089},
}

type binStop struct {
	Bg    string
	Fg    string
	Label string
	Range string
}

//nolint:gochecknoglobals // static bin heat scale
var binStops = []binStop{
	{Bg: "#b91c1c", Fg: "#ffffff", Label: "Overfull", Range: "95–100"},
	{Bg: "#ea580c", Fg: "#ffffff", Label: "Full", Range: "85–94"},
	{Bg: "#f59e0b", Fg: "#1c2b24", Label: "High", Range: "70–84"},
	{Bg: "#a9d3bb", Fg: "#1c2b24", Label: "Healthy", Range: "55–69"},
	{Bg: "#1d5c3f", Fg: "#ffffff", Label: "Room", Range: "0–54"},
}

func binColors(v int) (bg, fg string) {
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	switch {
	case v >= 95:
		return binStops[0].Bg, binStops[0].Fg
	case v >= 85:
		return binStops[1].Bg, binStops[1].Fg
	case v >= 70:
		return binStops[2].Bg, binStops[2].Fg
	case v >= 55:
		return binStops[3].Bg, binStops[3].Fg
	default:
		return binStops[4].Bg, binStops[4].Fg
	}
}

func binStatus(v int) string {
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	switch {
	case v >= 95:
		return binStops[0].Label
	case v >= 85:
		return binStops[1].Label
	case v >= 70:
		return binStops[2].Label
	case v >= 55:
		return binStops[3].Label
	default:
		return binStops[4].Label
	}
}

func binCellTitle(aisle int, level string, value int) string {
	return fmt.Sprintf("Aisle %c · %s · %d%% full · %s", 'A'+aisle, level, value, binStatus(value))
}

func aisleAverages() [10]int {
	var avgs [10]int
	if len(sampleBins) == 0 {
		return avgs
	}
	for col := 0; col < 10; col++ {
		sum := 0
		for _, row := range sampleBins {
			sum += row.Cells[col]
		}
		avgs[col] = sum / len(sampleBins)
	}
	return avgs
}

func hottestAisle() (string, int) {
	avgs := aisleAverages()
	best, bestAvg := 0, avgs[0]
	for i := 1; i < len(avgs); i++ {
		if avgs[i] > bestAvg {
			best, bestAvg = i, avgs[i]
		}
	}
	return fmt.Sprintf("%c", 'A'+best), bestAvg
}

func coldestAisle() (string, int) {
	avgs := aisleAverages()
	best, bestAvg := 0, avgs[0]
	for i := 1; i < len(avgs); i++ {
		if avgs[i] < bestAvg {
			best, bestAvg = i, avgs[i]
		}
	}
	return fmt.Sprintf("%c", 'A'+best), bestAvg
}
