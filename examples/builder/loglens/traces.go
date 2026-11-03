package main

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	tracePercentPrimary = 62
	tracePercentSecond  = 41
	tracePercentThird   = 18
	traceTopologyRatio  = 4.2
	traceSkeletonHeight = 10
)

// Traces panel measurements, kept together so the layout reads as one block.
const (
	traceScrollHeight = 190
	traceRecentRows   = 10
	traceRowHeight    = 34
	traceCardGap      = 8
	traceCardPadding  = 12
)

func (a *app) tracesPanel() *ui.Element {
	tracks := []struct {
		id       string
		owner    string
		name     string
		duration string
		percent  int
	}{
		{id: "7f3a9c21", owner: "AK", name: "Ada Kim", duration: "1.24s", percent: tracePercentPrimary},
		{id: "5d21b8e0", owner: "RB", name: "Ravi Bose", duration: "820ms", percent: tracePercentSecond},
		{id: "a10e77c9", owner: "LT", name: "Linus Tan", duration: "360ms", percent: tracePercentThird},
	}
	active := max(a.traceIndex, 0) % len(tracks)
	track := tracks[active]

	selectors := make([]*ui.Element, 0, len(tracks))
	for i, candidate := range tracks {
		variant := chadcn.ButtonOutlined
		if i == active {
			variant = chadcn.ButtonFilled
		}
		selectors = append(selectors, chadcn.Button(candidate.id, chadcn.ButtonOptions{
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
			chadcn.Badge("trace", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			chadcn.Badge(track.id, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
			ui.Div().Class("app-col").Gap(ui.Px(3)).Children(
				chadcn.Progress(fmt.Sprintf("Latency budget for trace %s", track.id), track.percent).
					Width(ui.Px(168)),
				ui.Text(track.duration+" of a 2.0s budget").Class("app-muted"),
			),
			chadcn.Avatar(chadcn.AvatarOptions{Initials: track.owner, Status: chadcn.AvatarStatusOnline, Size: chadcn.AvatarSmall}),
			ui.Text(track.name+" · owner").Class("app-muted"),
			chadcn.Kbd("⌘T"),
		),
		chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Trace selector"}, selectors...),
	)

	rootName, focusName := "span pending", "span pending"
	if len(sampleSpans) > 0 {
		rootName = sampleSpans[0].Name
		focusName = rootName
		for _, span := range sampleSpans {
			if span.Status == "warning" {
				focusName = span.Name
				break
			}
		}
	}

	traceRows := make([]chadcn.TraceRow, 0, len(sampleSpans))
	for _, span := range sampleSpans {
		traceRows = append(traceRows, chadcn.TraceRow{
			Label:  span.Name,
			Detail: span.Service + " · " + span.Duration,
			Start:  span.Start,
			Width:  span.Width,
			Status: span.Status,
		})
	}

	waterfall := chadcn.Trace(chadcn.TraceOptions{Rows: traceRows})
	if len(sampleSpans) == 0 {
		waterfall = chadcn.Empty(chadcn.EmptyOptions{
			Icon:        ui.Text("⌁"),
			Title:       "No spans for this trace",
			Description: "The sampler kept the trace ID but no span reached the collector.",
		})
	}

	attributeRows := []*ui.Element{
		chadcn.TableRow(chadcn.TableCell("trace.id"), chadcn.TableCell(track.id)),
		chadcn.TableRow(chadcn.TableCell("span.name"), chadcn.TableCell(focusName)),
		chadcn.TableRow(chadcn.TableCell("db.system"), chadcn.TableCell("postgresql")),
		chadcn.TableRow(chadcn.TableCell("latency"),
			chadcn.TableCell(fmt.Sprintf("%s · %d%% of budget", track.duration, track.percent))),
	}

	waterfallCard := chadcn.Card(
		chadcn.CardHeader(
			chadcn.H2("Span waterfall").FontSize(ui.Px(16)),
			chadcn.CardDescription(fmt.Sprintf("%d spans · %s end to end", len(sampleSpans), track.duration)),
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
  "status": "ok"
}`, track.id, rootName, track.duration, len(sampleSpans))

	payloadCard := chadcn.Card(
		chadcn.CardContent(
			chadcn.CodeBlock(chadcn.CodeBlockOptions{
				Filename: "trace-" + track.id + ".json",
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

	recentCount := min(len(sampleLogs), traceRecentRows)
	recentRows := make([]*ui.Element, 0, recentCount)
	for i := 0; i < recentCount; i++ {
		entry := sampleLogs[i]
		recentRows = append(recentRows, ui.Div().Class("app-row").
			Height(ui.Px(traceRowHeight)).
			Padding(ui.Edges(ui.Px(0), ui.Px(10))).
			Children(
				ui.Text("⌁").Color(ui.ColorVar("--of-muted", ui.Hex("#737373"))),
				ui.Text(entry.Trace).FontSize(ui.Px(12)).FontFamily(ui.FontFamilyMonospace),
				chadcn.Badge(entry.Service, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
				ui.Div().Class("app-spacer"),
				chadcn.Badge(entry.Duration, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
				ui.Text(entry.Time).Class("app-muted"),
			))
	}
	a.traceScroll.SetRange(recentCount*traceRowHeight, traceScrollHeight)

	recentCard := ui.Div().Class("app-col").Children(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Text("Recent traces").Class("app-section-label"),
			chadcn.Muted(fmt.Sprintf("%d of 128 · last 15 minutes", recentCount)),
		),
		chadcn.ScrollArea(chadcn.ScrollAreaOptions{
			ID:          "trace-scroll",
			Label:       "Recent traces",
			State:       &a.traceScroll,
			TrackPixels: traceScrollHeight,
			Height:      ui.Px(traceScrollHeight),
			Children:    recentRows,
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

	ltrCard := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Arabic · forced LTR"),
			chadcn.CardDescription("For a log line that ignores the reader's script."),
		),
		chadcn.CardContent(
			chadcn.Direction(chadcn.DirectionOptions{Mode: chadcn.DirectionLTR},
				chadcn.P(arabic).FontSize(ui.Px(13)).LineHeight(ui.Number(1.7)),
				chadcn.Muted("billing.charge · 640ms · declined"),
			),
		),
		chadcn.CardFooter(
			chadcn.Label("dir=ltr"),
			ui.Div().Class("app-spacer"),
			chadcn.Kbd("⌘L"),
		),
	).Gap(ui.Px(traceCardGap)).Padding(ui.Px(traceCardPadding))

	topologyCard := chadcn.Card(
		chadcn.CardTitle("Span topology"),
		chadcn.CardContent(
			chadcn.AspectRatio(chadcn.AspectRatioOptions{Ratio: traceTopologyRatio},
				ui.Div().Class("app-canvas").Children(
					ui.Text("web → api-gateway → postgres · 1.24s · 5 spans").FontSize(ui.Px(13)),
				),
			),
		),
	).Gap(ui.Px(traceCardGap)).Padding(ui.Px(traceCardPadding))

	indexed, _ := a.extra["traces.indexed"].(bool)
	var indexBody *ui.Element
	if indexed {
		indexBody = ui.Div().Class("app-row").Children(
			chadcn.Badge("ready", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			ui.Text("1,284 spans · 42ms").Class("app-muted"),
			ui.Div().Class("app-spacer"),
			chadcn.Kbd("⌘I"),
		)
	} else {
		indexBody = ui.Div().Class("app-col").Gap(ui.Px(9)).Children(
			ui.Div().Class("app-row").Children(
				chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerSmall, Label: "Indexing spans"}),
				chadcn.Label("Indexing span attributes"),
				ui.Div().Class("app-spacer"),
				chadcn.Badge("streaming", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			),
			chadcn.Skeleton(chadcn.SkeletonOptions{Height: traceSkeletonHeight, Rounded: true}),
			chadcn.Skeleton(chadcn.SkeletonOptions{Height: traceSkeletonHeight, Rounded: true}).Width(ui.Percent(72)),
			chadcn.Skeleton(chadcn.SkeletonOptions{Height: traceSkeletonHeight, Rounded: true}).Width(ui.Percent(44)),
		)
	}
	indexCard := chadcn.Card(
		chadcn.CardTitle("Span index"),
		chadcn.CardContent(indexBody),
		chadcn.CardFooter(
			chadcn.Switch("Warm index", indexed).ID("trace-index-switch").
				OnClick(func(ctx context.Context) error {
					a.extra["traces.indexed"] = !indexed
					return a.refresh(ctx)
				}),
			ui.Div().Class("app-spacer"),
			chadcn.Label("Updated 14:32"),
		),
	).Gap(ui.Px(traceCardGap)).Padding(ui.Px(traceCardPadding))

	action := chadcn.Item(chadcn.ItemOptions{
		Media: chadcn.Avatar(chadcn.AvatarOptions{Initials: track.owner, Status: chadcn.AvatarStatusOnline, Size: chadcn.AvatarSmall}),
		Title: "Own trace " + track.id,
		Description: fmt.Sprintf("%s · %s · %d spans · last touched 6 minutes ago",
			track.name, track.duration, len(sampleSpans)),
		Actions: chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Trace actions"},
			chadcn.Button("Assign to me", chadcn.ButtonOptions{
				Size:    chadcn.ButtonSmall,
				Variant: chadcn.ButtonOutlined,
			}).OnClick(func(ctx context.Context) error {
				return a.showToast(ctx, "Assigned", "Trace "+track.id+" is now owned by you.")
			}),
			chadcn.Button("Share", chadcn.ButtonOptions{
				Size:    chadcn.ButtonSmall,
				Variant: chadcn.ButtonOutlined,
			}).OnClick(func(ctx context.Context) error {
				return a.showToast(ctx, "Link copied", "trace-"+track.id+" was copied to the clipboard.")
			}),
		),
	}).Padding(ui.Px(traceCardPadding))

	return ui.Section().Class("app-stack").Children(
		header,
		chadcn.Lead("One request, five spans, and the payload the collector kept."),
		ui.Div().Class("app-split").Children(
			chadcn.Resizable(chadcn.ResizableOptions{
				ID: "trace-split", Ratio: a.splitRatio, First: waterfallCard, Second: payloadCard,
			}),
		),
		recentCard,
		ui.Div().Class("app-card-grid").Children(directionCard, ltrCard),
		ui.Div().Class("app-grid-2").Children(topologyCard, indexCard),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		action,
	)
}
