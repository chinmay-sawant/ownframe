//nolint:mnd // Sample audience fixtures are literal display content: channel shares and range windows are fixed demo values.
package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// audienceDotTotal sizes the 10-dot share bars on the growth card.
const audienceDotTotal = 10

type audienceSource struct {
	label  string
	count  string
	share  string
	filled int
}

func (a *app) audienceSources() []audienceSource {
	switch a.rangeDays() {
	case 7:
		return []audienceSource{
			{label: "Referrals", count: "95", share: "33.5%", filled: 3},
			{label: "Organic", count: "79", share: "27.8%", filled: 3},
			{label: "Import", count: "60", share: "21.1%", filled: 2},
			{label: "API", count: "31", share: "10.9%", filled: 1},
			{label: "Other", count: "19", share: "6.7%", filled: 1},
		}
	case 90:
		return []audienceSource{
			{label: "Referrals", count: "1,149", share: "33.6%", filled: 3},
			{label: "Organic", count: "951", share: "27.8%", filled: 3},
			{label: "Import", count: "725", share: "21.2%", filled: 2},
			{label: "API", count: "373", share: "10.9%", filled: 1},
			{label: "Other", count: "222", share: "6.5%", filled: 1},
		}
	default:
		return []audienceSource{
			{label: "Referrals", count: "405", share: "33.6%", filled: 3},
			{label: "Organic", count: "335", share: "27.8%", filled: 3},
			{label: "Import", count: "255", share: "21.2%", filled: 2},
			{label: "API", count: "131", share: "10.9%", filled: 1},
			{label: "Other", count: "78", share: "6.5%", filled: 1},
		}
	}
}

type audienceLocation struct {
	city  string
	count string
}

func audienceLocations() []audienceLocation {
	return []audienceLocation{
		{city: "New York", count: "2,410"},
		{city: "London", count: "1,880"},
		{city: "Berlin", count: "1,204"},
		{city: "Toronto", count: "986"},
		{city: "Sydney", count: "842"},
	}
}

type subscriber struct {
	name   string
	email  string
	plan   string
	open   string
	status string
}

func subscriberRows() []subscriber {
	return []subscriber{
		{name: "Ada Kim", email: "ada@creatorbrief.co", plan: "Paid", open: "68%", status: "Active"},
		{name: "Bruno Alves", email: "bruno@mailpulse.co", plan: "Free", open: "52%", status: "Active"},
		{name: "Cara Nguyen", email: "cara@roundup.io", plan: "Paid", open: "61%", status: "Active"},
		{name: "Dev Patel", email: "dev@indiehacker.dev", plan: "Free", open: "44%", status: "Trial"},
		{name: "Emma Rossi", email: "emma@newsletter.co", plan: "Paid", open: "71%", status: "Active"},
		{name: "Farah Ahmed", email: "farah@tools.fm", plan: "Free", open: "38%", status: "Dormant"},
		{name: "Gabe Liu", email: "gabe@creatoreco.com", plan: "Free", open: "57%", status: "Bounced"},
		{name: "Hana Sato", email: "hana@tokyodev.jp", plan: "Paid", open: "49%", status: "Unsubscribed"},
	}
}

func subscriberStatusVariant(status string) chadcn.BadgeVariant {
	switch status {
	case "Bounced", "Unsubscribed":
		return chadcn.BadgeDanger
	case "Trial", "Dormant":
		return chadcn.BadgeOutline
	default:
		return chadcn.BadgeSecondary
	}
}

// audiencePage renders the Audience tab: list size, engagement, recent
// growth, and churn, plus growth-by-source and top-location cards.
func (a *app) audiencePage() *ui.Element {
	rangeBtn := chadcn.Button(a.rangeLabel+" ▾", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "range:cycle")
	exportBtn := chadcn.Button("↓ Export", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	head := ui.Div().Class("mp-pagehead").Children(
		ui.Div().Class("mp-pagehead-text").Children(
			chadcn.H1("Audience"),
			chadcn.Muted("Who is on the list and how they joined."),
		),
		ui.Div().Class("mp-pagehead-actions").Children(rangeBtn, exportBtn),
	)
	return ui.Div().Class("mp-card mp-stack").Children(
		head,
		a.audienceKPIs(),
		ui.Div().Class("mp-grid-bot").Children(
			a.audienceGrowthCard(),
			a.audienceLocationsCard(),
		),
	)
}

func (a *app) audienceKPIs() *ui.Element {
	switch a.rangeDays() {
	case 7:
		return ui.Div().Class("mp-kpis").Children(
			audienceKPI("Total subscribers", "18,442", "Across every channel", ""),
			audienceKPI("Engaged", "4,210", "Opened in the last 7 days", "23%"),
			audienceKPI(a.rangeNewLabel(), "284", "Joined in the last 7 days", ""),
			audienceKPI("Unsubscribed", "22", "0.1% of the list", ""),
		)
	case 90:
		return ui.Div().Class("mp-kpis").Children(
			audienceKPI("Total subscribers", "18,442", "Across every channel", ""),
			audienceKPI("Engaged", "15,480", "Opened in the last 90 days", "84%"),
			audienceKPI(a.rangeNewLabel(), "3,420", "Joined in the last 90 days", ""),
			audienceKPI("Unsubscribed", "310", "1.7% of the list", ""),
		)
	default:
		return ui.Div().Class("mp-kpis").Children(
			audienceKPI("Total subscribers", "18,442", "Across every channel", ""),
			audienceKPI("Engaged", "12,908", "Opened in the last 30 days", "70%"),
			audienceKPI(a.rangeNewLabel(), "1,204", "Joined in the last 30 days", ""),
			audienceKPI("Unsubscribed", "96", "0.5% of the list", ""),
		)
	}
}

func audienceKPI(label, value, foot, badge string) *ui.Element {
	middle := []*ui.Element{ui.Tag("strong").Class("mp-kpi-value").Text(value)}
	if badge != "" {
		middle = append(middle, chadcn.Badge(badge, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}))
	}
	return ui.Div().Class("mp-kpi").Children(
		ui.Div().Class("mp-kpi-top").Children(
			ui.Span().Class("mp-kpi-label").Text(label),
		),
		ui.Div().Class("mp-kpi-value-row").Children(middle...),
		ui.Div().Class("mp-kpi-foot").Children(
			ui.Span().Class("mp-muted").Text(foot),
		),
	)
}

func (a *app) audienceGrowthCard() *ui.Element {
	sources := a.audienceSources()
	rows := make([]*ui.Element, 0, len(sources))
	for _, source := range sources {
		rows = append(rows, chadcn.TableRow(
			chadcn.TableCell(source.label),
			chadcn.TableCell(source.count),
			chadcn.TableCell(source.share),
			chadcn.TableCell("", audienceDots(source.filled)),
		).Class("of-table-row "+selectableRow(a.isSelected("src", source.label))).Attr("data-action", "select:src:"+source.label+":"+source.label))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Growth by source"),
			chadcn.CardDescription("New subscribers by channel, "+a.rangeWindow()+"."),
		),
		chadcn.CardContent(
			chadcn.Table(
				chadcn.TableHeader(chadcn.TableRow(
					chadcn.TableHead("Source"),
					chadcn.TableHead("Subscribers"),
					chadcn.TableHead("Share"),
					chadcn.TableHead("Trend"),
				)),
				chadcn.TableBody(rows...),
			),
		),
	)
}

func (a *app) audienceLocationsCard() *ui.Element {
	locations := audienceLocations()
	rows := make([]*ui.Element, 0, len(locations))
	for _, location := range locations {
		rows = append(rows, chadcn.TableRow(
			chadcn.TableCell(location.city),
			chadcn.TableCell(location.count),
		).Class("of-table-row "+selectableRow(a.isSelected("loc", location.city))).Attr("data-action", "select:loc:"+location.city+":"+location.city))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Top locations"),
			chadcn.CardDescription("Subscribers by metro, all time."),
		),
		chadcn.CardContent(
			chadcn.Table(
				chadcn.TableHeader(chadcn.TableRow(
					chadcn.TableHead("Location"),
					chadcn.TableHead("Subscribers"),
				)),
				chadcn.TableBody(rows...),
			),
		),
	)
}

// audienceDots renders a 10-dot share bar with the first filledCount lit.
func audienceDots(filledCount int) *ui.Element {
	dots := make([]*ui.Element, 0, audienceDotTotal)
	for i := range audienceDotTotal {
		class := "mp-dot"
		if i < filledCount {
			class += " mp-dot--on"
		}
		dots = append(dots, ui.Span().ID("mp-src-dot-"+strconv.Itoa(i)).Class(class))
	}
	return ui.Div().Class("mp-dots").Children(dots...)
}

// subscribersPage renders the Subscribers tab: a search field and the
// profile table. Row clicks report select:sub:... actions; selection state
// itself is owned by core. Typing in the search filters the table through
// the host's subQuery state, mirrored from the page form on every change.
func (a *app) subscribersPage() *ui.Element {
	exportBtn := chadcn.Button("↓ Export", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	profiles := filterSubscribers(a.subQuery)
	rows := make([]*ui.Element, 0, len(profiles))
	for _, profile := range profiles {
		id := profile.email
		class := "of-table-row " + selectableRow(a.isSelected("sub", id))
		rows = append(rows, chadcn.TableRow(
			chadcn.TableCell(profile.name),
			chadcn.TableCell(profile.email),
			chadcn.TableCell(profile.plan),
			chadcn.TableCell(profile.open),
			chadcn.TableCell("", chadcn.Badge(profile.status, chadcn.BadgeOptions{Variant: subscriberStatusVariant(profile.status)})),
		).Class(class).Attr("data-action", "select:sub:"+profile.email+":"+profile.name))
	}
	if len(rows) == 0 {
		rows = append(rows, chadcn.TableRow(
			chadcn.TableCell("No subscribers match \""+a.subQuery+"\"."),
		).Class("of-table-row"))
	}
	countLine := "Showing " + strconv.Itoa(len(profiles)) + " of 18,442 profiles."
	head := ui.Div().Class("mp-pagehead").Children(
		ui.Div().Class("mp-pagehead-text").Children(
			chadcn.H1("Subscribers"),
			chadcn.Muted("18.4k profiles. Search filters the table."),
		),
		ui.Div().Class("mp-pagehead-actions").Children(exportBtn),
	)
	return ui.Div().Class("mp-card mp-stack").Children(
		head,
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("All subscribers"),
				chadcn.CardDescription(countLine),
			),
			chadcn.CardContent(
				ui.Div().Class("mp-search").Children(
					chadcn.InputWithOptions("sub-search", a.subQuery, chadcn.InputOptions{
						Placeholder: "Search subscribers",
						Type:        "search",
					}),
				),
				chadcn.Table(
					chadcn.TableHeader(chadcn.TableRow(
						chadcn.TableHead("Name"),
						chadcn.TableHead("Email"),
						chadcn.TableHead("Plan"),
						chadcn.TableHead("Open"),
						chadcn.TableHead("Status"),
					)),
					chadcn.TableBody(rows...),
				),
			),
		),
	)
}

// filterSubscribers returns the profiles matching q across name, email, and
// plan. An empty query returns every profile.
func filterSubscribers(q string) []subscriber {
	all := subscriberRows()
	needle := strings.ToLower(strings.TrimSpace(q))
	if needle == "" {
		return all
	}
	out := make([]subscriber, 0, len(all))
	for _, profile := range all {
		haystack := strings.ToLower(profile.name + " " + profile.email + " " + profile.plan + " " + profile.status)
		if strings.Contains(haystack, needle) {
			out = append(out, profile)
		}
	}
	return out
}
