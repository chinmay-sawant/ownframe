package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	overviewKPISkeletonWidth     = 92
	overviewKPISkeletonHeight    = 26
	overviewLineHeight           = 6
	overviewChartSkeletonHeight  = 54
	overviewRegionRatio          = 3.4
	overviewHeatSpan             = 7
	overviewHeatStride           = 3
	overviewHeatBaseOpacity      = 0.12
	overviewHeatStep             = 0.1
	overviewFooterSkeletonHeight = 10
	overviewFooterSkeletonWidth  = 120
)

// overviewPanel is the landing tab: SLO burn, headline KPIs, traffic, service
// health, rollouts, on-call coverage, and edge regions.
func (a *app) overviewPanel() *ui.Element {
	return ui.Div().Class("app-stack").Children(
		a.overviewSLOAlert(),
		a.overviewKPIs(),
		a.overviewTrafficCard(),
		ui.Div().Class("app-grid-2").Children(a.overviewServicesCard(), a.overviewDeploymentsCard()),
		ui.Div().Class("app-grid-2").
			GridTemplateColumns(ui.RawValue("minmax(0, 360px) minmax(0, 1fr)")).
			Children(a.overviewOnCallCard(), a.overviewRegionsCard()),
		a.overviewToolkitCard(),
	)
}

func (a *app) overviewSLOAlert() *ui.Element {
	return chadcn.Alert(chadcn.AlertOptions{
		Variant:     chadcn.AlertWarning,
		Icon:        ui.Text("▲"),
		Title:       "Checkout SLO is burning 4.2× the hourly budget",
		Description: "0.42% of checkout requests failed in the last hour against a 0.10% target. 38% of the error budget is left.",
		Action: chadcn.Button("Open incident room", chadcn.ButtonOptions{
			Variant: chadcn.ButtonDanger,
			Size:    chadcn.ButtonSmall,
		}).OnClick(func(ctx context.Context) error {
			return a.setTab(ctx, "incident")
		}),
	})
}

func (a *app) overviewKPIs() *ui.Element {
	const (
		peakCapacity  = 78
		errorBudget   = 62
		uptimePercent = 99
	)
	kpis := []struct {
		Label  string
		Value  string
		Delta  string
		Tone   chadcn.BadgeVariant
		Detail *ui.Element
	}{
		{"Request volume", "1.84M", "▲ 12.4%", chadcn.BadgeSecondary, overviewProgress("Peak throughput against capacity", peakCapacity, "Peak 14:20", "2.4k req/min")},
		{"Error rate", "0.42%", "▲ 0.19 pts", chadcn.BadgeDanger, overviewProgress("Error budget used", errorBudget, "Budget 38% left", "5xx + 429")},
		{"p95 latency", "310 ms", "−18 ms", chadcn.BadgeOutline, overviewLatencyDetail()},
		{"Uptime", "99.94%", "30d", chadcn.BadgeSecondary, overviewProgress("Uptime against the 99.95% target", uptimePercent, "−0.01 pts", "SLO 99.95%")},
	}
	cards := make([]*ui.Element, 0, len(kpis))
	for index, kpi := range kpis {
		value := overviewKPIValue(kpi.Value)
		badge := chadcn.Badge(kpi.Delta, chadcn.BadgeOptions{Variant: kpi.Tone})
		detail := kpi.Detail
		// The error KPI doubles as the skeleton and spinner preview.
		if index == 1 && a.loading {
			value = chadcn.Skeleton(chadcn.SkeletonOptions{Width: overviewKPISkeletonWidth, Height: overviewKPISkeletonHeight})
			badge = chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerSmall, Label: "Refreshing error rate"})
			detail = ui.Div().Class("app-col").Children(
				chadcn.Skeleton(chadcn.SkeletonOptions{Height: overviewLineHeight, Rounded: true}),
				ui.Text("Refreshing the last 15 minutes…").Class("app-muted"),
			)
		}
		cards = append(cards, overviewKPICard(kpi.Label, value, badge, detail))
	}
	return ui.Div().Class("app-kpis").Children(cards...)
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
	labels, current, previous := overviewTrafficSeries(a.rangeMode)
	return overviewCard(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Div().Class("app-col").Gap(ui.Px(3)).Children(
				chadcn.CardTitle("Request volume"),
				chadcn.CardDescription("Requests across all services, this period against the previous one."),
			),
			ui.Div().Class("app-row").Children(
				chadcn.Label("Range", chadcn.LabelOptions{For: "overview-range-24h"}),
				chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
					Size:     chadcn.ButtonSmall,
					Selected: []string{a.rangeMode},
					Items: []chadcn.ToggleItem{
						{ID: "overview-range-24h", Value: "24h", Label: "24h"},
						{ID: "overview-range-7d", Value: "7d", Label: "7d"},
						{ID: "overview-range-30d", Value: "30d", Label: "30d"},
					},
					OnSelect: func(ctx context.Context, value string) error {
						a.rangeMode = value
						return a.refresh(ctx)
					},
				}),
				chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Chart actions"},
					chadcn.Button("Refresh", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
						OnClick(func(ctx context.Context) error {
							a.loading = true
							return a.showToast(ctx, "Refreshing metrics", "Pulling the last 15 minutes across 6 services.")
						}),
					chadcn.Button("Export", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
						OnClick(func(ctx context.Context) error {
							return a.showToast(ctx, "Export queued", "Request volume for "+a.rangeMode+" will land in your inbox.")
						}),
				),
			),
		),
		chadcn.CardContent(chadcn.Chart(chadcn.ChartOptions{
			Labels: labels,
			Series: []chadcn.ChartSeries{
				{Label: "This period", Values: current},
				{Label: "Previous period", Values: previous},
			},
		}).
			CustomProperty("--of-chart-background", "transparent").
			CustomProperty("--of-chart-header-padding", "0 0 8px").
			CustomProperty("--of-chart-body-padding", "12px 0 0").
			CustomProperty("--of-chart-plot-height", "168px")),
	)
}

func overviewTrafficSeries(mode string) ([]string, []float64, []float64) {
	switch mode {
	case "7d":
		return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
			[]float64{1820000, 2140000, 1980000, 2460000, 2210000, 1640000, 1490000},
			[]float64{1710000, 2010000, 2050000, 2280000, 2080000, 1580000, 1410000}
	case "30d":
		return []string{"W1", "W2", "W3", "W4", "W5"},
			[]float64{12800000, 14100000, 13550000, 15200000, 14620000},
			[]float64{11950000, 13300000, 12900000, 14450000, 14010000}
	default:
		return []string{"00:00", "04:00", "08:00", "12:00", "16:00", "20:00"},
			[]float64{228000, 176000, 312000, 448000, 392000, 284000},
			[]float64{204000, 158000, 278000, 398000, 348000, 251000}
	}
}

func (a *app) overviewServicesCard() *ui.Element {
	rows := make([]*ui.Element, 0, len(sampleServices))
	for _, service := range sampleServices {
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
		overviewCardHeader("Service health", "Rolling 15-minute window · error rate and p95 per service.",
			ui.Div().Class("app-row").Children(
				chadcn.Badge("6 services", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
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

func (a *app) overviewDeploymentsCard() *ui.Element {
	type rollout struct{ Service, Version, Actor, When, Status string }
	rollouts := []rollout{
		{"api-gateway", "v4.18.2", "AK", "12m ago", "deployed"},
		{"web", "v2.9.0", "LT", "48m ago", "deployed"},
		{"worker", "v1.14.6", "RS", "3h ago", "rolled back"},
		{"billing", "v3.2.1", "AK", "yesterday", "deployed"},
	}
	body := ui.Div().Class("app-col").Gap(ui.Px(8))
	if a.loading {
		for range rollouts {
			body.Children(chadcn.Skeleton(chadcn.SkeletonOptions{Height: overviewChartSkeletonHeight, Rounded: true}))
		}
	} else {
		for _, item := range rollouts {
			body.Children(chadcn.Item(chadcn.ItemOptions{
				Media:       chadcn.Avatar(chadcn.AvatarOptions{Initials: item.Actor, Size: chadcn.AvatarSmall}),
				Title:       item.Service + " " + item.Version,
				Description: item.When + " · main",
				Actions:     overviewStatusBadge(item.Status),
			}).Padding(ui.Edges(ui.Px(10), ui.Px(12))))
		}
		body.Children(chadcn.Empty(chadcn.EmptyOptions{
			Title:       "End of the rollout feed",
			Description: "Older releases are kept in the audit archive.",
			Actions: []*ui.Element{chadcn.Button("Browse archive", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
				OnClick(func(ctx context.Context) error {
					return a.showToast(ctx, "Archive opened", "Showing releases from the last 90 days.")
				})},
		}).CustomProperty("--of-empty-padding", "16px 8px 4px"))
	}
	return overviewCard(
		overviewCardHeader("Recent deployments", "Merges to main in the last 14 days.",
			chadcn.Badge("main", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})),
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
	return overviewCard(
		overviewCardHeader("Edge regions", "Request share and error rate by point of presence.",
			chadcn.Badge("14 PoPs", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})),
		chadcn.CardContent(ui.Div().Class("app-col").Gap(ui.Px(12)).Children(
			chadcn.AspectRatio(chadcn.AspectRatioOptions{Ratio: overviewRegionRatio},
				ui.Div().Class("app-canvas").Children(overviewMesh())),
			ui.Div().Class("app-col").Gap(ui.Px(6)).Children(
				ui.Div().Class("app-row app-row--between").Children(
					ui.Div().Class("app-section-label").Text("Errors per PoP · last 60 min"),
					ui.Text("us-east-1 hottest").Class("app-muted"),
				),
				ui.Div().Class("app-heat").Children(overviewHeatBars(
					[]float64{12, 18, 26, 22, 34, 58, 44, 62, 71, 55, 48, 66, 92, 84, 73, 59, 64, 52, 47, 38, 31, 27, 21, 16},
				)...),
			),
		)),
	)
}

func overviewMesh() *ui.Element {
	const meshDots = 60
	dots := make([]*ui.Element, 0, meshDots)
	for index := range meshDots {
		dots = append(dots, ui.Span().Width(ui.Px(9)).Height(ui.Px(9)).BorderRadius(ui.Px(2)).
			BackgroundColor(ui.ColorVar("--of-primary", ui.Hex("#fafafa"))).
			Opacity(ui.Number(overviewHeatBaseOpacity+float64((index*overviewHeatStride)%overviewHeatSpan)/10*overviewHeatStep)))
	}
	return ui.Div().Class("app-col").AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).Children(
		ui.Div().Display(ui.DisplayGrid).GridTemplateColumns(ui.RawValue("repeat(12, 9px)")).Gap(ui.Px(3)).Children(dots...),
		ui.Text("6 regions · 14 points of presence · 99.96% edge availability").Class("app-muted"),
	)
}

func overviewHeatBars(values []float64) []*ui.Element {
	bars := make([]*ui.Element, 0, len(values))
	for _, value := range values {
		bars = append(bars, ui.Span().Height(ui.Percent(value)).Opacity(ui.Number(0.25+value/200)))
	}
	return bars
}

func (a *app) overviewToolkitCard() *ui.Element {
	return overviewCard(
		ui.Div().Class("app-row app-row--between").Children(
			chadcn.H1("Responder toolkit").FontSize(ui.Px(20)),
			chadcn.Badge("Design system", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
		),
		chadcn.Lead("From first page to postmortem."),
		chadcn.P("Every card reads live telemetry. Drill into a service, open its trace, then leave a note for the next responder."),
		chadcn.Muted("Reviewed by Ada Kim · 2 days ago"),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		ui.Div().Class("app-row app-row--between").Children(
			ui.Div().Class("app-row").Children(
				chadcn.Kbd("⌘K"),
				chadcn.Kbd("G O"),
				chadcn.Kbd("⇧L"),
				ui.Text("keyboard shortcuts").Class("app-muted"),
			),
			ui.Div().Class("app-row").Children(
				chadcn.Label("Skeletons", chadcn.LabelOptions{For: "overview-loading-toggle"}),
				chadcn.Toggle("Preview", a.loading, chadcn.ToggleOptions{Size: chadcn.ButtonSmall}).
					ID("overview-loading-toggle").
					OnClick(func(ctx context.Context) error {
						a.loading = !a.loading
						return a.refresh(ctx)
					}),
			),
		),
		chadcn.H2("Loading states").FontSize(ui.Px(12)),
		overviewLoadingPreview(a.loading),
	)
}

func overviewLoadingPreview(loading bool) *ui.Element {
	if !loading {
		return chadcn.Muted("Idle. Turn on Preview to render skeleton and spinner states.")
	}
	return ui.Div().Class("app-col").Children(
		ui.Div().Class("app-row").Children(
			chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerSmall, Label: "Loading panel data"}),
			ui.Text("Fetching the last 15 minutes…").Class("app-muted"),
		),
		chadcn.Skeleton(chadcn.SkeletonOptions{Height: overviewFooterSkeletonHeight, Rounded: true}),
		chadcn.Skeleton(chadcn.SkeletonOptions{Width: overviewFooterSkeletonWidth, Height: overviewFooterSkeletonHeight, Rounded: true}),
	)
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
