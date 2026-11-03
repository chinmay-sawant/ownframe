package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	overviewRegionRatio = 3.4
)

// overviewPanel is the landing tab: SLO burn, headline KPIs, traffic, service
// health, rollouts, on-call coverage, and edge regions.
func (a *app) overviewPanel() *ui.Element {
	return ui.Div().Class("app-stack").Children(
		a.overviewSLOAlert(),
		a.overviewKPIs(),
		a.overviewTrafficCard(),
		ui.Div().Class("app-grid-2").Children(a.overviewServicesCard(), a.overviewRecentEventsCard()),
		ui.Div().Class("app-grid-2").
			GridTemplateColumns(ui.RawValue("minmax(0, 1fr) minmax(0, 2fr)")).
			Children(a.overviewOnCallCard(), a.overviewRegionsCard()),
	)
}

func (a *app) overviewSLOAlert() *ui.Element {
	metrics := a.telemetryMetrics()
	variant := chadcn.AlertWarning
	if metrics.Errors > 4 {
		variant = chadcn.AlertDanger
	}
	return chadcn.Alert(chadcn.AlertOptions{
		Variant:     variant,
		Icon:        ui.Text("▲"),
		Title:       fmt.Sprintf("%d errors in the generated event feed", metrics.Errors),
		Description: fmt.Sprintf("%d errors and %d warnings across %d retained events. This is generated showcase telemetry.", metrics.Errors, metrics.Warnings, metrics.Total),
		Action: chadcn.Button("Open incident room", chadcn.ButtonOptions{
			Variant: chadcn.ButtonDanger,
			Size:    chadcn.ButtonSmall,
		}).OnClick(func(ctx context.Context) error {
			return a.setTab(ctx, "incident")
		}),
	})
}

func (a *app) overviewKPIs() *ui.Element {
	const progressScale = 100
	metrics := a.telemetryMetrics()
	errorRate := fmt.Sprintf("%.1f%%", metrics.ErrorRate)
	availability := fmt.Sprintf("%.1f%%", 100-metrics.ErrorRate)
	latency := metrics.P95Latency.String()
	if metrics.P95Latency == 0 {
		latency = "—"
	}
	kpis := []struct {
		Label  string
		Value  string
		Delta  string
		Tone   chadcn.BadgeVariant
		Detail *ui.Element
	}{
		{"Events captured", fmt.Sprintf("%d", metrics.Total), overviewLatestDelta(a.lastRefreshCount), chadcn.BadgeSecondary, overviewProgress("Retained generated events", min(metrics.Total, progressScale), fmt.Sprintf("%d retained", metrics.Total), "max 500")},
		{"Error rate", errorRate, fmt.Sprintf("%d errors", metrics.Errors), chadcn.BadgeDanger, overviewProgress("Errors in generated events", int(metrics.ErrorRate), fmt.Sprintf("%d warnings", metrics.Warnings), "sample feed")},
		{"p95 latency", latency, "sampled", chadcn.BadgeOutline, overviewLatencyDetail()},
		{"Non-error share", availability, "sampled", chadcn.BadgeSecondary, overviewProgress("Events without errors", int(100-metrics.ErrorRate), "from current feed", "sample feed")},
	}
	cards := make([]*ui.Element, 0, len(kpis))
	for _, kpi := range kpis {
		value := overviewKPIValue(kpi.Value)
		badge := chadcn.Badge(kpi.Delta, chadcn.BadgeOptions{Variant: kpi.Tone})
		detail := kpi.Detail
		cards = append(cards, overviewKPICard(kpi.Label, value, badge, detail))
	}
	return ui.Div().Class("app-kpis").Children(cards...)
}

func overviewLatestDelta(count int) string {
	if count == 0 {
		return "seed data"
	}
	return fmt.Sprintf("+%d latest", count)
}

func overviewKPICard(label string, value, badge, detail *ui.Element) *ui.Element {
	return chadcn.Card(
		ui.Div().Class("app-col").Gap(ui.Px(4)).Children(
			ui.Div().Class("app-kpi-label").Text(label),
			ui.Div().Class("app-row").Children(value, badge),
		),
		detail,
	).Padding(ui.Px(14)).Gap(ui.Px(10))
}

func overviewKPIValue(text string) *ui.Element { return ui.Div().Class("app-kpi-value").Text(text) }

func overviewProgress(label string, value int, left, right string) *ui.Element {
	return ui.Div().Class("app-col").Children(
		chadcn.Progress(label, value),
		ui.Div().Class("app-kpi-foot").Children(ui.Text(left).Class("app-muted"), ui.Text(right).Class("app-muted")),
	)
}

func overviewLatencyDetail() *ui.Element {
	return ui.Div().Class("app-kpi-foot").Children(
		ui.Div().Class("app-row").Children(
			chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK", Size: chadcn.AvatarSmall, Status: chadcn.AvatarStatusOnline}),
			ui.Text("Ada · on call").FontSize(ui.Px(11)),
		),
		chadcn.Kbd("⌘T"),
	)
}

func (a *app) overviewTrafficCard() *ui.Element {
	labels, current, previous := a.eventSeries()
	return overviewCard(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Div().Class("app-col").Gap(ui.Px(3)).Children(
				chadcn.CardTitle("Recent events by service"),
				chadcn.CardDescription("Newest 10 generated events compared with the previous 10."),
			),
			ui.Div().Class("app-row").Children(
				chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Chart actions"},
					chadcn.Button("Refresh", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
						OnClick(func(ctx context.Context) error {
							return a.refreshTelemetry(ctx)
						}),
					chadcn.Button("Export", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
						OnClick(func(ctx context.Context) error {
							return a.showToast(ctx, "Export queued", "Recent generated event counts will land in your inbox.")
						}),
				),
			),
		),
		chadcn.CardContent(chadcn.Chart(chadcn.ChartOptions{
			Labels: labels,
			Series: []chadcn.ChartSeries{
				{Label: "Newest 10", Values: current},
				{Label: "Previous 10", Values: previous},
			},
		}).
			CustomProperty("--of-chart-background", "transparent").
			CustomProperty("--of-chart-header-padding", "0 0 8px").
			CustomProperty("--of-chart-body-padding", "12px 0 0").
			CustomProperty("--of-chart-plot-height", "168px")),
	)
}

func (a *app) overviewServicesCard() *ui.Element {
	services := a.liveServices()
	rows := make([]*ui.Element, 0, len(services))
	for _, service := range services {
		rows = append(rows, chadcn.TableRow(
			chadcn.TableCell(service.Name),
			chadcn.TableCell("", overviewStatusBadge(service.Status)),
			chadcn.TableCell(service.Requests),
			chadcn.TableCell(service.Errors),
			chadcn.TableCell(service.Latency),
			chadcn.TableCell(service.Owner),
		))
	}
	return overviewCard(
		overviewCardHeader("Service health", "Counts and latency derived from the generated event feed.",
			ui.Div().Class("app-row").Children(
				chadcn.Badge(fmt.Sprintf("%d services", len(services)), chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
				chadcn.Button("View logs", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
					OnClick(func(ctx context.Context) error { return a.setTab(ctx, "logs") }),
			)),
		chadcn.CardContent(ui.Div().Class("app-table-wrap").Children(
			chadcn.Table(
				chadcn.TableHeader(chadcn.TableRow(
					chadcn.TableHead("Service"), chadcn.TableHead("Status"), chadcn.TableHead("Requests"),
					chadcn.TableHead("Errors"), chadcn.TableHead("p95"), chadcn.TableHead("Owner"),
				)),
				chadcn.TableBody(rows...),
			),
		)),
	)
}

func overviewStatusBadge(status string) *ui.Element {
	switch status {
	case "degraded", "rolled back":
		return chadcn.Badge(status, chadcn.BadgeOptions{Variant: chadcn.BadgeDanger})
	case "warning":
		return chadcn.Badge(status, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})
	default:
		return chadcn.Badge(status, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})
	}
}

func (a *app) overviewRecentEventsCard() *ui.Element {
	body := ui.Div().Class("app-col").Gap(ui.Px(8))
	entries := []logEntry(nil)
	if a.feed != nil {
		entries = a.feed.logs
	}
	for index, entry := range entries[:min(4, len(entries))] {
		body.Children(chadcn.Item(chadcn.ItemOptions{
			Media:       chadcn.Badge(strings.ToUpper(entry.Level), chadcn.BadgeOptions{Variant: logLevelVariant(entry.Level)}),
			Title:       entry.Service + " · " + statusLabel(entry.Status),
			Description: entry.Time + " · " + entry.Message,
			Actions: chadcn.Button("Open", chadcn.ButtonOptions{Size: chadcn.ButtonSmall, Variant: chadcn.ButtonText}).
				OnClick(func(ctx context.Context) error {
					a.service = "All services"
					a.severity = nil
					a.query = ""
					a.pageIndex = index / logPageSize
					return a.setTab(ctx, "logs")
				}),
		}).Padding(ui.Edges(ui.Px(10), ui.Px(12))))
	}
	return overviewCard(
		overviewCardHeader("Recent events", "Latest generated log entries across services.",
			chadcn.Badge("DEMO", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})),
		chadcn.CardContent(body),
	)
}

func (a *app) overviewOnCallCard() *ui.Element {
	return overviewCard(
		overviewCardHeader("On-call rotation", "Primary responder by day. Pick a date to swap coverage.",
			chadcn.Badge("Week 45", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})),
		chadcn.CardContent(ui.Div().Class("app-col").Children(
			ui.Div().Class("app-row app-row--between").Children(
				ui.Div().Class("app-row").Children(
					chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK", Size: chadcn.AvatarSmall, Status: chadcn.AvatarStatusOnline}),
					ui.Div().Class("app-col").Children(
						ui.Text("Ada Kim").FontWeight(ui.FontWeightBold).FontSize(ui.Px(12)),
						ui.Text("Primary · Europe").Class("app-muted"),
					),
				),
				chadcn.Kbd("⌘⇧C"),
			),
			chadcn.Calendar(chadcn.CalendarOptions{
				Month:    a.calendarMonth,
				Selected: a.fromDate,
				OnSelect: func(ctx context.Context, day time.Time) error {
					a.fromDate = day
					return a.showToast(ctx, "Coverage updated", "Primary responder moved to "+day.Format("Jan 2")+".")
				},
				OnMonthChange: func(ctx context.Context, month time.Time) error {
					a.calendarMonth = month
					return a.refresh(ctx)
				},
			}).CustomProperty("--of-calendar-width", "100%"),
		)),
	)
}

func (a *app) overviewRegionsCard() *ui.Element {
	values := overviewRegionErrors()
	if a.feed != nil {
		values = a.feed.regionErrors()
	}
	return overviewCard(
		overviewCardHeader("Edge regions", "Generated sample error rate by point of presence.",
			chadcn.Badge("14 PoPs", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})),
		chadcn.CardContent(ui.Div().Class("app-col").Gap(ui.Px(12)).Children(
			chadcn.AspectRatio(chadcn.AspectRatioOptions{Ratio: overviewRegionRatio},
				ui.Div().ID("overview-region-map").Class("app-canvas").Children(overviewMesh(values))),
			ui.Div().Class("app-col").Gap(ui.Px(6)).Children(
				ui.Div().Class("app-row app-row--between").Children(
					ui.Div().Class("app-section-label").Text("Errors per PoP · last 60 min"),
					ui.Text("us-east-1 hottest").Class("app-muted"),
				),
				ui.Div().Class("app-heat").Children(overviewHeatBars(values)...),
			),
		)),
	)
}

func overviewRegionErrors() []float64 {
	return []float64{12, 18, 26, 22, 34, 58, 44, 62, 71, 55, 48, 66, 92, 84}
}

func overviewErrorColor(value float64) string {
	if value <= 25 {
		return "#34d399"
	}
	if value < 70 {
		return "#fbbf24"
	}
	return "#f87171"
}

func overviewMesh(values []float64) *ui.Element {
	dots := make([]*ui.Element, 0, len(values))
	for index, value := range values {
		label := overviewPointName(index) + " · " + strconv.FormatFloat(value, 'f', 0, 64) + " errors"
		dots = append(dots, ui.Div().ID("region-map-"+strconv.Itoa(index+1)).Class("app-region-cell").
			Attr("aria-label", label).BackgroundColor(ui.Hex(overviewErrorColor(value))).
			Children(ui.Span().Class("app-heat-tooltip").Text(label)))
	}
	legend := ui.Div().Class("app-region-legend")
	for _, item := range []struct{ label, color string }{{"Low ≤25", "#34d399"}, {"Elevated 26–69", "#fbbf24"}, {"High ≥70", "#f87171"}} {
		legend.Children(ui.Div().Class("app-region-key").Children(
			ui.Span().Class("app-region-swatch").BackgroundColor(ui.Hex(item.color)), ui.Text(item.label)))
	}
	return ui.Div().Class("app-col").AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).Children(
		ui.Div().Class("app-region-cells").Children(dots...),
		ui.Text("6 regions · 14 points of presence · 99.96% edge availability").Class("app-muted"),
		legend,
	)
}

func overviewHeatBars(values []float64) []*ui.Element {
	bars := make([]*ui.Element, 0, len(values))
	for index, value := range values {
		label := overviewPointName(index)
		bars = append(bars, ui.Div().ID("region-heat-"+strconv.Itoa(index+1)).
			Class("app-heat-bar").
			BackgroundColor(ui.Hex(overviewErrorColor(value))).
			Attr("aria-label", label+" · "+strconv.FormatFloat(value, 'f', 0, 64)+" errors").
			Height(ui.Percent(value)).
			Children(ui.Span().ID("region-heat-"+strconv.Itoa(index+1)+"-tooltip").
				Class("app-heat-tooltip").
				Text(label+" · "+strconv.FormatFloat(value, 'f', 0, 64)+" errors")))
	}
	return bars
}

func overviewPointName(index int) string {
	points := [...]string{
		"us-west-1", "us-west-2", "us-east-2", "ca-central-1", "sa-east-1",
		"eu-west-1", "eu-west-2", "eu-central-1", "eu-north-1", "me-south-1",
		"ap-south-1", "ap-southeast-1", "us-east-1", "ap-northeast-1",
	}
	if index < 0 || index >= len(points) {
		return "PoP " + strconv.Itoa(index+1)
	}
	return points[index]
}

func overviewCard(children ...*ui.Element) *ui.Element {
	return chadcn.Card(children...).Padding(ui.Px(16)).Gap(ui.Px(10))
}

func overviewCardHeader(title, description string, trailing *ui.Element) *ui.Element {
	return chadcn.CardHeader(
		ui.Div().Class("app-row app-row--between").Children(chadcn.CardTitle(title), trailing),
		chadcn.CardDescription(description),
	)
}
