package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	attachmentComplete = 100
	impactRatio        = 16.0 / 7.0
	skeletonLineHeight = 12
	skeletonWideWidth  = 180
	skeletonMidWidth   = 120
)

const (
	incidentChatHeight    = 320
	incidentSnapshotCount = 3
	incidentMitigation    = 68
	incidentReplyID       = "incident-reply"
)

// incidentPanel renders the incident room: the bridge thread and open
// decisions on the left, the playbook and live snapshots on the right, then
// the response timeline and the postmortem draft.
func (a *app) incidentPanel() *ui.Element {
	if _, seeded := a.extra["incident.seeded"]; !seeded {
		a.extra["incident.seeded"] = true
		a.accordionOpen = []string{"triage"}
	}
	return ui.Div().Class("app-stack").Children(
		incidentHeader(a),
		ui.Div().Class("app-grid-2").Children(
			ui.Div().Class("app-stack").Children(
				incidentChatCard(),
				incidentDecisionCard(a),
				incidentActionCard(),
			),
			ui.Div().Class("app-stack").Children(
				incidentPlaybookCard(a),
				incidentSnapshotCard(a),
				incidentCustomerNoteCard(),
				incidentImpactCard(),
			),
		),
		ui.Div().Class("app-grid-2").Children(
			incidentTimelineCard(),
			incidentDraftCard(),
		),
	)
}

func incidentHeader(a *app) *ui.Element {
	// The identity block grows and the actions keep their width, so the row
	// stays on one line: a growing spacer between them stole the space the
	// buttons needed and wrapped them under the title.
	return chadcn.Card(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Div().Class("app-row").Flex(ui.Number(1)).MinWidth(ui.Px(0)).Children(
				chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK", Alt: "Ada Kim, incident commander", Status: chadcn.AvatarStatusOnline}),
				ui.Div().Class("app-col").Gap(ui.Px(4)).Children(
					ui.Div().Class("app-row").Children(
						chadcn.H3("Elevated 504s on api-gateway"),
						chadcn.Badge("SEV-2", chadcn.BadgeOptions{Variant: chadcn.BadgeDanger}),
						chadcn.Badge("investigating", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
					),
					ui.Text("Opened 14:27 · Ada Kim on call · 6 responders on the bridge").Class("app-muted"),
				),
			),
			chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Incident actions"},
				chadcn.Button("Acknowledge", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).OnClick(func(ctx context.Context) error {
					return a.showToast(ctx, "Acknowledged", "Ada Kim was paged again at 14:33. The next update is yours.")
				}),
				chadcn.Button("Resolve", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).OnClick(func(ctx context.Context) error {
					return a.showToast(ctx, "Resolve blocked", "Two action items are still open. Close them before resolving.")
				}),
			),
		),
		ui.Div().Class("app-col").Children(
			ui.Div().Class("app-row").Children(
				chadcn.Label("Mitigation progress"),
				ui.Div().Class("app-spacer"),
				chadcn.Muted("4 of 6 steps complete · next update due 14:36"),
			),
			chadcn.Progress("Mitigation progress", incidentMitigation),
		),
	)
}

func incidentChatCard() *ui.Element {
	rows := make([]*ui.Element, 0, len(sampleMessages)+1)
	for _, message := range sampleMessages {
		align, variant := chadcn.MessageStart, chadcn.BubbleDefault
		if message.Side == "end" {
			align, variant = chadcn.MessageEnd, chadcn.BubblePrimary
		}
		parts := []*ui.Element{chadcn.Bubble(chadcn.BubbleOptions{Variant: variant}, ui.Text(message.Text))}
		if message.Attachment != "" {
			parts = append(parts, chadcn.Attachment(chadcn.AttachmentOptions{Name: message.Attachment, Meta: "848 KB · heap dump", Progress: attachmentComplete}))
		}
		rows = append(rows, chadcn.Message(chadcn.MessageOptions{
			Name:     message.Name,
			Time:     message.Time,
			Align:    align,
			Avatar:   chadcn.Avatar(chadcn.AvatarOptions{Initials: message.Initials, Size: chadcn.AvatarSmall}),
			Children: parts,
		}))
	}
	rows = append(rows, chadcn.Marker(chadcn.MarkerOptions{Label: "Status page updated", Detail: "14:32 · EU tenants see an investigating banner", State: chadcn.MarkerSuccess}))
	return chadcn.Card(
		ui.Div().Class("app-row").Children(
			chadcn.CardTitle("Incident channel"),
			chadcn.Badge("#inc-eu-gateway", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			ui.Div().Class("app-spacer"),
			chadcn.Muted("6 responders · bridge open"),
		),
		ui.Div().Class("app-chat").Children(
			chadcn.MessageScroller(chadcn.MessageScrollerOptions{Height: ui.Px(incidentChatHeight), Children: rows}),
		),
		ui.Div().Class("app-chat-row app-chat-row--end").Children(
			chadcn.Avatar(chadcn.AvatarOptions{Initials: "YO", Size: chadcn.AvatarSmall}),
			ui.Div().Class("app-chat-body").Children(
				ui.Div().Class("app-chat-meta").Children(
					chadcn.Label("Reply in the incident channel", chadcn.LabelOptions{For: incidentReplyID}),
					ui.Text(" · "),
					chadcn.Kbd("⌘↵"),
				),
				chadcn.InputWithOptions(incidentReplyID, "", chadcn.InputOptions{Placeholder: "Post an update to the bridge…", Size: chadcn.InputSmall}),
			),
		),
	)
}

func incidentDecisionCard(a *app) *ui.Element {
	return chadcn.Card(chadcn.Questionnaire(chadcn.QuestionnaireOptions{
		Title:       "Decisions to record",
		Description: "Answers copy into the timeline and the postmortem draft.",
		OnSelect: func(ctx context.Context, _ string, value string) error {
			a.questionAnswer = value
			return a.refresh(ctx)
		},
		Questions: []chadcn.Question{
			{
				Label:    "What is the blast radius?",
				Help:     "Use the answer in the 14:36 status update.",
				Selected: a.questionAnswer,
				Choices: []chadcn.ChoiceItem{
					{Value: "eu-only", Label: "EU tenants only", Detail: "0.8% of EU requests failing · 3 enterprise accounts"},
					{Value: "global", Label: "All regions", Detail: "US-East p99 is flat at 210ms"},
				},
			},
			{
				Label:    "Which mitigation ships first?",
				Help:     "The room executes the selected step.",
				Selected: a.questionAnswer,
				Choices: []chadcn.ChoiceItem{
					{Value: "scale-workers", Label: "Scale the worker pool", Detail: "8 → 16 replicas · queue already draining"},
					{Value: "rollback", Label: "Roll back gateway 2.14.1", Detail: "Last known good deploy at 13:40"},
				},
			},
		},
	}))
}

func incidentActionCard() *ui.Element {
	return chadcn.Card(
		chadcn.CardTitle("Action items", chadcn.Badge("3 open", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})),
		chadcn.Item(chadcn.ItemOptions{
			Media:       chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK", Size: chadcn.AvatarSmall}),
			Title:       "Verify autoscaling caught the queue",
			Description: "Platform · due 16:00 · confirm p99 is back under 800ms",
			Actions:     chadcn.Badge("in progress", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
		}),
		chadcn.Item(chadcn.ItemOptions{
			Media:       chadcn.Avatar(chadcn.AvatarOptions{Initials: "LT", Size: chadcn.AvatarSmall}),
			Title:       "Attach the gateway trace bundle",
			Description: "Observability · due 15:30 · 504 trace for the EU cluster",
			Actions:     chadcn.Badge("blocked", chadcn.BadgeOptions{Variant: chadcn.BadgeDanger}),
		}),
		chadcn.Item(chadcn.ItemOptions{
			Media:       chadcn.Avatar(chadcn.AvatarOptions{Initials: "YO", Size: chadcn.AvatarSmall}),
			Title:       "Draft the 16:00 customer update",
			Description: "Comms · due 15:45 · reuse the 14:32 wording",
			Actions:     chadcn.Badge("queued", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
		}),
	)
}

func incidentPlaybookCard(a *app) *ui.Element {
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Response playbook", chadcn.Badge("3 steps", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})),
			chadcn.CardDescription("Run in order. Every expansion is visible to the room."),
		),
		chadcn.Accordion(chadcn.AccordionOptions{
			Open: a.accordionOpen,
			OnToggle: func(ctx context.Context, id string) error {
				a.accordionOpen = incidentToggle(a.accordionOpen, id)
				return a.refresh(ctx)
			},
			Items: []chadcn.AccordionItem{
				{ID: "triage", Title: "1 · Confirm the blast radius", Content: chadcn.P("Compare the 504 rate against the EU tenant split and the gateway p99 from the last 15 minutes.")},
				{ID: "mitigate", Title: "2 · Drain the worker queue", Content: chadcn.P("Scale api-gateway workers 8 → 16, hold the 13:40 deploy, and watch queue depth fall under 500.")},
				{ID: "comms", Title: "3 · Update customers", Content: chadcn.P("Post the EU impact note to the status page, then schedule the 14:36 update here.")},
			},
		}),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		chadcn.Collapsible(chadcn.CollapsibleOptions{
			ID:   "incident-advanced",
			Open: a.collapsibleOpen,
			OnOpenChange: func(ctx context.Context, open bool) error {
				a.collapsibleOpen = open
				return a.refresh(ctx)
			},
			Trigger: ui.Div().Class("app-row").Children(
				chadcn.Button("Advanced notes", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}),
				chadcn.Kbd("A"),
				ui.Div().Class("app-spacer"),
				chadcn.Badge("runbook", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			),
		},
			ui.Div().Class("app-col").Children(
				chadcn.Muted("Keep the bridge open while the pool drains. Page the vendor only after queue depth is under 500."),
				chadcn.InlineCode("handoff: ada → linus · 15:00 UTC"),
			),
		),
	)
}

func incidentSnapshotCard(a *app) *ui.Element {
	return chadcn.Card(
		chadcn.CardTitle("Live snapshots"),
		chadcn.Carousel(chadcn.CarouselOptions{
			ID:     "incident-snapshots",
			Label:  "Incident snapshots",
			Index:  a.carouselIndex,
			Slides: incidentSnapshotSlides(),
			OnPrev: func(ctx context.Context) error {
				if a.carouselIndex > 0 {
					a.carouselIndex--
				}
				return a.refresh(ctx)
			},
			OnNext: func(ctx context.Context) error {
				if a.carouselIndex < incidentSnapshotCount-1 {
					a.carouselIndex++
				}
				return a.refresh(ctx)
			},
		}),
	)
}

func incidentSnapshotSlides() []*ui.Element {
	snapshots := []struct {
		Title    string
		Value    string
		Badge    string
		Detail   string
		Variant  chadcn.BadgeVariant
		Progress int
	}{
		{"Gateway errors", "4.8%", "critical", "58K of 1.2M requests · SLO 0.5%", chadcn.BadgeDanger, 96},
		{"p99 latency", "1.42s", "watching", "EU p99 · SLO 800ms", chadcn.BadgeOutline, 71},
		{"Worker queue", "1,842", "draining", "pending jobs · was 2,410 at 14:28", chadcn.BadgeSecondary, 46},
	}
	slides := make([]*ui.Element, 0, len(snapshots))
	for _, snapshot := range snapshots {
		slides = append(slides, ui.Div().Class("app-col").Width(ui.Percent(100)).Children(
			ui.Div().Class("app-row").Children(
				chadcn.H4(snapshot.Title),
				chadcn.Badge(snapshot.Badge, chadcn.BadgeOptions{Variant: snapshot.Variant}),
			),
			ui.Div().Class("app-kpi-value").Text(snapshot.Value),
			chadcn.Muted(snapshot.Detail),
			chadcn.Progress(snapshot.Title, snapshot.Progress),
		))
	}
	return slides
}

func incidentCustomerNoteCard() *ui.Element {
	return chadcn.Card(
		ui.Div().Class("app-section-label").Text("Customer note · AR"),
		chadcn.Direction(chadcn.DirectionOptions{Mode: chadcn.DirectionRTL},
			ui.Div().Class("app-rtl app-col").Children(
				chadcn.Label("تحديث حالة الخدمة"),
				ui.Text("نراقب معدل الأخطاء في المصادقة والفوترة. سنحدّث صفحة الحالة كل 15 دقيقة.").Class("app-muted"),
			),
		),
	)
}

func incidentImpactCard() *ui.Element {
	return chadcn.Card(
		ui.Div().Class("app-row").Children(
			chadcn.CardTitle("Impact map"),
			chadcn.Badge("eu-west-1", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
			ui.Div().Class("app-spacer"),
			chadcn.Muted("3 enterprise accounts"),
		),
		chadcn.AspectRatio(chadcn.AspectRatioOptions{Ratio: impactRatio},
			ui.Div().Class("app-canvas").Width(ui.Percent(100)).Height(ui.Percent(100)).Children(
				ui.Div().Class("app-col").AlignItems(ui.AlignItemsCenter).Children(
					ui.Div().Class("app-row").Children(
						chadcn.Badge("web", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
						ui.Text("→").FontSize(ui.Px(11)),
						chadcn.Badge("api-gateway", chadcn.BadgeOptions{Variant: chadcn.BadgeDanger}),
						ui.Text("→").FontSize(ui.Px(11)),
						chadcn.Badge("worker", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
					),
					chadcn.Muted("auth and billing healthy · search unaffected"),
				),
			),
		),
	)
}

func incidentTimelineCard() *ui.Element {
	return chadcn.Card(
		chadcn.CardTitle("Response timeline", chadcn.Badge("4 events", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})),
		ui.Div().Class("app-table-wrap").Children(chadcn.Table(
			chadcn.TableHeader(chadcn.TableRow(
				chadcn.TableHead("Time"),
				chadcn.TableHead("Event"),
				chadcn.TableHead("Owner"),
				chadcn.TableHead("Status"),
			)),
			chadcn.TableBody(
				incidentTimelineRow("14:27", "SEV-2 paged · 504s above 4% on api-gateway", "PagerDuty", "paged", chadcn.BadgeDanger),
				incidentTimelineRow("14:28", "Ada Kim acknowledged and opened the bridge", "Ada Kim", "acknowledged", chadcn.BadgeSecondary),
				incidentTimelineRow("14:30", "Worker pool scaled 8 → 16 replicas", "You", "mitigating", chadcn.BadgeOutline),
				incidentTimelineRow("14:32", "Status page updated for EU tenants", "Comms", "customer facing", chadcn.BadgeSecondary),
			),
		)),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		ui.Div().Class("app-section-label").Text("Next up"),
		ui.Div().Class("app-timeline").Children(
			chadcn.Marker(chadcn.MarkerOptions{Label: "14:36 · public update due", Detail: "Comms drafts, you approve", State: chadcn.MarkerWarning}),
			chadcn.Marker(chadcn.MarkerOptions{Label: "15:00 · handoff to EMEA", Detail: "Ada → Linus", State: chadcn.MarkerDefault}),
			chadcn.Marker(chadcn.MarkerOptions{Label: "16:00 · postmortem draft", Detail: "Owner: you · template ready", State: chadcn.MarkerDefault}),
		),
	)
}

func incidentTimelineRow(at, event, owner, status string, variant chadcn.BadgeVariant) *ui.Element {
	return chadcn.TableRow(
		chadcn.TableCell(at),
		chadcn.TableCell(event),
		chadcn.TableCell(owner),
		chadcn.TableCell("", chadcn.Badge(status, chadcn.BadgeOptions{Variant: variant})),
	)
}

func incidentDraftCard() *ui.Element {
	return chadcn.Card(
		chadcn.CardTitle("Postmortem", chadcn.Badge("draft", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})),
		ui.Div().Class("app-col").Children(
			ui.Div().Class("app-row").Children(
				chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerSmall, Label: "Assembling the trace bundle"}),
				chadcn.Muted("Assembling the trace bundle"),
			),
			chadcn.Skeleton(chadcn.SkeletonOptions{Height: skeletonLineHeight}),
			chadcn.Skeleton(chadcn.SkeletonOptions{Width: skeletonWideWidth, Height: skeletonLineHeight}),
			chadcn.Skeleton(chadcn.SkeletonOptions{Width: skeletonMidWidth, Height: skeletonLineHeight}),
		),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		chadcn.Empty(chadcn.EmptyOptions{
			Icon:        ui.Text("◇"),
			Title:       "No draft yet",
			Description: "Draft unlocks when every action item closes.",
		}),
	)
}

// incidentToggle adds or removes one accordion item ID.
func incidentToggle(open []string, id string) []string {
	for index, value := range open {
		if value == id {
			return append(open[:index], open[index+1:]...)
		}
	}
	return append(open, id)
}
