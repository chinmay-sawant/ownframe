package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Traces panel measurements, kept together so the layout reads as one block.
const (
	traceRowHeight   = 34
	traceCardGap     = 8
	traceCardPadding = 12

	traceSelectorMax = 3

	traceBudgetWindow = 2 * time.Second
)

func traceBudgetPercent(duration string) int {
	value, err := time.ParseDuration(duration)
	if err != nil {
		return 0
	}
	return max(0, min(telemetryPercentScale, int(value*telemetryPercentScale/traceBudgetWindow)))
}

func traceStatusVariant(status string) chadcn.BadgeVariant {
	switch status {
	case "error":
		return chadcn.BadgeDanger
	case "warning":
		return chadcn.BadgeOutline
	default:
		return chadcn.BadgeSecondary
	}
}

func (a *app) tracesPanel() *ui.Element {
	traces, traceCount, tracePages := a.feed.tracePage(a.tracePageIndex)
	if len(traces) == 0 {
		traces, traceCount, tracePages = a.feed.tracePage(0)
	}
	active := max(0, min(a.traceIndex, len(traces)-1))
	track := traces[active]

	selectors := make([]*ui.Element, 0, min(traceSelectorMax, len(traces)))
	for i, candidate := range traces[:min(traceSelectorMax, len(traces))] {
		variant := chadcn.ButtonOutlined
		if i == active {
			variant = chadcn.ButtonFilled
		}
		selectors = append(selectors, chadcn.Button(candidate.ID, chadcn.ButtonOptions{
			Size:    chadcn.ButtonSmall,
			Variant: variant,
		}).OnClick(func(ctx context.Context) error {
			a.traceIndex = i
			a.traceSpan = ""
			return a.refresh(ctx)
		}))
	}

	header := ui.Div().Class("app-row app-row--between").Children(
		ui.Div().Class("app-row").Children(
			chadcn.Badge("generated trace", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			chadcn.Badge(track.ID, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
			ui.Div().Class("app-col").Gap(ui.Px(3)).Children(
				chadcn.Progress(fmt.Sprintf("Latency budget for trace %s", track.ID), traceBudgetPercent(track.Duration)).
					Width(ui.Px(168)),
				ui.Text(track.Duration+" sampled").Class("app-muted"),
			),
			chadcn.Avatar(chadcn.AvatarOptions{Initials: track.Owner, Status: chadcn.AvatarStatusOnline, Size: chadcn.AvatarSmall}),
			ui.Text(track.Owner+" · generated trace").Class("app-muted"),
			chadcn.Kbd("⌘T"),
		),
		chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Trace selector"}, selectors...),
	)

	rootName, focusName := "span pending", "span pending"
	if len(track.Spans) > 0 {
		rootName = track.Spans[0].Name
		focusName = rootName
		for _, span := range track.Spans {
			if span.Status == "warning" {
				focusName = span.Name
				break
			}
		}
	}

	traceRows := make([]chadcn.TraceRow, 0, len(track.Spans))
	for _, span := range track.Spans {
		traceRows = append(traceRows, chadcn.TraceRow{
			Label:  span.Name,
			Detail: span.Service + " · " + span.Duration,
			Start:  span.Start,
			Width:  span.Width,
			Status: span.Status,
		})
	}

	waterfall := chadcn.Trace(chadcn.TraceOptions{Rows: traceRows})
	if len(track.Spans) == 0 {
		waterfall = chadcn.Empty(chadcn.EmptyOptions{
			Icon:        ui.Text("⌁"),
			Title:       "No spans for this trace",
			Description: "The sampler kept the trace ID but no span reached the collector.",
		})
	}

	attributeRows := []*ui.Element{
		chadcn.TableRow(chadcn.TableCell("trace.id"), chadcn.TableCell(track.ID)),
		chadcn.TableRow(chadcn.TableCell("span.name"), chadcn.TableCell(focusName)),
		chadcn.TableRow(chadcn.TableCell("db.system"), chadcn.TableCell("postgresql")),
		chadcn.TableRow(chadcn.TableCell("latency"),
			chadcn.TableCell(track.Duration+" · "+track.Status)),
	}

	waterfallCard := chadcn.Card(
		chadcn.CardHeader(
			chadcn.H2("Span waterfall").FontSize(ui.Px(16)),
			chadcn.CardDescription(fmt.Sprintf("%d spans · %s end to end", len(track.Spans), track.Duration)),
		),
		chadcn.CardContent(
			ui.Div().Class("app-col").Gap(ui.Px(12)).Children(
				waterfall,
				ui.Div().Class("app-table-wrap").Children(
					chadcn.Table(
						chadcn.TableHeader(chadcn.TableRow(
							chadcn.TableHead("Attribute"),
							chadcn.TableHead("Value"),
						)),
						chadcn.TableBody(attributeRows...),
					),
				),
			),
		),
	).Gap(ui.Px(traceCardGap)).Padding(ui.Px(traceCardPadding))

	scope := a.traceSpan
	if scope == "" {
		scope = rootName
	}
	payload := fmt.Sprintf(`{
  "traceId": "%s",
  "root": "%s",
  "duration": "%s",
  "spanCount": %d,
  "sampled": true,
  "status": "%s"
}`, track.ID, rootName, track.Duration, len(track.Spans), track.Status)

	payloadCard := chadcn.Card(
		chadcn.CardContent(
			chadcn.CodeBlock(chadcn.CodeBlockOptions{
				Filename: "trace-" + track.ID + ".json",
				Language: "json",
				Code:     payload,
			}),
		),
		chadcn.CardFooter(
			chadcn.Button("Focus slow span", chadcn.ButtonOptions{
				Size:    chadcn.ButtonSmall,
				Variant: chadcn.ButtonText,
			}).OnClick(func(ctx context.Context) error {
				a.traceSpan = focusName
				return a.showToast(ctx, "Span focused", focusName+" is pinned in the payload footer.")
			}),
			ui.Div().Class("app-spacer"),
			chadcn.Label("Span scope · "+scope),
			chadcn.Kbd("⌘J"),
		),
	).Gap(ui.Px(traceCardGap)).Padding(ui.Px(traceCardPadding))

	recentRows := make([]*ui.Element, 0, len(traces))
	for i, entry := range traces {
		row := ui.Div().ID("recent-trace-"+entry.ID).Class("app-row app-trace-row").
			Attr("role", "button").Attr("tabindex", "0").
			Height(ui.Px(traceRowHeight)).
			Padding(ui.Edges(ui.Px(0), ui.Px(10))).
			Children(
				ui.Text("⌁").Color(ui.ColorVar("--of-muted", ui.Hex("#737373"))),
				ui.Text(entry.ID).FontSize(ui.Px(12)).FontFamily(ui.FontFamilyMonospace),
				ui.Text(entry.Name).FontSize(ui.Px(12)),
				chadcn.Badge(entry.Service, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
				ui.Div().Class("app-spacer"),
				chadcn.Badge(entry.Duration, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
				chadcn.Badge(entry.Status, chadcn.BadgeOptions{Variant: traceStatusVariant(entry.Status)}),
				ui.Text(entry.Timestamp.Format("15:04:05")).Class("app-muted"),
			)
		row.OnClick(func(ctx context.Context) error {
			a.traceIndex = i
			a.traceSpan = ""
			return a.refresh(ctx)
		})
		recentRows = append(recentRows, row)
	}

	recentCard := ui.Div().Class("app-col").Children(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Text("Recent traces · generated demo feed").Class("app-section-label"),
			chadcn.Muted(fmt.Sprintf("page %d of %d · %d total", a.tracePageIndex+1, tracePages, traceCount)),
		),
		ui.Div().Class("app-col").Children(recentRows...),
		chadcn.Pagination(chadcn.PaginationOptions{
			Page: a.tracePageIndex + 1, TotalPages: tracePages,
			OnSelect: func(ctx context.Context, page int) error {
				a.tracePageIndex = page - 1
				a.traceIndex = 0
				a.traceSpan = ""
				return a.refresh(ctx)
			},
		}),
	)

	mode, directionNote := chadcn.DirectionLTR, "Left to right is active."
	if a.rtl {
		mode, directionNote = chadcn.DirectionRTL, "Right to left is active."
	}
	arabic := "تعذّر الاتصال بخدمة الفوترة بعد ثلاث محاولات، لذلك أُعيد الطلب إلى قائمة الانتظار."
	directionCard := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Arabic · live direction"),
			chadcn.CardDescription(directionNote),
		),
		chadcn.CardContent(
			chadcn.Direction(chadcn.DirectionOptions{Mode: mode},
				chadcn.P(arabic).FontSize(ui.Px(13)).LineHeight(ui.Number(1.7)),
				chadcn.Muted("billing.charge · 640ms · declined"),
			),
		),
		chadcn.CardFooter(
			chadcn.Switch("Right to left", a.rtl).ID("trace-rtl-switch").
				OnClick(func(ctx context.Context) error {
					a.rtl = !a.rtl
					return a.refresh(ctx)
				}),
			ui.Div().Class("app-spacer"),
			chadcn.Kbd("⌘R"),
		),
	).Gap(ui.Px(traceCardGap)).Padding(ui.Px(traceCardPadding))

	action := chadcn.Item(chadcn.ItemOptions{
		Media: chadcn.Avatar(chadcn.AvatarOptions{Initials: track.Owner, Status: chadcn.AvatarStatusOnline, Size: chadcn.AvatarSmall}),
		Title: "Own trace " + track.ID,
		Description: fmt.Sprintf("%s · %s · %d spans · sampled %s",
			track.Name, track.Duration, len(track.Spans), track.Timestamp.Format("15:04:05")),
		Actions: chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Trace actions"},
			chadcn.Button("Assign to me", chadcn.ButtonOptions{
				Size:    chadcn.ButtonSmall,
				Variant: chadcn.ButtonOutlined,
			}).OnClick(func(ctx context.Context) error {
				return a.showToast(ctx, "Assigned", "Trace "+track.ID+" is now owned by you.")
			}),
			chadcn.Button("Share", chadcn.ButtonOptions{
				Size:    chadcn.ButtonSmall,
				Variant: chadcn.ButtonOutlined,
			}).OnClick(func(ctx context.Context) error {
				return a.showToast(ctx, "Link copied", "trace-"+track.ID+" was copied to the clipboard.")
			}),
		),
	}).Padding(ui.Px(traceCardPadding))

	return ui.Section().Class("app-stack").Children(
		header,
		chadcn.Lead("Generated request spans, their timing, and the sampled payload."),
		ui.Div().Class("app-split").Children(
			chadcn.Resizable(chadcn.ResizableOptions{
				ID: "trace-split", Ratio: a.splitRatio, First: waterfallCard, Second: payloadCard,
			}),
		),
		recentCard,
		directionCard,
		chadcn.Separator(chadcn.SeparatorOptions{}),
		action,
	)
}
