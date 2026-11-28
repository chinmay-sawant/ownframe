package main

import (
	"context"
	"fmt"
	"html"
	"strconv"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	windowWidth       = 960
	windowHeight      = 640
	initialOpenCount  = 3
	initialDoneCount  = 1
	initialTotalCount = 4
	initialUptimePct  = 99.95
	bannerPadding     = 10
	bannerRadius      = 8
	bannerFontSize    = 12
)

// App holds the compiled view state and the ownframe page.
type App struct {
	page    *ownframe.Page
	view    *DashboardView
	nextSeq int
}

// New builds the interactive codegen dashboard page.
func New(devtools, perf bool) (*App, error) {
	bannerEl := ui.Div().
		BackgroundColor(ui.Hex("#172554")).
		BorderRadius(ui.Px(bannerRadius)).
		Padding(ui.Px(bannerPadding)).
		Color(ui.Hex("#bfdbfe")).
		FontSize(ui.Px(bannerFontSize)).
		Text("Embedded ui.Element slot rendered via {{raw .BannerHTML}}")

	bannerHTML, err := ui.Render(bannerEl)
	if err != nil {
		return nil, fmt.Errorf("render banner slot: %w", err)
	}

	view := NewDashboardView().
		WithTitle("OwnFrame Codegen Dashboard").
		WithOnline(true).
		WithBannerHTML(bannerHTML).
		WithOpenCount(initialOpenCount).
		WithDoneCount(initialDoneCount).
		WithTotalCount(initialTotalCount).
		WithUptimePct(initialUptimePct).
		WithTasks(defaultTasks()...)

	p, err := ownframe.New(ownframe.Config{
		Title:       "OwnFrame — Codegen Dashboard",
		Render:      view.Render,
		Width:       windowWidth,
		Height:      windowHeight,
		RetainState: true,
		DevTools:    devtools,
		Perf:        perf,
	})
	if err != nil {
		return nil, err
	}

	a := &App{
		page:    p,
		view:    view,
		nextSeq: initialTotalCount + 1,
	}

	p.Handle(ownframe.Handlers{
		ClickPaint: a.handleClick,
	})

	return a, nil
}

// Page returns the underlying ownframe page.
func (a *App) Page() *ownframe.Page {
	return a.page
}

func defaultTasks() []DashboardViewTask {
	tasks := []DashboardViewTask{
		{Name: "Compile view.html with ownframegen", Owner: "build", Done: true},
		{Name: "Wire Config.Render without html/template", Owner: "render", Done: false},
		{Name: "Batch updates through PatchBuilder", Owner: "dom", Done: false},
		{Name: "Benchmark against html/template", Owner: "perf", Done: false},
	}
	for i := range tasks {
		tasks[i].ID = i + 1
	}

	return tasks
}

func (a *App) handleClick(ctx context.Context, box ownframe.Box) (bool, error) {
	if box.Action == "" {
		return false, nil
	}
	if err := a.handleAction(ctx, box.Action); err != nil {
		return false, err
	}

	return true, nil
}

func (a *App) handleAction(ctx context.Context, action string) error {
	switch action {
	case ActionToggleStatus:
		a.view.Online = !a.view.Online
		label := "Online"
		if !a.view.Online {
			label = "Offline"
		}

		return a.view.Patch().
			StatusBadgeText(label).
			StatusBadgeToggleClass("offline").
			LastEventText("Status switched to "+label).
			Apply(ctx, a.page)

	case ActionIncrementDone:
		if a.view.OpenCount > 0 {
			a.view.OpenCount--
		}
		a.view.DoneCount++

		return a.view.Patch().
			OpenCountText(strconv.Itoa(a.view.OpenCount)).
			DoneCountText(strconv.Itoa(a.view.DoneCount)).
			LastEventText("Completed task #"+strconv.Itoa(a.view.DoneCount)).
			Apply(ctx, a.page)

	case ActionAppendTask:
		id := a.nextSeq
		a.nextSeq++
		a.view.OpenCount++
		a.view.TotalCount++
		name := fmt.Sprintf("Generated task #%d", id)
		frag := fmt.Sprintf(
			`<div id="task-%d" class="task-row"><span class="task-id">#%d</span><span class="task-title">%s</span><span class="task-owner">ops</span></div>`,
			id, id, html.EscapeString(name),
		)

		return a.view.Patch().
			TaskListAppendHTML(frag).
			OpenCountText(strconv.Itoa(a.view.OpenCount)).
			TotalCountText(strconv.Itoa(a.view.TotalCount)).
			LastEventText("Appended "+name).
			Apply(ctx, a.page)

	case ActionResetTasks:
		a.view.OpenCount = initialOpenCount
		a.view.DoneCount = initialDoneCount
		a.view.TotalCount = initialTotalCount
		a.nextSeq = initialTotalCount + 1
		var rows string
		for _, task := range defaultTasks() {
			cls := "task-row"
			if task.Done {
				cls = "task-row done"
			}
			rows += fmt.Sprintf(
				`<div id="task-%d" class="%s"><span class="task-id">#%d</span><span class="task-title">%s</span><span class="task-owner">%s</span></div>`,
				task.ID, cls, task.ID, html.EscapeString(task.Name), html.EscapeString(task.Owner),
			)
		}

		return a.view.Patch().
			TaskListInnerHTML(rows).
			OpenCountText(strconv.Itoa(a.view.OpenCount)).
			DoneCountText(strconv.Itoa(a.view.DoneCount)).
			TotalCountText(strconv.Itoa(a.view.TotalCount)).
			LastEventText("Reset task list").
			Apply(ctx, a.page)

	default:
		return nil
	}
}
