package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const logRowPlaceholderHeight = 46

// logsPanel is the request stream. The filter bar comes from logFilters; the
// count row, live tail, row menu, paged table, and detail drawer live here.
func (a *app) logsPanel() *ui.Element {
	logs := a.visibleLogs()
	rows := a.pageLogs()

	errors, warns := 0, 0
	for _, entry := range logs {
		switch entry.Level {
		case "error":
			errors++
		case "warn":
			warns++
		}
	}

	// The row menu targets the first selected row on this page, otherwise the
	// first row, so its actions always have a log to work on.
	pageStart := a.pageIndex * logPageSize
	if pageStart >= len(logs) {
		pageStart = max(0, (a.logPageCount()-1)*logPageSize)
	}
	detailIndex := pageStart
	for index := range rows {
		if a.selectedRows[index] {
			detailIndex = pageStart + index
			break
		}
	}
	if detailIndex >= len(logs) {
		detailIndex = max(0, len(logs)-1)
	}

	return ui.Div().Class("app-stack").Children(
		a.logFilters(),
		a.logsErrorAlert(logs, errors, warns),
		a.logsToolbar(logs, detailIndex),
		a.logsTable(rows),
	)
}

// logsErrorAlert reports the failure share of the visible rows and offers a
// one-click rule. It is nil while every visible row is below error level.
func (a *app) logsErrorAlert(logs []logEntry, errors, warns int) *ui.Element {
	if errors == 0 {
		return nil
	}
	latest := logs[0]
	for _, entry := range logs {
		if entry.Level == "error" {
			latest = entry
			break
		}
	}
	const percentBase = 100
	share := percentBase * float64(errors) / float64(len(logs))
	action := chadcn.Button("Create alert", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).OnClick(func(ctx context.Context) error {
		return a.showToast(ctx, "Alert drafted",
			fmt.Sprintf("A %s rule watches this view while errors stay above 1%% for 5 minutes.", latest.Service))
	})

	return chadcn.Alert(chadcn.AlertOptions{
		Icon:        ui.Text("!"),
		Title:       fmt.Sprintf("%d errors in this view", errors),
		Description: fmt.Sprintf("%.1f%% of %d entries failed · %d warnings · newest %s at %s (%d)", share, len(logs), warns, latest.Service, latest.Time, latest.Status),
		Variant:     chadcn.AlertDanger,
		Action:      action,
	})
}

// logsToolbar carries the count, the export and share controls, the live-tail
// switch, the row menu trigger, and the page selector.
func (a *app) logsToolbar(logs []logEntry, detailIndex int) *ui.Element {
	target := logEntry{}
	if len(logs) > 0 {
		target = logs[detailIndex]
	}
	selected := 0
	for _, on := range a.selectedRows {
		if on {
			selected++
		}
	}
	stats := []*ui.Element{
		ui.Text("Request stream").FontWeight(ui.FontWeightBold).FontSize(ui.Px(14)),
		chadcn.Badge(fmt.Sprintf("%d entries", len(logs)), chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
	}
	if selected > 0 {
		stats = append(stats, chadcn.Badge(fmt.Sprintf("%d selected", selected), chadcn.BadgeOptions{}))
	}
	if len(a.severity) > 0 {
		stats = append(stats, chadcn.Badge(strings.Join(a.severity, " + "), chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}))
	}
	if a.loading {
		stats = append(stats, chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerSmall, Label: "Loading logs"}))
	}
	menu := chadcn.ContextMenu(chadcn.ContextMenuOptions{
		ID:    "logs-row-menu",
		Label: "Row actions",
		Open:  a.contextOpen,
		X:     a.contextX,
		Y:     a.contextY,
		Items: []chadcn.MenuItem{
			{Label: "Open detail", Shortcut: "↵", Disabled: len(logs) == 0},
			{Label: "Copy trace", Shortcut: "⌘C", Disabled: len(logs) == 0},
			{Separator: true},
			{Label: "Create alert", Shortcut: "⌘A", Disabled: len(logs) == 0},
		},
		OnSelect: func(ctx context.Context, label string) error {
			a.contextOpen = false
			switch label {
			case "Open detail":
				a.drawerIndex = detailIndex
				a.drawerOpen = true
				return a.refresh(ctx)
			case "Copy trace":
				return a.showToast(ctx, "Trace copied", target.Trace+" · "+target.Service+" is on the clipboard.")
			case "Create alert":
				return a.showToast(ctx, "Alert drafted", "Watching "+target.Service+" for errors above 1% for 5 minutes.")
			}
			return a.refresh(ctx)
		},
	})
	rowActions := ui.Div().Position(ui.PositionRelative).Children(
		chadcn.Button("Row actions", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
			ID("logs-row-actions").
			Attr("aria-haspopup", "menu").
			Attr("aria-expanded", fmt.Sprintf("%t", a.contextOpen)).
			OnClick(func(ctx context.Context) error {
				const menuX, menuY = -150, 34
				a.contextOpen = !a.contextOpen
				// The menu is absolutely positioned inside this wrapper, so
				// anchor it below the trigger and open it to the left.
				a.contextX, a.contextY = menuX, menuY
				return a.refresh(ctx)
			}),
		menu,
	)
	return ui.Div().Class("app-row app-row--between").Children(
		ui.Div().Class("app-row").Children(stats...),
		ui.Div().Class("app-toolbar").Children(
			chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Log actions"},
				chadcn.Button("Export", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
					OnClick(func(ctx context.Context) error {
						return a.showToast(ctx, "Export queued", fmt.Sprintf("%d rows will land as CSV in your inbox.", len(logs)))
					}),
				chadcn.Button("Share", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
					OnClick(func(ctx context.Context) error {
						return a.showToast(ctx, "View link copied", "Anyone on the team can open this filtered stream.")
					}),
			),
			chadcn.Toggle("Live tail", a.liveTail, chadcn.ToggleOptions{Size: chadcn.ButtonSmall}).
				OnClick(func(ctx context.Context) error {
					a.liveTail = !a.liveTail
					if a.liveTail {
						return a.showToast(ctx, "Live tail on", "New entries appear as they are written.")
					}
					return a.refresh(ctx)
				}),
			rowActions,
			chadcn.Pagination(chadcn.PaginationOptions{
				Page:       a.pageIndex + 1,
				TotalPages: a.logPageCount(),
				OnSelect: func(ctx context.Context, page int) error {
					a.pageIndex = page - 1
					a.selectedRows = map[int]bool{}
					return a.refresh(ctx)
				},
			}),
		),
	)
}

// logsTable renders the page rows, a skeleton while loading, or an empty state
// when the filters match nothing.
func (a *app) logsTable(rows []logEntry) *ui.Element {
	if a.loading {
		placeholder := ui.Div().Class("app-stack").Children(
			ui.Div().Class("app-row").Children(
				chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerSmall, Label: "Loading logs"}),
				ui.Text("Loading the last 24 hours…").Class("app-muted"),
			),
		)
		for range 6 {
			placeholder.Children(chadcn.Skeleton(chadcn.SkeletonOptions{Height: logRowPlaceholderHeight, Rounded: true}))
		}
		return placeholder
	}
	if len(rows) == 0 {
		return chadcn.Empty(chadcn.EmptyOptions{
			Icon:        ui.Text("≡"),
			Title:       "No logs match this view",
			Description: "Clear the search or widen the time range to bring rows back.",
			Actions: []*ui.Element{
				chadcn.Button("Reset filters", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
					OnClick(func(ctx context.Context) error {
						a.query = ""
						a.service = "All services"
						a.severity = nil
						a.pageIndex = 0
						a.selectedRows = map[int]bool{}
						return a.refresh(ctx)
					}),
			},
		})
	}
	tableRows := make([][]string, 0, len(rows))
	for _, entry := range rows {
		tableRows = append(tableRows, []string{
			entry.Time,
			strings.ToUpper(entry.Level),
			entry.Service,
			entry.Message,
			entry.Duration,
			statusLabel(entry.Status),
		})
	}

	return ui.Div().Class("app-table-wrap").Children(
		chadcn.DataTable(chadcn.DataTableOptions{
			ID:       "logs-table",
			Caption:  "Request log entries",
			Selected: a.selectedRows,
			OnSelect: func(ctx context.Context, row int) error {
				if a.selectedRows == nil {
					a.selectedRows = map[int]bool{}
				}
				if a.selectedRows[row] {
					delete(a.selectedRows, row)
				} else {
					a.selectedRows[row] = true
				}
				return a.refresh(ctx)
			},
			Columns: []chadcn.DataColumn{
				{Label: "Time", Width: ui.Px(104)},
				{Label: "Level", Width: ui.Px(84)},
				{Label: "Service", Width: ui.Px(132)},
				{Label: "Message"},
				{Label: "Duration", Align: "right", Width: ui.Px(88)},
				{Label: "Status", Align: "right", Width: ui.Px(76)},
			},
			Rows:  tableRows,
			Empty: "No log entries on this page.",
		}),
	)
}

// logsDetailDrawer shows the payload, export tile, and trace for the row the
// menu opened. The backdrop and Escape path both clear drawerOpen.
func (a *app) logsDetailDrawer() *ui.Element {
	entry, ok := a.selectedLog()
	title := "Log detail"
	description := "No row is selected."
	if ok {
		title = "Log detail · " + entry.Trace
		description = entry.Service + " · " + entry.Time + " · status " + statusLabel(entry.Status)
	}

	body := ui.Div().Class("app-drawer-body app-stack")
	if ok {
		body.Children(a.logsDetailBody(entry))
	} else {
		body.Children(chadcn.Empty(chadcn.EmptyOptions{
			Title:       "No entry selected",
			Description: "Open a row from the table or the row menu.",
		}))
	}
	return chadcn.Drawer(chadcn.DrawerOptions{
		ID:          "logs-detail-drawer",
		Open:        a.drawerOpen,
		Title:       title,
		Description: description,
		OnOpenChange: func(ctx context.Context, open bool) error {
			a.drawerOpen = open
			return a.refresh(ctx)
		},
		Footer: []*ui.Element{
			chadcn.Button("Close", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).
				OnClick(func(ctx context.Context) error {
					a.drawerOpen = false
					return a.refresh(ctx)
				}),
		},
	}, body)
}

// logsDetailBody is the drawer payload for one entry.
func (a *app) logsDetailBody(entry logEntry) *ui.Element {
	payload := fmt.Sprintf(`{
  "timestamp": "%s",
  "level": "%s",
  "service": "%s",
  "message": %q,
  "trace_id": "%s",
  "duration": "%s",
  "status": %d,
  "region": "eu-central-1"
}`, entry.Time, entry.Level, entry.Service, entry.Message, entry.Trace, entry.Duration, entry.Status)
	spans := make([]chadcn.TraceRow, 0, len(sampleSpans))
	for _, span := range sampleSpans {
		spans = append(spans, chadcn.TraceRow{
			Label:  span.Name,
			Detail: span.Service + " · " + span.Detail + " · " + span.Duration,
			Start:  span.Start,
			Width:  span.Width,
			Status: span.Status,
		})
	}
	return ui.Div().Class("app-stack").Children(
		ui.Div().Class("app-row").Children(
			chadcn.Badge(strings.ToUpper(entry.Level), chadcn.BadgeOptions{Variant: logLevelVariant(entry.Level)}),
			chadcn.Badge(statusLabel(entry.Status), chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
			ui.Text(entry.Duration+" · "+entry.Service).Class("app-muted"),
		),
		ui.Text(entry.Message).Class("app-log-message"),
		chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorHorizontal}),
		chadcn.CodeBlock(chadcn.CodeBlockOptions{
			Filename: entry.Trace + ".json",
			Language: "json",
			Code:     payload,
		}),
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("Trace timeline"),
				chadcn.CardDescription("Five spans from the gateway through to the response write."),
			),
			chadcn.CardContent(chadcn.Trace(chadcn.TraceOptions{Rows: spans})),
		),
		chadcn.Attachment(chadcn.AttachmentOptions{
			Name: entry.Trace + "-trace.json",
			Meta: "Trace export · 5 spans · 1.24s",
		}),
		ui.Text("Retained for 30 days. Tokens and passwords are redacted in exports.").Class("app-muted"),
	)
}
