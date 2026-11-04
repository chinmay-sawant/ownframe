package main

type navItem struct {
	icon   string
	label  string
	badge  string
	active bool
	action string
}

type navSection struct {
	title string
	items []navItem
}

func navSections() []navSection {
	return []navSection{
		{
			title: "MAIN MENU",
			items: []navItem{
				{icon: "▦", label: "Overview", action: "nav-overview"},
				{icon: "●", label: "Audience", action: "nav-audience"},
				{icon: "✈", label: "Campaigns", action: "nav-campaigns"},
				{icon: "✎", label: "Content", action: "nav-content"},
				{icon: "◉", label: "Engagement", active: true, action: "nav-engagement"},
				{icon: "◆", label: "Revenue", action: "nav-revenue"},
			},
		},
		{
			title: "MANAGE",
			items: []navItem{
				{icon: "○", label: "Subscribers", badge: "18.4k", action: "nav-subscribers"},
				{icon: "↻", label: "Automations", action: "nav-automations"},
				{icon: "◍", label: "Segments", action: "nav-segments"},
				{icon: "◇", label: "Integrations", action: "nav-integrations"},
			},
		},
		{
			title: "RESOURCES",
			items: []navItem{
				{icon: "▤", label: "Analytics", action: "nav-analytics"},
				{icon: "▧", label: "Reports", action: "nav-reports"},
				{icon: "✦", label: "Learn", action: "nav-learn"},
			},
		},
		{
			title: "SETTINGS",
			items: []navItem{
				{icon: "?", label: "Help Center", action: "nav-help"},
				{icon: "⚙", label: "System Settings", action: "nav-settings"},
			},
		},
	}
}

type kpi struct {
	label    string
	value    string
	delta    string
	up       bool
	footnote string
	tint     string
}

func kpis() []kpi {
	return []kpi{
		{label: "Open rate", value: "47.8%", delta: "+3.1%", up: true, footnote: "Unique opens ÷ delivered", tint: "orange"},
		{label: "Click rate", value: "18.6%", delta: "+1.9%", up: true, footnote: "Unique clicks ÷ delivered", tint: "blue"},
		{label: "Click-to-open", value: "38.9%", delta: "+0.8%", up: true, footnote: "Clicks ÷ opens", tint: "purple"},
		{label: "Reply rate", value: "2.4%", delta: "-0.2%", up: false, footnote: "426 replies this month", tint: "green"},
	}
}

type funnelRow struct {
	label  string
	count  string
	pct    string
	filled int
}

func funnelRows() []funnelRow {
	return []funnelRow{
		{label: "Delivered", count: "17,764", pct: "100.0%", filled: 20},
		{label: "Opened", count: "8,491", pct: "47.8%", filled: 10},
		{label: "Clicked", count: "3,304", pct: "18.6%", filled: 4},
		{label: "Replied", count: "426", pct: "2.4%", filled: 1},
		{label: "Upgraded to paid", count: "38", pct: "0.2%", filled: 1},
	}
}

type topLink struct {
	url        string
	context    string
	clicks     int
	share      string
	clicksText string
}

func topLinks() []topLink {
	return []topLink{
		{url: "notion.so/creator-templates", context: "Weekly Creator Brief #52", clicks: 1204, share: "6.8%", clicksText: "1,204"},
		{url: "mailpulse.co/referrals", context: "AI Tools Worth Trying", clicks: 842, share: "4.8%", clicksText: "842"},
		{url: "youtube.com/@creatorbrief", context: "The Creator Economy Report", clicks: 511, share: "3.0%", clicksText: "511"},
	}
}

type engageRow struct {
	label    string
	openPct  string
	clickPct string
	filled   int
}

func engageRows() []engageRow {
	return []engageRow{
		{label: "Paid deep dives", openPct: "68%", clickPct: "31% clicks", filled: 7},
		{label: "Tool roundups", openPct: "52%", clickPct: "21% clicks", filled: 5},
		{label: "Interviews", openPct: "49%", clickPct: "17% clicks", filled: 5},
		{label: "Essays", openPct: "44%", clickPct: "12% clicks", filled: 4},
		{label: "Sponsor posts", openPct: "38%", clickPct: "9% clicks", filled: 4},
	}
}

func heatDays() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
}

func heatHours() []string {
	return []string{"6a", "7a", "8a", "9a", "10a", "11a", "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p"}
}

func heatValue(day, col int) int {
	if day < 0 || day > 6 || col < 0 || col > 17 {
		return 0
	}
	// Pinned peaks: Tue and Thu at 9a.
	if (day == 1 || day == 3) && col == 3 {
		return 4
	}
	// Pinned highs: Mon and Wed at 9-10a.
	if (day == 0 || day == 2) && (col == 3 || col == 4) {
		return 3
	}
	// Pinned evening peak: Sun at 6-7p.
	if day == 6 && (col == 12 || col == 13) {
		return 3
	}
	// Saturday stays quiet.
	if day == 5 {
		if col == 3 || col == 4 {
			return 1
		}
		return 0
	}
	// Sunday mornings are soft, afternoons build to the evening peak.
	if day == 6 {
		switch {
		case col == 3 || col == 4:
			return 2
		case col >= 5 && col <= 11:
			return 1 + (col+day)%2
		case col == 14 || col == 15:
			return 1
		default:
			return 0
		}
	}
	// Weekday shape: strong 8-11a mornings, steady midday, fading evenings.
	base := 0
	switch {
	case col == 2 || col == 5:
		base = 2
	case col == 4:
		base = 2 + (day % 2)
	case col >= 6 && col <= 10:
		base = 1 + ((day + col) % 2)
	case col == 11 || col == 12:
		base = 1
	case col == 0 || col == 1 || col == 13:
		base = (day + col) % 2
	default:
		base = 0
	}
	if base > 3 {
		base = 3
	}
	return base
}
