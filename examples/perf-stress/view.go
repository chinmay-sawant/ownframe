package main

import "strconv"

// Row is one data-grid row: five cells, so 280 rows add 1400 nodes.
// Seq is the global index; the template uses it for stable row ids.
type Row struct {
	Seq    int
	Num    string
	Title  string
	Body   string
	Status string
	Delta  string
}

// Card is one dashboard stat card.
type Card struct {
	Label string
	Value string
	Hint  string
}

// View carries the whole dashboard: nav, cards, grid, and tick state.
// Rows is the visible window; TopPad and BotPad hold the laid-out height
// of the rows above and below it, so the scrollbar stays honest.
type View struct {
	Title    string
	Tick     int
	Progress int
	Active   string
	Nav      []string
	Cards    []Card
	Rows     []Row
	TopPad   int
	BotPad   int
}

var navItems = []string{"Overview", "Orders", "Customers", "Inventory", "Reports", "Alerts", "Settings", "Profile"}

var statuses = []string{"✅ paid", "⏳ pending", "⚠️ review", "❌ failed"}

// makeView builds a dashboard with n grid rows.
func makeView(n int) View {
	v := View{Title: "Stress", Active: "Overview", Nav: navItems}
	v.Cards = []Card{
		{"Revenue", "$184.2k", "▲ 12% 🚀"},
		{"Orders", "8,412", "▲ 4% 📦"},
		{"Latency p99", "212ms", "▼ 9% ⚡"},
		{"Errors", "37", "▼ 21% ✅"},
		{"Queue", "1,024", "▲ 2% ⏳"},
		{"Uptime", "99.98%", "🔥 stable"},
	}
	v.Rows = make([]Row, n)
	for i := range v.Rows {
		v.Rows[i] = Row{
			Seq:    i,
			Num:    "#" + strconv.Itoa(i+1),
			Title:  "Order " + strconv.Itoa(orderBase+i),
			Body:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit 🎉",
			Status: statuses[i%len(statuses)],
			Delta:  "+" + strconv.Itoa((i*deltaFactor)%deltaMod),
		}
	}
	return v
}
