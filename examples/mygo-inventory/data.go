// Data for the MyGo-inventory dashboard: copies of the fixtures in
// examples/builder/inventory, so the two apps stay with respect to each
// other. Products drive the stock table and the product heat map; bins
// drive the bin utilization grid; POs drive the inbound panel; categories
// drive the value-by-category panel.
package main

import "fmt"

// product is one separate inventory product.
type product struct {
	name     string
	short    string
	code     string
	category string
	status   string
	onHand   int
	inbound  int // -1 means none, rendered as an em dash
	daysLeft int
	reorder  float64 // 0..1 fractions for the stock bar and marker
	stock    float64
}

// products mirrors builder/inventory sampleSKUs verbatim.
var products = []product{
	{"Cold brew filter packs, 50 ct", "Cold brew", "NW-CF-2210", "Coffee & tea", "Stockout", 0, 1200, 0, 0.05, 0.05},
	{"Ceremonial matcha, 1 kg", "Matcha", "NW-CF-3302", "Coffee & tea", "Stockout", 0, 200, 0, 0.12, 0.12},
	{"Oat milk barista, 32 oz x 12", "Oat milk", "NW-DA-0418", "Dairy alternatives", "Reorder", 86, 600, 3, 0.35, 0.30},
	{"Compostable cups 12 oz, case 1,000", "Cups", "NW-PK-1120", "Packaging", "Reorder", 412, -1, 9, 0.55, 0.30},
	{"House espresso blend, 5 lb", "Espresso", "NW-CF-1001", "Coffee & tea", "Healthy", 640, 480, 24, 0.60, 0.55},
	{"Nitrile gloves, large, box 100", "Gloves", "NW-JS-5510", "Janitorial", "Healthy", 1120, -1, 33, 0.65, 0.60},
	{"Paper straws 7.75 in, case 5,000", "Straws", "NW-PK-0907", "Packaging", "Healthy", 1940, -1, 41, 0.85, 0.65},
	{"Vanilla syrup, 750 ml", "Vanilla", "NW-SY-0031", "Syrups", "Overstock", 2380, -1, 122, 0.85, 0.85},
}

// bins mirrors builder/inventory sampleBins: 5 rack levels x 10 aisles,
// values are % of cubic capacity.
var bins = [][10]int{
	{42, 55, 61, 38, 70, 66, 48, 35, 52, 28},   // L5
	{68, 74, 80, 62, 88, 85, 71, 59, 73, 47},   // L4
	{81, 86, 92, 77, 95, 91, 83, 72, 84, 63},   // L3
	{88, 93, 97, 85, 99, 96, 90, 81, 89, 70},   // L2
	{94, 98, 100, 91, 100, 99, 95, 87, 93, 76}, // L1
}

var binLevels = []string{"L5", "L4", "L3", "L2", "L1"}

// binStop is one step of the inventory heat scale.
type binStop struct {
	bg    string
	fg    string
	label string
	scale string
}

// binStops mirrors builder/inventory binStops verbatim.
var binStops = []binStop{
	{"#b91c1c", "#ffffff", "Overfull", "95-100"},
	{"#ea580c", "#ffffff", "Full", "85-94"},
	{"#f59e0b", "#1c2b24", "High", "70-84"},
	{"#a9d3bb", "#1c2b24", "Healthy", "55-69"},
	{"#1d5c3f", "#ffffff", "Room", "0-54"},
}

func clampBin(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func binColors(v int) (bg, fg string) {
	switch v = clampBin(v); {
	case v >= 95:
		return binStops[0].bg, binStops[0].fg
	case v >= 85:
		return binStops[1].bg, binStops[1].fg
	case v >= 70:
		return binStops[2].bg, binStops[2].fg
	case v >= 55:
		return binStops[3].bg, binStops[3].fg
	default:
		return binStops[4].bg, binStops[4].fg
	}
}

func binStatus(v int) string {
	switch v = clampBin(v); {
	case v >= 95:
		return binStops[0].label
	case v >= 85:
		return binStops[1].label
	case v >= 70:
		return binStops[2].label
	case v >= 55:
		return binStops[3].label
	default:
		return binStops[4].label
	}
}

// aisleAverages is the mean bin fill per aisle over the 5 levels.
func aisleAverages() [10]int {
	var avgs [10]int
	for col := 0; col < 10; col++ {
		sum := 0
		for _, row := range bins {
			sum += row[col]
		}
		avgs[col] = sum / len(bins)
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

func binSummary() string {
	hotName, hotAvg := hottestAisle()
	coldName, coldAvg := coldestAisle()
	return fmt.Sprintf("Hottest aisle %s · %d%% avg · Coldest aisle %s · %d%% avg", hotName, hotAvg, coldName, coldAvg)
}

func binCellTitle(aisle int, level string, value int) string {
	return fmt.Sprintf("Aisle %c · %s · %d%% full · %s", 'A'+aisle, level, value, binStatus(value))
}

// cellValue scales the aisle's average fill by the product's own stock
// health: 0.5 + StockPct, so a Stockout product reads cold and an
// Overstock product reads hot on the same aisle.
func cellValue(pi, ai int) int {
	avgs := aisleAverages()
	return clampBin(int(float64(avgs[ai])*(0.5+products[pi].stock) + 0.5))
}

func cellLabel(pi, ai int) string {
	return fmt.Sprintf("cell-%d-%d", pi, ai)
}

func cellTip(pi, ai int) string {
	v := cellValue(pi, ai)
	p := products[pi]
	return fmt.Sprintf("%s (%s) · Aisle %c · %d%% full · %s", p.name, p.code, 'A'+ai, v, binStatus(v))
}

// inboundPO mirrors builder/inventory inboundPO verbatim.
type inboundPO struct {
	vendor string
	poNo   string
	detail string
	eta    string
	amount string
	badge  string
}

// purchaseOrders mirrors builder/inventory samplePOs verbatim.
var purchaseOrders = []inboundPO{
	{"Pacific Bean", "PO-10482", "Dock 3 · 1680 units", "Today 2:00 PM", "$48,600", "Arriving today"},
	{"Oat Co. Foodservice", "PO-10477", "truck 62% of route · 600 cases", "Sat Nov 1", "$21,300", "In transit"},
	{"Uji Matcha Trading", "PO-10469", "Port of Oakland · 200 kg", "Thu Nov 6", "$38,920", "On the water"},
	{"GreenWare Packaging", "PO-10491", "rail Chicago · 2400 cases", "Mon Nov 10", "$77,600", "In transit"},
}

// catSplit mirrors builder/inventory catSplit verbatim.
type catSplit struct {
	label  string
	amount string
	pct    string
	share  float64
}

// categories mirrors builder/inventory sampleCats verbatim.
var categories = []catSplit{
	{"Coffee & tea", "$842k", "34.9%", 0.349},
	{"Packaging", "$518k", "21.5%", 0.215},
	{"Dairy alt.", "$331k", "13.7%", 0.137},
	{"Syrups", "$289k", "12.0%", 0.120},
	{"Snacks", "$216k", "9.0%", 0.090},
	{"Janitorial", "$214k", "8.9%", 0.089},
}

// catColors mirrors the green ramp in builder/inventory categories.go.
var catColors = []string{"#1d5c3f", "#3d8b5f", "#7ab894", "#a9d3bb", "#cfe7d8", "#e3efe7"}
