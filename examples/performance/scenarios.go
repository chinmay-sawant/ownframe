//nolint:mnd // The benchmark fixtures are curated display content: viewport sizes, deterministic multipliers, and percentages are literal fixture values.
package main

import (
	"fmt"
	"strconv"
)

// AllScenarios returns the full suite of 10 performance scenarios.
func AllScenarios() []Scenario {
	return []Scenario{
		makeScenario1SimpleRows50(),
		makeScenario2ZebraTable200(),
		makeScenario3ProductGrid200(),
		makeScenario4SocialTimeline200(),
		makeScenario5OrderBook500(),
		makeScenario6EnterpriseGrid1000(),
		makeScenario7AnalyticsDashboard(),
		makeScenario8MultiPageDoc200(),
		makeScenario9CSSKitchenSink150(),
		makeScenario10ExtremeApp1000(),
	}
}

// -----------------------------------------------------------------------------
// Scenario 1: Simple Row List (50 Rows) - Level: Low
// -----------------------------------------------------------------------------

type SimpleRow struct {
	ID        int
	Title     string
	Category  string
	Status    string
	Timestamp string
}

type SimpleRowsData struct {
	Title string
	Total int
	Rows  []SimpleRow
}

func makeScenario1SimpleRows50() Scenario {
	const n = 50
	rows := make([]SimpleRow, n)
	statuses := []string{"Active", "Pending", "Resolved", "Archived"}
	categories := []string{"Core", "Network", "Auth", "Storage"}

	for i := 0; i < n; i++ {
		rows[i] = SimpleRow{
			ID:        i + 1,
			Title:     fmt.Sprintf("System process worker #%d", i+1),
			Category:  categories[i%len(categories)],
			Status:    statuses[i%len(statuses)],
			Timestamp: fmt.Sprintf("10:%02d:%02d UTC", (i/60)%24, i%60),
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #f8fafc; font-family: sans-serif; }
.header { background-color: #0f172a; color: #ffffff; padding: 14px 20px; font-size: 16px; font-weight: bold; }
.container { padding: 16px; }
.row {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  margin-bottom: 6px;
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  border-radius: 4px;
}
.idx { font-weight: bold; color: #64748b; width: 40px; }
.title { color: #1e293b; font-size: 14px; font-weight: 500; width: 260px; }
.cat { color: #475569; font-size: 12px; background-color: #f1f5f9; padding: 3px 8px; border-radius: 4px; }
.status { font-size: 12px; font-weight: bold; color: #0284c7; width: 80px; text-align: center; }
.time { color: #94a3b8; font-size: 12px; width: 100px; text-align: right; }
</style>
</head>
<body>
<div class="header">{{.Title}} (Total: {{.Total}})</div>
<div class="container">
{{range .Rows}}
  <div class="row">
    <span class="idx">#{{.ID}}</span>
    <span class="title">{{.Title}}</span>
    <span class="cat">{{.Category}}</span>
    <span class="status">{{.Status}}</span>
    <span class="time">{{.Timestamp}}</span>
  </div>
{{end}}
</div>
</body>
</html>`

	return Scenario{
		ID:          "01-simple-rows-50",
		Name:        "Simple Rows (50)",
		Level:       "Low",
		Description: "50-row flat list with basic flex layout, borders, and status chips",
		ItemCount:   n,
		CSSProperties: []string{
			"display: flex", "flex-direction", "justify-content", "align-items", "gap",
			"background-color", "color", "border-bottom", "border-radius", "padding",
			"margin-bottom", "font-family", "font-size", "font-weight", "text-align",
		},
		HTML:   html,
		Data:   SimpleRowsData{Title: "Task Queue Status", Total: n, Rows: rows},
		Width:  1024,
		Height: 768,
		Mutate: func(d any) any {
			v, ok := d.(SimpleRowsData)
			if !ok {
				return d
			}
			v.Rows[0].Status = "Modified"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 2: Standard Data Table (200 Rows) - Level: Medium-Low
// -----------------------------------------------------------------------------

type TableRow struct {
	ID        int
	Customer  string
	Category  string
	Status    string
	StatusCls string
	Amount    string
}

type TableData struct {
	Title string
	Total int
	Rows  []TableRow
}

func makeScenario2ZebraTable200() Scenario {
	const n = 200
	rows := make([]TableRow, n)
	statuses := []string{"Paid", "Pending", "Refunded", "Failed"}
	statusClasses := []string{"status-paid", "status-pending", "status-refunded", "status-failed"}
	categories := []string{"Software", "Hardware", "Consulting", "Cloud Host"}

	for i := 0; i < n; i++ {
		sIdx := i % len(statuses)
		rows[i] = TableRow{
			ID:        1000 + i,
			Customer:  fmt.Sprintf("Acme Client Corporation %d", (i*7)%99+1),
			Category:  categories[i%len(categories)],
			Status:    statuses[sIdx],
			StatusCls: statusClasses[sIdx],
			Amount:    fmt.Sprintf("$%d.%02d", (i*37+12)%5000+50, (i*13)%100),
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #f1f5f9; font-family: sans-serif; color: #0f172a; }
.banner { background-color: #1e293b; color: #ffffff; padding: 16px 24px; font-size: 18px; font-weight: bold; }
.wrap { padding: 20px; }
table { width: 100%; border-collapse: collapse; border-spacing: 0; background-color: #ffffff; border-radius: 6px; }
th { background-color: #e2e8f0; color: #334155; font-size: 13px; font-weight: bold; text-align: left; padding: 10px 14px; border-bottom: 2px solid #cbd5e1; }
td { font-size: 13px; padding: 10px 14px; border-bottom: 1px solid #e2e8f0; }
.row-even { background-color: #f8fafc; }
.row-odd { background-color: #ffffff; }
.col-id { font-family: monospace; font-weight: bold; color: #64748b; width: 80px; }
.col-cust { font-weight: 500; }
.col-amount { text-align: right; font-family: monospace; font-weight: bold; color: #0f172a; width: 110px; }
.pill { display: inline-block; padding: 3px 8px; border-radius: 12px; font-size: 11px; font-weight: bold; text-align: center; }
.status-paid { background-color: #dcfce7; color: #15803d; }
.status-pending { background-color: #fef9c3; color: #a16207; }
.status-refunded { background-color: #e0e7ff; color: #4338ca; }
.status-failed { background-color: #fee2e2; color: #b91c1c; }
</style>
</head>
<body>
<div class="banner">{{.Title}} - {{.Total}} Invoices</div>
<div class="wrap">
  <table>
    <thead>
      <tr>
        <th>Invoice ID</th>
        <th>Customer</th>
        <th>Category</th>
        <th>Status</th>
        <th style="text-align: right;">Amount</th>
      </tr>
    </thead>
    <tbody>
      {{range $i, $r := .Rows}}
      <tr class="{{if eq (printf "%d" $i) "0"}}row-even{{else}}row-odd{{end}}">
        <td class="col-id">#{{$r.ID}}</td>
        <td class="col-cust">{{$r.Customer}}</td>
        <td>{{$r.Category}}</td>
        <td><span class="pill {{$r.StatusCls}}">{{$r.Status}}</span></td>
        <td class="col-amount">{{$r.Amount}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
</div>
</body>
</html>`

	return Scenario{
		ID:          "02-zebra-table-200",
		Name:        "Zebra Table (200 Rows)",
		Level:       "Medium-Low",
		Description: "200-row tabular dataset with thead, tbody, alternating zebra stripes, and pill badges",
		ItemCount:   n,
		CSSProperties: []string{
			"display: table", "border-collapse: collapse", "border-spacing: 0", "padding",
			"border-bottom", "background-color", "color", "font-family", "font-size",
			"font-weight", "text-align", "border-radius", "display: inline-block",
		},
		HTML:   html,
		Data:   TableData{Title: "Enterprise Billing Registry", Total: n, Rows: rows},
		Width:  1280,
		Height: 800,
		Mutate: func(d any) any {
			v, ok := d.(TableData)
			if !ok {
				return d
			}
			v.Rows[0].Amount = "$9,999.99"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 3: Product Catalog Grid (200 Cards) - Level: Medium
// -----------------------------------------------------------------------------

type ProductCard struct {
	ID        int
	Title     string
	Category  string
	Price     string
	OldPrice  string
	Discount  string
	Rating    string
	GradClass string
}

type ProductGridData struct {
	Title string
	Total int
	Cards []ProductCard
}

func makeScenario3ProductGrid200() Scenario {
	const n = 200
	cards := make([]ProductCard, n)
	categories := []string{"Audio", "Compute", "Wearable", "Display", "Accessories"}
	gradClasses := []string{"g-blue", "g-purple", "g-emerald", "g-amber", "g-rose"}

	for i := 0; i < n; i++ {
		cards[i] = ProductCard{
			ID:        i + 1,
			Title:     fmt.Sprintf("Ultra Wireless Device Pro %d", i+1),
			Category:  categories[i%len(categories)],
			Price:     fmt.Sprintf("$%d", (i*23)%800+49),
			OldPrice:  fmt.Sprintf("$%d", (i*23)%800+119),
			Discount:  fmt.Sprintf("-%d%%", (i*7)%40+10),
			Rating:    "★★★★☆ (4.8)",
			GradClass: gradClasses[i%len(gradClasses)],
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #0f172a; font-family: sans-serif; color: #f8fafc; }
.topbar { background-color: #1e293b; padding: 16px 24px; border-bottom: 1px solid #334155; font-size: 18px; font-weight: bold; }
.grid-container {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 24px;
}
.card {
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.thumb {
  height: 90px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: bold;
  color: #ffffff;
}
.g-blue { background: linear-gradient(135deg, #1d4ed8, #3b82f6); }
.g-purple { background: linear-gradient(135deg, #6d28d9, #a855f7); }
.g-emerald { background: linear-gradient(135deg, #047857, #10b981); }
.g-amber { background: linear-gradient(135deg, #b45309, #f59e0b); }
.g-rose { background: linear-gradient(135deg, #be123c, #f43f5e); }
.card-body { padding: 12px; display: flex; flex-direction: column; gap: 6px; }
.tag-row { display: flex; justify-content: space-between; align-items: center; }
.cat-tag { font-size: 11px; background-color: #334155; color: #94a3b8; padding: 2px 6px; border-radius: 4px; }
.disc-tag { font-size: 11px; background-color: #f43f5e; color: #ffffff; padding: 2px 6px; border-radius: 4px; font-weight: bold; }
.prod-title { font-size: 13px; font-weight: bold; color: #f1f5f9; margin-top: 4px; }
.rating { font-size: 11px; color: #fbbf24; }
.price-row { display: flex; justify-content: space-between; align-items: center; margin-top: 6px; }
.price { font-size: 16px; font-weight: bold; color: #38bdf8; }
.old-price { font-size: 12px; color: #64748b; }
.btn {
  background-color: #0284c7;
  color: #ffffff;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: bold;
  text-align: center;
  margin-top: 8px;
}
</style>
</head>
<body>
<div class="topbar">{{.Title}} ({{.Total}} Products)</div>
<div class="grid-container">
{{range .Cards}}
  <div class="card">
    <div class="thumb {{.GradClass}}">{{.Category}} Mockup</div>
    <div class="card-body">
      <div class="tag-row">
        <span class="cat-tag">{{.Category}}</span>
        <span class="disc-tag">{{.Discount}}</span>
      </div>
      <div class="prod-title">{{.Title}}</div>
      <div class="rating">{{.Rating}}</div>
      <div class="price-row">
        <span class="price">{{.Price}}</span>
        <span class="old-price">{{.OldPrice}}</span>
      </div>
      <div class="btn">Add to Cart</div>
    </div>
  </div>
{{end}}
</div>
</body>
</html>`

	return Scenario{
		ID:          "03-product-grid-200",
		Name:        "Product Catalog Grid (200 Cards)",
		Level:       "Medium",
		Description: "200-card e-commerce catalog in a 4-column CSS Grid with gradients and action buttons",
		ItemCount:   n,
		CSSProperties: []string{
			"display: grid", "grid-template-columns: repeat(4, 1fr)", "gap: 16px",
			"display: flex", "flex-direction: column", "linear-gradient", "border-radius: 8px",
			"border", "overflow: hidden", "align-items: center", "justify-content",
			"background-color", "color", "font-weight", "font-size", "padding",
		},
		HTML:   html,
		Data:   ProductGridData{Title: "Storefront Inventory", Total: n, Cards: cards},
		Width:  1280,
		Height: 900,
		Mutate: func(d any) any {
			v, ok := d.(ProductGridData)
			if !ok {
				return d
			}
			v.Cards[0].Price = "$999"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 4: Social / Chat Feed (200 Posts) - Level: Medium-High
// -----------------------------------------------------------------------------

type SocialPost struct {
	ID       int
	Author   string
	Handle   string
	Initials string
	Color    string
	TimeAgo  string
	Content  string
	Tag      string
	Likes    int
	Comments int
	Reposts  int
}

type SocialFeedData struct {
	Title string
	Total int
	Posts []SocialPost
}

func makeScenario4SocialTimeline200() Scenario {
	const n = 200
	posts := make([]SocialPost, n)
	authors := []string{"Sarah Jenkins", "Alex Rivera", "David Chen", "Elena Rostova", "Marcus Vance"}
	handles := []string{"@sjenkins", "@arivera_dev", "@dchen_ai", "@elena_ux", "@mvance_sec"}
	initials := []string{"SJ", "AR", "DC", "ER", "MV"}
	avatarColors := []string{"#3b82f6", "#8b5cf6", "#ec4899", "#10b981", "#f59e0b"}
	tags := []string{"#engineering", "#architecture", "#performance", "#security", "#design"}

	for i := 0; i < n; i++ {
		idx := i % len(authors)
		posts[i] = SocialPost{
			ID:       i + 1,
			Author:   authors[idx],
			Handle:   handles[idx],
			Initials: initials[idx],
			Color:    avatarColors[idx],
			TimeAgo:  fmt.Sprintf("%dm ago", (i*3)%59+1),
			Content:  fmt.Sprintf("Dispatched real-world benchmark iteration %d across distributed worker clusters. Latency metrics stay optimal with sub-millisecond dispatch cycles.", i+1),
			Tag:      tags[i%len(tags)],
			Likes:    (i * 19) % 340,
			Comments: (i * 7) % 85,
			Reposts:  (i * 3) % 42,
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #f8fafc; font-family: sans-serif; color: #0f172a; }
.header-bar { background-color: #ffffff; border-bottom: 1px solid #e2e8f0; padding: 14px 20px; font-size: 16px; font-weight: bold; }
.feed-stream { max-width: 720px; margin: 0 auto; padding: 20px 0; }
.post-card {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  display: flex;
  flex-direction: row;
  gap: 12px;
}
.avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  color: #ffffff;
  font-weight: bold;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.post-main { display: flex; flex-direction: column; gap: 8px; flex-grow: 1; }
.meta-row { display: flex; flex-direction: row; align-items: center; gap: 8px; }
.author-name { font-weight: bold; font-size: 14px; color: #0f172a; }
.author-handle { font-size: 12px; color: #64748b; }
.time-ago { font-size: 12px; color: #94a3b8; }
.post-content { font-size: 13px; line-height: 1.5; color: #334155; }
.tag-badge { display: inline-block; font-size: 11px; color: #0284c7; font-weight: 500; }
.action-bar { display: flex; flex-direction: row; gap: 24px; margin-top: 6px; font-size: 12px; color: #64748b; font-weight: 500; }
.action-item { display: flex; align-items: center; gap: 4px; }
</style>
</head>
<body>
<div class="header-bar">{{.Title}} - {{.Total}} Messages</div>
<div class="feed-stream">
{{range .Posts}}
  <div class="post-card">
    <div class="avatar" style="background-color: {{.Color}};">{{.Initials}}</div>
    <div class="post-main">
      <div class="meta-row">
        <span class="author-name">{{.Author}}</span>
        <span class="author-handle">{{.Handle}}</span>
        <span class="time-ago">· {{.TimeAgo}}</span>
      </div>
      <div class="post-content">{{.Content}}</div>
      <div class="tag-badge">{{.Tag}}</div>
      <div class="action-bar">
        <span class="action-item">💬 {{.Comments}}</span>
        <span class="action-item">🔄 {{.Reposts}}</span>
        <span class="action-item">❤️ {{.Likes}}</span>
        <span class="action-item">↗ Share</span>
      </div>
    </div>
  </div>
{{end}}
</div>
</body>
</html>`

	return Scenario{
		ID:          "04-social-timeline-200",
		Name:        "Social Chat Timeline (200 Posts)",
		Level:       "Medium-High",
		Description: "200-post social messaging feed with avatars, nested flexboxes, metadata chips, and action bars",
		ItemCount:   n,
		CSSProperties: []string{
			"display: flex", "flex-direction: row/column", "border-radius: 50%", "border-radius: 8px",
			"align-items: center", "gap", "margin: 0 auto", "margin-bottom: 12px", "padding: 16px",
			"border", "line-height: 1.5", "flex-shrink: 0", "flex-grow: 1", "background-color",
		},
		HTML:   html,
		Data:   SocialFeedData{Title: "Team Collaboration Stream", Total: n, Posts: posts},
		Width:  1024,
		Height: 800,
		Mutate: func(d any) any {
			v, ok := d.(SocialFeedData)
			if !ok {
				return d
			}
			v.Posts[0].Likes++
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 5: Financial Trading / Order Book (500 Rows) - Level: High
// -----------------------------------------------------------------------------

type OrderBookRow struct {
	Side         string // "bid" or "ask"
	Price        string
	Amount       string
	Total        string
	DepthPercent int
}

type OrderBookData struct {
	Pair  string
	Total int
	Rows  []OrderBookRow
}

func makeScenario5OrderBook500() Scenario {
	const n = 500
	rows := make([]OrderBookRow, n)
	basePrice := 67500.0

	for i := 0; i < n; i++ {
		isBid := i < n/2
		side := "bid"
		p := basePrice - float64(i)*0.5
		if !isBid {
			side = "ask"
			p = basePrice + float64(i-n/2)*0.5
		}
		amount := float64((i*17)%45+5) / 10.0
		total := p * amount
		depth := (i%100)*1 + 5

		rows[i] = OrderBookRow{
			Side:         side,
			Price:        fmt.Sprintf("%.2f", p),
			Amount:       fmt.Sprintf("%.4f", amount),
			Total:        fmt.Sprintf("%.2f", total),
			DepthPercent: depth,
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #0b0e14; font-family: monospace; color: #cbd5e1; }
.ob-header { background-color: #111827; padding: 12px 16px; font-size: 14px; font-weight: bold; border-bottom: 1px solid #1f2937; }
.ladder { width: 600px; margin: 16px auto; border: 1px solid #1f2937; border-radius: 4px; overflow: hidden; }
.ladder-cols { display: flex; justify-content: space-between; background-color: #111827; padding: 8px 12px; font-size: 11px; color: #9ca3af; border-bottom: 1px solid #1f2937; }
.row-wrap { position: relative; height: 22px; border-bottom: 1px solid rgba(255,255,255,0.03); }
.depth-bar-bid { position: absolute; top: 0; bottom: 0; right: 0; background-color: rgba(34, 197, 94, 0.15); }
.depth-bar-ask { position: absolute; top: 0; bottom: 0; right: 0; background-color: rgba(239, 68, 68, 0.15); }
.row-content { position: relative; display: flex; justify-content: space-between; align-items: center; padding: 0 12px; height: 22px; font-size: 12px; z-index: 2; }
.price-bid { color: #22c55e; font-weight: bold; width: 120px; }
.price-ask { color: #ef4444; font-weight: bold; width: 120px; }
.cell-amt { color: #e2e8f0; width: 120px; text-align: right; }
.cell-tot { color: #94a3b8; width: 140px; text-align: right; }
</style>
</head>
<body>
<div class="ob-header">Market Depth Ladder: {{.Pair}} ({{.Total}} Depth Rows)</div>
<div class="ladder">
  <div class="ladder-cols">
    <span>PRICE (USD)</span>
    <span style="text-align: right;">SIZE (BTC)</span>
    <span style="text-align: right;">TOTAL (USD)</span>
  </div>
{{range .Rows}}
  <div class="row-wrap">
    {{if eq .Side "bid"}}
      <div class="depth-bar-bid" style="width: {{.DepthPercent}}%;"></div>
    {{else}}
      <div class="depth-bar-ask" style="width: {{.DepthPercent}}%;"></div>
    {{end}}
    <div class="row-content">
      {{if eq .Side "bid"}}
        <span class="price-bid">{{.Price}}</span>
      {{else}}
        <span class="price-ask">{{.Price}}</span>
      {{end}}
      <span class="cell-amt">{{.Amount}}</span>
      <span class="cell-tot">{{.Total}}</span>
    </div>
  </div>
{{end}}
</div>
</body>
</html>`

	return Scenario{
		ID:          "05-order-book-500",
		Name:        "Market Order Book (500 Rows)",
		Level:       "High",
		Description: "500-row financial order book depth ladder (250 bids, 250 asks) with proportional depth bars",
		ItemCount:   n,
		CSSProperties: []string{
			"position: relative", "position: absolute", "top: 0", "bottom: 0", "right: 0",
			"z-index", "display: flex", "justify-content: space-between", "align-items: center",
			"font-family: monospace", "text-align: right", "background-color: rgba(...)",
			"color", "border-bottom", "width: %",
		},
		HTML:   html,
		Data:   OrderBookData{Pair: "BTC-USD Realtime", Total: n, Rows: rows},
		Width:  1024,
		Height: 800,
		Mutate: func(d any) any {
			v, ok := d.(OrderBookData)
			if !ok {
				return d
			}
			v.Rows[0].Amount = "99.9999"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 6: Dense Enterprise Data Grid (1,000 Rows) - Level: High
// -----------------------------------------------------------------------------

type EnterpriseRow struct {
	Index      int
	Key        string
	Title      string
	Severity   string
	SevClass   string
	Department string
	Owner      string
	Date       string
	Status     string
}

type EnterpriseGridData struct {
	Title string
	Total int
	Rows  []EnterpriseRow
}

func makeScenario6EnterpriseGrid1000() Scenario {
	const n = 1000
	rows := make([]EnterpriseRow, n)
	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
	sevClasses := []string{"sev-crit", "sev-high", "sev-med", "sev-low"}
	depts := []string{"Security", "Backend", "Frontend", "Database", "Infra"}
	owners := []string{"c.sawant", "m.smith", "j.doe", "e.taylor", "l.vance"}
	statuses := []string{"Open", "In Review", "Mitigated", "Resolved"}

	for i := 0; i < n; i++ {
		sIdx := i % len(severities)
		rows[i] = EnterpriseRow{
			Index:      i + 1,
			Key:        fmt.Sprintf("INC-%05d", 10000+i),
			Title:      fmt.Sprintf("High-throughput edge node latency spike in cluster us-east-%d", (i%8)+1),
			Severity:   severities[sIdx],
			SevClass:   sevClasses[sIdx],
			Department: depts[i%len(depts)],
			Owner:      owners[i%len(owners)],
			Date:       fmt.Sprintf("2026-10-%02d", (i%30)+1),
			Status:     statuses[i%len(statuses)],
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #0f172a; font-family: sans-serif; color: #e2e8f0; }
.banner { background-color: #1e293b; border-bottom: 2px solid #334155; padding: 14px 20px; font-size: 16px; font-weight: bold; }
.grid-head { display: flex; align-items: center; gap: 8px; background-color: #1e293b; padding: 8px 16px; font-size: 12px; font-weight: bold; color: #94a3b8; border-bottom: 1px solid #334155; }
.row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  background-color: #0f172a;
  border-bottom: 1px solid #1e293b;
  font-size: 12px;
}
.c-chk { width: 24px; text-align: center; }
.c-idx { width: 48px; color: #64748b; font-family: monospace; }
.c-key { width: 88px; font-weight: bold; color: #38bdf8; font-family: monospace; }
.c-title { width: 340px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #f1f5f9; }
.c-sev { width: 80px; text-align: center; }
.c-dept { width: 84px; color: #cbd5e1; }
.c-owner { width: 90px; color: #94a3b8; }
.c-date { width: 90px; color: #64748b; }
.c-act { width: 60px; text-align: right; }
.btn-view { background-color: #334155; color: #e2e8f0; padding: 2px 8px; border-radius: 4px; font-size: 11px; }
.chip { display: inline-block; padding: 2px 6px; border-radius: 4px; font-size: 10px; font-weight: bold; }
.sev-crit { background-color: #7f1d1d; color: #fecaca; }
.sev-high { background-color: #7c2d12; color: #fed7aa; }
.sev-med { background-color: #713f12; color: #fef08a; }
.sev-low { background-color: #14532d; color: #bbf7d0; }
</style>
</head>
<body>
<div class="banner">{{.Title}} (Total Rows: {{.Total}})</div>
<div class="grid-head">
  <span class="c-chk">☑</span>
  <span class="c-idx">#</span>
  <span class="c-key">INCIDENT</span>
  <span class="c-title">SUMMARY</span>
  <span class="c-sev">SEVERITY</span>
  <span class="c-dept">DEPT</span>
  <span class="c-owner">OWNER</span>
  <span class="c-date">DATE</span>
  <span class="c-act">ACTION</span>
</div>
{{range .Rows}}
<div class="row">
  <span class="c-chk">☐</span>
  <span class="c-idx">{{.Index}}</span>
  <span class="c-key">{{.Key}}</span>
  <span class="c-title">{{.Title}}</span>
  <span class="c-sev"><span class="chip {{.SevClass}}">{{.Severity}}</span></span>
  <span class="c-dept">{{.Department}}</span>
  <span class="c-owner">{{.Owner}}</span>
  <span class="c-date">{{.Date}}</span>
  <span class="c-act"><span class="btn-view">View</span></span>
</div>
{{end}}
</body>
</html>`

	return Scenario{
		ID:          "06-enterprise-grid-1000",
		Name:        "Enterprise Data Grid (1,000 Rows)",
		Level:       "High",
		Description: "1,000-row enterprise record grid with checkboxes, severity tags, department chips, and actions",
		ItemCount:   n,
		CSSProperties: []string{
			"display: flex", "align-items: center", "gap: 8px", "padding: 6px 16px",
			"white-space: nowrap", "overflow: hidden", "text-overflow: ellipsis",
			"border-bottom: 1px solid #1e293b", "border-radius: 4px", "font-size: 12px",
			"color", "background-color", "width",
		},
		HTML:   html,
		Data:   EnterpriseGridData{Title: "Enterprise Production Issue Registry", Total: n, Rows: rows},
		Width:  1280,
		Height: 800,
		Mutate: func(d any) any {
			v, ok := d.(EnterpriseGridData)
			if !ok {
				return d
			}
			v.Rows[0].Severity = "RESOLVED"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 7: Heavy CSS Styled Analytics Dashboard - Level: Complex
// -----------------------------------------------------------------------------

type KPICard struct {
	Title     string
	Value     string
	Change    string
	IsUp      bool
	GradClass string
}

type TimelineItem struct {
	Step      int
	Time      string
	Title     string
	Detail    string
	StatusCol string
}

type ResourceMeter struct {
	Label    string
	Percent  int
	Color    string
	UsageStr string
}

type DashboardData struct {
	AppTitle  string
	KPIs      []KPICard
	Timeline  []TimelineItem
	Resources []ResourceMeter
}

func makeScenario7AnalyticsDashboard() Scenario {
	kpis := []KPICard{
		{Title: "Gross Platform Volume", Value: "$4,289,120", Change: "+18.4%", IsUp: true, GradClass: "kpi-indigo"},
		{Title: "Active Service Mesh Pods", Value: "12,450", Change: "+4.2%", IsUp: true, GradClass: "kpi-teal"},
		{Title: "P99 Layout Latency", Value: "1.42 ms", Change: "-24.1%", IsUp: true, GradClass: "kpi-emerald"},
		{Title: "Memory Allocation Frame", Value: "482 KB", Change: "-12.5%", IsUp: true, GradClass: "kpi-violet"},
	}

	const tCount = 20
	timeline := make([]TimelineItem, tCount)
	tTitles := []string{"Distributed Cache Invalidation", "TLS Key Rotation Verified", "Ingress Pipeline Scaled", "Storage Compaction Done"}
	tColors := []string{"#3b82f6", "#10b981", "#f59e0b", "#8b5cf6"}

	for i := 0; i < tCount; i++ {
		timeline[i] = TimelineItem{
			Step:      i + 1,
			Time:      fmt.Sprintf("%02d:%02d UTC", (i*2)%24, (i*13)%60),
			Title:     fmt.Sprintf("%s #%d", tTitles[i%len(tTitles)], i+1),
			Detail:    fmt.Sprintf("Cluster node-pool-zone-%d dispatched heartbeat confirmation.", (i%4)+1),
			StatusCol: tColors[i%len(tColors)],
		}
	}

	resources := []ResourceMeter{
		{Label: "CPU Cores Allocated", Percent: 68, Color: "#3b82f6", UsageStr: "21.7 / 32 Cores"},
		{Label: "Unified Memory Bandwidth", Percent: 54, Color: "#10b981", UsageStr: "69.1 / 128 GB"},
		{Label: "NVMe IOPS Utilization", Percent: 82, Color: "#f59e0b", UsageStr: "82,000 / 100,000 IOPS"},
		{Label: "Network Ingress/Egress", Percent: 43, Color: "#8b5cf6", UsageStr: "4.3 / 10 Gbps"},
		{Label: "Kage Shader Buffer Heap", Percent: 31, Color: "#06b6d4", UsageStr: "78 / 256 MB"},
		{Label: "DisplayList Op Cache", Percent: 88, Color: "#ec4899", UsageStr: "14,200 / 16,000 Ops"},
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #090d16; font-family: sans-serif; color: #f1f5f9; }
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 24px;
  background-color: #111827;
  border-bottom: 1px solid #1f2937;
}
.brand { font-size: 18px; font-weight: bold; color: #38bdf8; }
.profile-pill { background-color: #1f2937; padding: 4px 12px; border-radius: 16px; font-size: 12px; }
.content-wrapper { padding: 20px; display: flex; flex-direction: column; gap: 20px; }
.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.kpi-card {
  border-radius: 10px;
  padding: 16px;
  border: 1px solid rgba(255,255,255,0.1);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.kpi-indigo { background: linear-gradient(135deg, #1e1b4b, #312e81); }
.kpi-teal { background: linear-gradient(135deg, #134e4a, #115e59); }
.kpi-emerald { background: linear-gradient(135deg, #064e3b, #065f46); }
.kpi-violet { background: linear-gradient(135deg, #4c1d95, #5b21b6); }
.kpi-title { font-size: 12px; color: #94a3b8; font-weight: 500; }
.kpi-value { font-size: 22px; font-weight: bold; color: #ffffff; }
.kpi-change { font-size: 12px; font-weight: bold; color: #34d399; }
.main-split { display: grid; grid-template-columns: 2fr 1fr; gap: 20px; }
.panel {
  background-color: #111827;
  border: 1px solid #1f2937;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.panel-head { font-size: 15px; font-weight: bold; color: #e2e8f0; border-bottom: 1px solid #1f2937; padding-bottom: 10px; }
.t-item { display: flex; gap: 12px; align-items: flex-start; }
.t-badge { width: 28px; height: 28px; border-radius: 50%; color: #fff; font-weight: bold; font-size: 11px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.t-body { display: flex; flex-direction: column; gap: 2px; }
.t-title { font-size: 13px; font-weight: bold; color: #f3f4f6; }
.t-desc { font-size: 11px; color: #9ca3af; }
.meter-item { display: flex; flex-direction: column; gap: 4px; margin-bottom: 10px; }
.meter-top { display: flex; justify-content: space-between; font-size: 12px; }
.meter-bar { height: 8px; background-color: #1f2937; border-radius: 4px; overflow: hidden; }
.meter-fill { height: 100%; border-radius: 4px; }
</style>
</head>
<body>
<div class="top-nav">
  <span class="brand">{{.AppTitle}}</span>
  <span class="profile-pill">Cluster: Global-Mesh-Primary</span>
</div>
<div class="content-wrapper">
  <div class="kpi-row">
  {{range .KPIs}}
    <div class="kpi-card {{.GradClass}}">
      <span class="kpi-title">{{.Title}}</span>
      <span class="kpi-value">{{.Value}}</span>
      <span class="kpi-change">{{.Change}}</span>
    </div>
  {{end}}
  </div>
  <div class="main-split">
    <div class="panel">
      <div class="panel-head">Operational Activity Timeline</div>
      {{range .Timeline}}
      <div class="t-item">
        <div class="t-badge" style="background-color: {{.StatusCol}};">{{.Step}}</div>
        <div class="t-body">
          <span class="t-title">{{.Title}} <span style="font-size: 11px; color: #64748b; font-weight: normal;">({{.Time}})</span></span>
          <span class="t-desc">{{.Detail}}</span>
        </div>
      </div>
      {{end}}
    </div>
    <div class="panel">
      <div class="panel-head">Cluster Telemetry</div>
      {{range .Resources}}
      <div class="meter-item">
        <div class="meter-top">
          <span>{{.Label}}</span>
          <span style="color: #94a3b8; font-family: monospace;">{{.UsageStr}}</span>
        </div>
        <div class="meter-bar">
          <div class="meter-fill" style="width: {{.Percent}}%; background-color: {{.Color}};"></div>
        </div>
      </div>
      {{end}}
    </div>
  </div>
</div>
</body>
</html>`

	return Scenario{
		ID:          "07-analytics-dashboard",
		Name:        "Analytics Dashboard (Multi-Pane)",
		Level:       "Complex",
		Description: "Multi-pane executive analytics dashboard with KPI gradient cards, 2-column split, activity timeline, and progress bars",
		ItemCount:   tCount + len(kpis) + len(resources),
		CSSProperties: []string{
			"display: grid", "grid-template-columns: 2fr 1fr", "grid-template-columns: repeat(4, 1fr)",
			"display: flex", "linear-gradient", "border-radius: 10px", "border",
			"overflow: hidden", "align-items: center", "justify-content", "gap",
			"width: %", "background-color", "color", "padding",
		},
		HTML: html,
		Data: DashboardData{
			AppTitle:  "Executive Telemetry Operations Console",
			KPIs:      kpis,
			Timeline:  timeline,
			Resources: resources,
		},
		Width:  1280,
		Height: 900,
		Mutate: func(d any) any {
			v, ok := d.(DashboardData)
			if !ok {
				return d
			}
			v.KPIs[0].Value = "$4,295,000"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 8: Multi-Page Document (200 Pages / Sections) - Level: Complex
// -----------------------------------------------------------------------------

type PageDoc struct {
	PageNumber  int
	TotalPages  int
	Title       string
	Chapter     string
	Paragraph1  string
	Paragraph2  string
	CalloutText string
	ListItems   []string
	Date        string
}

type MultiPageData struct {
	DocTitle string
	Total    int
	Pages    []PageDoc
}

func makeScenario8MultiPageDoc200() Scenario {
	const n = 200
	pages := make([]PageDoc, n)

	for i := 0; i < n; i++ {
		pages[i] = PageDoc{
			PageNumber:  i + 1,
			TotalPages:  n,
			Title:       fmt.Sprintf("Section %d: Verification of Rendering Pipeline Invariants", i+1),
			Chapter:     fmt.Sprintf("Chapter %d.%d", (i/10)+1, (i%10)+1),
			Paragraph1:  "The rendering pipeline computes cascading style rules, element dimensions, font glyph bounding metrics, and transforms prior to drawing list emission. Ensuring predictable computational scaling across document sections is critical for deterministic execution in host environments.",
			Paragraph2:  "All layout passes operate without dynamic memory expansion across repeated frame updates when dirty region tracking isolates mutated subtrees. Software blits and layer composites preserve target frame timing.",
			CalloutText: fmt.Sprintf("Note 14.%d: Compliance verification confirmed under 60 Hz steady-state display timing.", i+1),
			ListItems: []string{
				"Memory bounds verified under strict user limits",
				"Glyph raster cache eviction policy active",
				"Dirty box containment validated without overflow",
			},
			Date: "2026-10-31",
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #cbd5e1; font-family: sans-serif; color: #1e293b; }
.doc-header-banner { background-color: #0f172a; color: #ffffff; padding: 16px 32px; font-size: 18px; font-weight: bold; }
.page-sheet {
  width: 780px;
  margin: 24px auto;
  background-color: #ffffff;
  border: 1px solid #94a3b8;
  border-radius: 4px;
  padding: 36px 44px;
  box-sizing: border-box;
}
.sheet-top {
  display: flex;
  justify-content: space-between;
  border-bottom: 2px solid #0f172a;
  padding-bottom: 8px;
  margin-bottom: 20px;
  font-size: 12px;
  font-weight: bold;
  color: #64748b;
}
.sheet-chapter { font-size: 13px; font-weight: bold; color: #0284c7; margin-bottom: 4px; }
.sheet-title { font-size: 18px; font-weight: bold; color: #0f172a; margin-bottom: 16px; }
.para { font-size: 13px; line-height: 1.6; color: #334155; margin-bottom: 12px; }
.callout {
  background-color: #f0f9ff;
  border-left: 4px solid #0284c7;
  padding: 10px 14px;
  margin: 16px 0;
  font-size: 12px;
  font-weight: 500;
  color: #0369a1;
}
.bullet-list { margin: 12px 0; padding-left: 20px; font-size: 12px; color: #475569; }
.bullet-list li { margin-bottom: 4px; }
.sheet-footer {
  margin-top: 28px;
  border-top: 1px solid #e2e8f0;
  padding-top: 10px;
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #94a3b8;
}
</style>
</head>
<body>
<div class="doc-header-banner">{{.DocTitle}} (Total Document Pages: {{.Total}})</div>
{{range .Pages}}
<div class="page-sheet">
  <div class="sheet-top">
    <span>ENGINEERING SPECIFICATION REPORT</span>
    <span>Page {{.PageNumber}} of {{.TotalPages}}</span>
  </div>
  <div class="sheet-chapter">{{.Chapter}}</div>
  <div class="sheet-title">{{.Title}}</div>
  <div class="para">{{.Paragraph1}}</div>
  <div class="callout">{{.CalloutText}}</div>
  <div class="para">{{.Paragraph2}}</div>
  <ul class="bullet-list">
    {{range .ListItems}}
    <li>{{.}}</li>
    {{end}}
  </ul>
  <div class="sheet-footer">
    <span>CONFIDENTIAL & PROPRIETARY</span>
    <span>Published: {{.Date}}</span>
  </div>
</div>
{{end}}
</body>
</html>`

	return Scenario{
		ID:          "08-multipage-doc-200",
		Name:        "Multi-Page Document (200 Pages)",
		Level:       "Complex",
		Description: "200-page paginated technical report with bordered sheet wrappers, headers, paragraphs, callout boxes, and footers",
		ItemCount:   n,
		CSSProperties: []string{
			"width: 780px", "margin: 24px auto", "padding: 36px 44px", "box-sizing: border-box",
			"border-left: 4px solid #0284c7", "border-bottom: 2px solid #0f172a", "border-top: 1px solid",
			"border-radius: 4px", "line-height: 1.6", "display: flex", "justify-content: space-between",
			"background-color", "color", "font-size", "font-weight",
		},
		HTML:   html,
		Data:   MultiPageData{DocTitle: "OwnFrame Architectural Reference Document", Total: n, Pages: pages},
		Width:  1024,
		Height: 800,
		Mutate: func(d any) any {
			v, ok := d.(MultiPageData)
			if !ok {
				return d
			}
			v.Pages[0].Title = "Modified Verification of Pipeline"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 9: CSS Properties Kitchen Sink (150 Cards) - Level: Complex
// -----------------------------------------------------------------------------

type SinkCard struct {
	ID        int
	Badge     string
	Title     string
	Subtitle  string
	GradClass string
	Radius    string
	Transform string
	LetterSp  string
	WordSp    string
	BStyle    string
	BColor    string
}

type SinkData struct {
	Title string
	Total int
	Cards []SinkCard
}

func makeScenario9CSSKitchenSink150() Scenario {
	const n = 150
	cards := make([]SinkCard, n)
	gradClasses := []string{"sink-g1", "sink-g2", "sink-g3", "sink-g4", "sink-g5"}
	radii := []string{
		"16px 4px 16px 4px",
		"4px 16px 4px 16px",
		"12px 12px 0 0",
		"0 0 12px 12px",
		"20px",
	}
	transforms := []string{
		"scale(0.98)",
		"translate(1px, 1px)",
		"scale(1.0)",
		"translate(0, 0)",
	}
	borderStyles := []string{"solid", "dashed", "dotted"}
	borderColors := []string{"#38bdf8", "#818cf8", "#34d399", "#f472b6", "#fbbf24"}

	for i := 0; i < n; i++ {
		cards[i] = SinkCard{
			ID:        i + 1,
			Badge:     fmt.Sprintf("PROPERTY-SET-%03d", i+1),
			Title:     fmt.Sprintf("CSS Composite Matrix Widget %d", i+1),
			Subtitle:  "Exercising border-radius, gradients, transforms, opacity, letter-spacing, and word-spacing.",
			GradClass: gradClasses[i%len(gradClasses)],
			Radius:    radii[i%len(radii)],
			Transform: transforms[i%len(transforms)],
			LetterSp:  fmt.Sprintf("%dpx", (i%4)+1),
			WordSp:    fmt.Sprintf("%dpx", (i%5)+2),
			BStyle:    borderStyles[i%len(borderStyles)],
			BColor:    borderColors[i%len(borderColors)],
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #030712; font-family: sans-serif; color: #f9fafb; }
.sink-banner { background-color: #111827; padding: 16px 24px; font-size: 18px; font-weight: bold; border-bottom: 1px solid #1f2937; }
.sink-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 24px;
}
.sink-card {
  padding: 18px;
  background-color: #111827;
  display: flex;
  flex-direction: column;
  gap: 10px;
  box-sizing: border-box;
}
.sink-g1 { background: linear-gradient(135deg, #1e1b4b 0%, #312e81 100%); }
.sink-g2 { background: linear-gradient(135deg, #022c22 0%, #064e3b 100%); }
.sink-g3 { background: linear-gradient(135deg, #3b0764 0%, #581c87 100%); }
.sink-g4 { background: linear-gradient(135deg, #450a0a 0%, #7f1d1d 100%); }
.sink-g5 { background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%); }
.badge-tag {
  display: inline-block;
  font-size: 10px;
  font-weight: bold;
  padding: 3px 8px;
  background-color: rgba(255, 255, 255, 0.15);
  border-radius: 4px;
  text-transform: uppercase;
}
.sink-title { font-size: 14px; font-weight: bold; color: #ffffff; }
.sink-sub { font-size: 12px; line-height: 1.5; color: #d1d5db; }
.prop-strip {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #9ca3af;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 8px;
}
</style>
</head>
<body>
<div class="sink-banner">{{.Title}} ({{.Total}} Combinatorial Widgets)</div>
<div class="sink-grid">
{{range .Cards}}
  <div class="sink-card {{.GradClass}}" style="border-radius: {{.Radius}}; border-width: 2px; border-style: {{.BStyle}}; border-color: {{.BColor}}; transform: {{.Transform}};">
    <div><span class="badge-tag">{{.Badge}}</span></div>
    <div class="sink-title" style="letter-spacing: {{.LetterSp}};">{{.Title}}</div>
    <div class="sink-sub" style="word-spacing: {{.WordSp}};">{{.Subtitle}}</div>
    <div class="prop-strip">
      <span>Radius: {{.Radius}}</span>
      <span>Style: {{.BStyle}}</span>
    </div>
  </div>
{{end}}
</div>
</body>
</html>`

	return Scenario{
		ID:          "09-css-kitchen-sink-150",
		Name:        "CSS Kitchen Sink (150 Widgets)",
		Level:       "Complex",
		Description: "150 feature showcase cards combining diverse supported CSS properties (elliptical radii, logical borders, gradients, transforms, typography)",
		ItemCount:   n,
		CSSProperties: []string{
			"border-radius: (varied)", "linear-gradient", "transform: scale/translate",
			"letter-spacing", "word-spacing", "text-transform: uppercase",
			"border-style: solid/dashed/dotted", "border-color", "display: grid",
			"grid-template-columns: repeat(3, 1fr)", "gap: 16px", "padding: 18px",
			"line-height: 1.5", "box-sizing: border-box",
		},
		HTML:   html,
		Data:   SinkData{Title: "CSS Property Permutation Matrix", Total: n, Cards: cards},
		Width:  1280,
		Height: 900,
		Mutate: func(d any) any {
			v, ok := d.(SinkData)
			if !ok {
				return d
			}
			v.Cards[0].Title = "MODIFIED PROPERTY"
			return v
		},
	}
}

// -----------------------------------------------------------------------------
// Scenario 10: Extreme Scale Application (1,000 Complex Rows) - Level: Extreme
// -----------------------------------------------------------------------------

type ExtremeRow struct {
	Rank      int
	Initials  string
	GradClass string
	Name      string
	Email     string
	Role      string
	Status    string
	StatusCls string
	Progress  int
	Amount    string
	Tag1      string
	Tag2      string
}

type ExtremeAppData struct {
	Title string
	Total int
	Rows  []ExtremeRow
}

func makeScenario10ExtremeApp1000() Scenario {
	const n = 1000
	rows := make([]ExtremeRow, n)
	gradClasses := []string{"ext-g1", "ext-g2", "ext-g3", "ext-g4", "ext-g5"}
	statuses := []string{"Active", "Auditing", "Synced", "Standby"}
	statusClasses := []string{"st-act", "st-aud", "st-syn", "st-std"}
	roles := []string{"Principal Architect", "Systems Specialist", "Cluster Operator", "Security Auditor"}
	tags := []string{"Kubernetes", "Blinkless", "EbitenEngine", "PostgreSQL", "KafkaStream"}

	for i := 0; i < n; i++ {
		sIdx := i % len(statuses)
		rows[i] = ExtremeRow{
			Rank:      i + 1,
			Initials:  fmt.Sprintf("%c%c", 'A'+(i%26), 'Z'-(i%26)),
			GradClass: gradClasses[i%len(gradClasses)],
			Name:      fmt.Sprintf("Engineer Operator %d", i+1),
			Email:     fmt.Sprintf("operator.%d@datacenter.internal", i+1),
			Role:      roles[i%len(roles)],
			Status:    statuses[sIdx],
			StatusCls: statusClasses[sIdx],
			Progress:  (i*13)%90 + 10,
			Amount:    fmt.Sprintf("$%s,%03d.00", strconv.Itoa((i*29)%800+12), (i*71)%999),
			Tag1:      tags[i%len(tags)],
			Tag2:      tags[(i+2)%len(tags)],
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; padding: 0; background-color: #0b0f19; font-family: sans-serif; color: #f8fafc; }
.ext-top {
  background-color: #111827;
  border-bottom: 2px solid #1f2937;
  padding: 16px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.ext-brand { font-size: 18px; font-weight: bold; color: #38bdf8; }
.ext-count { font-size: 12px; color: #94a3b8; font-weight: bold; }
.row-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background-color: #0b0f19;
  border-bottom: 1px solid #1e293b;
  font-size: 12px;
}
.c-rank {
  width: 32px;
  height: 32px;
  background-color: #1e293b;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 11px;
  color: #94a3b8;
  flex-shrink: 0;
}
.c-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 12px;
  flex-shrink: 0;
}
.ext-g1 { background: linear-gradient(135deg, #2563eb, #3b82f6); }
.ext-g2 { background: linear-gradient(135deg, #7c3aed, #a855f7); }
.ext-g3 { background: linear-gradient(135deg, #059669, #10b981); }
.ext-g4 { background: linear-gradient(135deg, #d97706, #f59e0b); }
.ext-g5 { background: linear-gradient(135deg, #e11d48, #f43f5e); }
.c-user-info { width: 220px; display: flex; flex-direction: column; gap: 2px; }
.user-name { font-weight: bold; color: #f1f5f9; font-size: 13px; }
.user-email { color: #64748b; font-size: 11px; }
.c-role { width: 140px; color: #cbd5e1; font-weight: 500; }
.c-status { width: 80px; text-align: center; }
.pill-status { display: inline-block; padding: 3px 8px; border-radius: 12px; font-size: 10px; font-weight: bold; }
.st-act { background-color: #064e3b; color: #6ee7b7; }
.st-aud { background-color: #78350f; color: #fde68a; }
.st-syn { background-color: #1e3a8a; color: #93c5fd; }
.st-std { background-color: #374151; color: #d1d5db; }
.c-prog { width: 130px; display: flex; flex-direction: column; gap: 3px; }
.prog-label { font-size: 10px; color: #94a3b8; display: flex; justify-content: space-between; }
.prog-track { height: 6px; background-color: #1e293b; border-radius: 3px; overflow: hidden; }
.prog-fill { height: 100%; background: linear-gradient(90deg, #10b981, #059669); border-radius: 3px; }
.c-amount { width: 110px; text-align: right; font-family: monospace; font-weight: bold; color: #38bdf8; }
.c-tags { display: flex; gap: 4px; width: 160px; }
.tag-chip { font-size: 10px; background-color: #1e293b; color: #94a3b8; padding: 2px 6px; border-radius: 4px; }
.c-action { width: 70px; text-align: right; }
.btn-action { background-color: #0284c7; color: #ffffff; padding: 4px 10px; border-radius: 4px; font-size: 11px; font-weight: bold; }
</style>
</head>
<body>
<div class="ext-top">
  <span class="ext-brand">{{.Title}}</span>
  <span class="ext-count">1,000 Complex Components Loaded</span>
</div>
{{range .Rows}}
<div class="row-item">
  <div class="c-rank">#{{.Rank}}</div>
  <div class="c-avatar {{.GradClass}}">{{.Initials}}</div>
  <div class="c-user-info">
    <span class="user-name">{{.Name}}</span>
    <span class="user-email">{{.Email}}</span>
  </div>
  <div class="c-role">{{.Role}}</div>
  <div class="c-status"><span class="pill-status {{.StatusCls}}">{{.Status}}</span></div>
  <div class="c-prog">
    <div class="prog-label"><span>Usage</span><span>{{.Progress}}%</span></div>
    <div class="prog-track">
      <div class="prog-fill" style="width: {{.Progress}}%;"></div>
    </div>
  </div>
  <div class="c-amount">{{.Amount}}</div>
  <div class="c-tags">
    <span class="tag-chip">{{.Tag1}}</span>
    <span class="tag-chip">{{.Tag2}}</span>
  </div>
  <div class="c-action"><span class="btn-action">Inspect</span></div>
</div>
{{end}}
</body>
</html>`

	return Scenario{
		ID:          "10-extreme-app-1000",
		Name:        "Extreme Scale Application (1,000 Complex Rows)",
		Level:       "Extreme",
		Description: "1,000 fully-featured application rows with rank chips, gradient avatars, multi-line titles, status badges, progress bars, currency metrics, tag lists, and buttons",
		ItemCount:   n,
		CSSProperties: []string{
			"display: flex", "align-items: center", "gap: 12px", "linear-gradient",
			"border-radius: 50%", "border-radius: 6px", "border-radius: 12px",
			"overflow: hidden", "font-family: monospace", "width: %", "background-color",
			"color", "font-size", "font-weight", "text-align: right", "flex-shrink: 0",
		},
		HTML:   html,
		Data:   ExtremeAppData{Title: "Enterprise Global Cluster Operations", Total: n, Rows: rows},
		Width:  1280,
		Height: 900,
		Mutate: func(d any) any {
			v, ok := d.(ExtremeAppData)
			if !ok {
				return d
			}
			v.Rows[0].Amount = "$999,999.00"
			return v
		},
	}
}
