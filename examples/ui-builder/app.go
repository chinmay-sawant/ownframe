package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	demoWidth         = 1120
	demoHeight        = 760
	demoMinWidth      = 900
	demoMinHeight     = 610
	builderDragLimitY = 68
	barBaseHeight     = 18
	builderPlacement  = "ui-builder"
)

type monitorApp struct {
	page    *ownframe.Page
	tab     string
	file    string
	metric  string
	sample  int
	notice  string
	cpuData []int
	memData []int
}

func newMonitorApp() (*monitorApp, error) {
	a := &monitorApp{
		tab:    "overview",
		file:   "main.go",
		metric: "CPU",
		notice: "Choose a metric, chart, section, or file to explore the builder.",
		cpuData: []int{
			31, 44, 38, 58, 67, 49, 73, 62, 54, 71, 65, 47,
		},
		memData: []int{
			51, 53, 56, 54, 59, 61, 58, 63, 65, 62, 66, 64,
		},
	}
	p, err := ui.NewPage(ownframe.Config{
		Title:     "OwnFrame · Builder Dashboard",
		Width:     demoWidth,
		Height:    demoHeight,
		MinWidth:  demoMinWidth,
		MinHeight: demoMinHeight,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new monitor app: %w", err)
	}
	a.page = p
	// Stateful demo actions use the same public data-action click route as
	// normal OwnFrame apps. A re-render uses the public Page.Load API.
	p.Handle(ownframe.Handlers{Click: a.handleClick})
	return a, nil
}

func (a *monitorApp) handleClick(ctx context.Context, box ownframe.Box) error {
	if name, ok := strings.CutPrefix(box.Action, "view:"); ok {
		a.selectView(name)
	} else if name, ok := strings.CutPrefix(box.Action, "file:"); ok {
		a.selectFile(name)
	} else if name, ok := strings.CutPrefix(box.Action, "metric:"); ok {
		a.selectMetric(name)
	} else {
		switch box.Action {
		case "refresh":
			a.sample++
			a.notice = "Sample refreshed. No device telemetry was read."
		case "save":
			a.notice = "Save clicked — no file was written (demo action)."
		case "cancel":
			a.tab, a.metric, a.file = "overview", "CPU", "main.go"
			a.notice = "Selection reset to the overview."
		case "win-min":
			a.page.Window().Minimize()
			return nil
		case "win-max":
			a.page.Window().ToggleMaximize()
			return nil
		case "win-close":
			a.page.Window().Close()
			return nil
		default:
			return nil
		}
	}
	return a.refresh(ctx)
}

func (a *monitorApp) selectView(name string) {
	switch name {
	case "overview", "processes", "charts":
		a.tab = name
		a.notice = map[string]string{
			"overview":  "Overview: sample CPU, memory, and disk metrics.",
			"processes": "Processes: illustrative application rows.",
			"charts":    "Charts: twelve synthetic data points per series.",
		}[name]
	}
}

func (a *monitorApp) selectFile(name string) {
	switch name {
	case "main.go", "go.mod", "README.md":
		a.file = name
		a.notice = "Selected " + a.file + " in the example explorer."
	}
}

func (a *monitorApp) selectMetric(name string) {
	switch name {
	case "cpu":
		a.metric = "CPU"
		a.notice = "CPU graph selected (demo data)."
	case "memory":
		a.metric = "Memory"
		a.notice = "Memory graph selected (demo data)."
	case "disk":
		a.metric = "Disk"
		a.notice = "Disk metric selected (demo data)."
	}
}

func (a *monitorApp) refresh(ctx context.Context) error {
	fragment, err := ui.Render(a.screen())
	if err != nil {
		return fmt.Errorf("render builder dashboard: %w", err)
	}
	if err := a.page.Load(ctx, renderBuilderDocument(fragment)); err != nil {
		return fmt.Errorf("reload builder dashboard: %w", err)
	}
	return nil
}

func renderBuilderDocument(fragment string) string {
	return "<!doctype html><html><head><meta charset=\"utf-8\"></head><body>" +
		fragment + "</body></html>"
}

func (a *monitorApp) cpuAt(offset int) int {
	return a.cpuData[offset%len(a.cpuData)]
}

func (a *monitorApp) memAt(offset int) int {
	return a.memData[offset%len(a.memData)]
}
