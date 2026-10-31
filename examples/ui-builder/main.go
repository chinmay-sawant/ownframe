// Command ui-builder showcases interactive OwnFrame components and an
// illustrative system-monitor dashboard. Values are local fixtures, not
// measurements collected from the user's device.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	demoWidth  = 1120
	demoHeight = 760
)

var (
	cpuSamples    = []int{31, 44, 38, 58, 67, 49, 73, 62, 54, 71, 65, 47}
	memorySamples = []int{51, 53, 56, 54, 59, 61, 58, 63, 65, 62, 66, 64}
)

type monitorApp struct {
	page     *ownframe.Page
	tab      string
	file     string
	metric   string
	sample   int
	notice   string
}

func newMonitorApp() (*monitorApp, error) {
	a := &monitorApp{
		tab: "overview", file: "main.go", metric: "CPU",
		notice: "Choose a metric, chart, section, or file to explore the builder.",
	}
	p, err := ui.NewPage(ownframe.Config{
		Title: "OwnFrame · Builder Dashboard",
		Width: demoWidth, Height: demoHeight,
		MinWidth: 900, MinHeight: 610,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.page = p
	// Stateful demo actions use the same public data-action click route as
	// normal OwnFrame apps. A re-render uses the public Page.Load API.
	p.Handle(ownframe.Handlers{Click: a.handleClick})
	return a, nil
}

func (a *monitorApp) handleClick(ctx context.Context, box ownframe.Box) error {
	switch box.Action {
	case "view:overview":
		a.tab = "overview"
		a.notice = "Overview: sample CPU, memory, and disk metrics."
	case "view:processes":
		a.tab = "processes"
		a.notice = "Processes: illustrative application rows."
	case "view:charts":
		a.tab = "charts"
		a.notice = "Charts: twelve synthetic data points per series."
	case "file:main.go", "file:go.mod", "file:README.md":
		a.file = box.Action[len("file:"):]
		a.notice = "Selected " + a.file + " in the example explorer."
	case "metric:cpu":
		a.metric = "CPU"
		a.notice = "CPU graph selected (demo data)."
	case "metric:memory":
		a.metric = "Memory"
		a.notice = "Memory graph selected (demo data)."
	case "metric:disk":
		a.metric = "Disk"
		a.notice = "Disk metric selected (demo data)."
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
	fragment, err := ui.Render(a.screen())
	if err != nil {
		return err
	}
	return a.page.Load(ctx, "<!doctype html><html><head><meta charset=\"utf-8\"></head><body>"+
		fragment+"</body></html>")
}

func main() {
	web := flag.Bool("web", false, "serve the preview in a browser")
	addr := flag.String("addr", "127.0.0.1:8131", "address for -web")
	flag.Parse()

	app, err := newMonitorApp()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := app.page.Redraw(ctx); err != nil {
		log.Fatal(err)
	}
	if *web {
		if err := ownframe.Serve(ctx, app.page, *addr); err != nil {
			log.Fatal(err)
		}
		return
	}
	// Native drag is handled by OwnFrame, not by injected JS or Chromium.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID: "ui-builder",
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < 68
		},
	}
	if err := ownframe.RunWithOptions(ctx, app.page, options); err != nil {
		log.Fatal(err)
	}
}

func (a *monitorApp) screen() *ui.Element {
	var center *ui.Element
	switch a.tab {
	case "processes":
		center = a.processesPanel()
	case "charts":
		center = a.chartsPanel()
	default:
		center = a.overviewPanel()
	}
	return ui.Div().Width("100%").Height("740px").
		Display("flex").FlexDirection("column").
		BackgroundColor("#111827").Color("#e9f0fa").
		CSS("font-family", "sans-serif").FontSize("14px").
		Children(
			a.titlebar(),
			ui.Div().Display("flex").CSS("flex", "1").
				CSS("min-height", "0px").Gap("16px").Padding("16px").
				Children(a.sidebar(), center),
			ui.Div().Padding("9px 18px").BackgroundColor("#182336").
				Color("#a4b5ca").FontSize("12px").
				Children(ui.Text(a.notice)),
		)
}

func (a *monitorApp) titlebar() *ui.Element {
	return ui.Header().Height("68px").Display("flex").
		AlignItems("center").CSS("justify-content", "space-between").
		Padding("0 18px").BackgroundColor("#202d42").
		CSS("border-bottom", "1px solid #34445c").
		Children(
			ui.Div().Display("flex").AlignItems("center").Gap("10px").
				Children(
					ui.Div().Padding("8px 11px").BorderRadius("9px").
						BackgroundColor("#4677ed").FontWeight("bold").
						Children(ui.Text("O")),
					ui.Div().Children(
						ui.Text("OwnFrame Builder").FontWeight("bold").FontSize("17px"),
						ui.P().Text("Native Go UI · No Chromium").Margin("3px 0 0").
							Color("#a3b5cc").FontSize("11px"),
					),
				),
			ui.Div().Display("flex").AlignItems("center").Gap("8px").
				Children(
					a.action("Save", "save", true),
					a.action("Cancel", "cancel", false),
					a.action("↻ Refresh", "refresh", false),
					a.action("−", "win-min", false),
					a.action("□", "win-max", false),
					a.action("×", "win-close", false),
				),
		)
}

func (a *monitorApp) action(label, command string, primary bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	if primary {
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", command).
		BackgroundColor(bg).Color(fg).CSS("border", border).
		Padding("9px 12px").BorderRadius("8px").
		FontSize("12px").FontWeight("bold")
}

func (a *monitorApp) sidebar() *ui.Element {
	return ui.Nav().Width("205px").CSS("flex-shrink", "0").
		BackgroundColor("#202d42").BorderRadius("11px").
		Padding("16px").Display("flex").FlexDirection("column").
		Gap("9px").Children(
			ui.Text("WORKSPACE").FontWeight("bold").FontSize("11px").Color("#9aacbf"),
			a.navButton("▦ Overview", "overview"),
			a.navButton("▥ Charts", "charts"),
			a.navButton("≡ Processes", "processes"),
			ui.P().Text("EXPLORER").Margin("24px 0 4px").
				FontSize("11px").FontWeight("bold").Color("#9aacbf"),
			a.fileButton("main.go"),
			a.fileButton("go.mod"),
			a.fileButton("README.md"),
			ui.Div().Margin("25px 0 0").Padding("12px").
				BackgroundColor("#2b3950").BorderRadius("8px").
				Children(
					ui.Text("Interactive builder").FontWeight("bold"),
					ui.P().Text("Every menu item, file, metric and action is wired to Go.").
						Margin("7px 0 0").FontSize("11px").Color("#b2c1d5"),
				),
		)
}

func (a *monitorApp) navButton(label, tab string) *ui.Element {
	bg := "#2b3950"
	if a.tab == tab {
		bg = "#355596"
	}
	return ui.Button(label).Attr("data-action", "view:"+tab).
		Width("100%").CSS("text-align", "left").
		BackgroundColor(bg).Color("#f1f5fd").
		CSS("border", "1px solid #40546f").
		Padding("11px").BorderRadius("8px").FontWeight("bold")
}

func (a *monitorApp) fileButton(file string) *ui.Element {
	bg := "#202d42"
	if a.file == file {
		bg = "#34445e"
	}
	return ui.Button("▸ "+file).Attr("data-action", "file:"+file).
		Width("100%").CSS("text-align", "left").
		BackgroundColor(bg).Color("#cbd8eb").
		CSS("border", "1px solid #3a4b62").
		Padding("9px 11px").BorderRadius("7px")
}

func (a *monitorApp) panelHeader(title, subtitle string) *ui.Element {
	return ui.Div().Children(
		ui.Text(title).FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
		ui.P().Text(subtitle).Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
	)
}

func (a *monitorApp) overviewPanel() *ui.Element {
	cpu := cpuSamples[a.sample%len(cpuSamples)]
	mem := memorySamples[a.sample%len(memorySamples)]
	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").
		Display("flex").FlexDirection("column").Gap("14px").
		Children(
			a.panelHeader("System overview", "Illustrative readings only — not live system telemetry."),
			ui.Div().Display("flex").Gap("12px").Children(
				a.metricCard("CPU utilization", fmt.Sprintf("%d%%", cpu), "12 logical cores", "metric:cpu"),
				a.metricCard("System memory", fmt.Sprintf("%d%%", mem), "10.2 of 16 GB · sample", "metric:memory"),
				a.metricCard("Disk usage", "68%", "340 of 500 GB · sample", "metric:disk"),
			),
			a.chartRow(),
			ui.Div().Padding("12px 15px").BackgroundColor("#202d42").
				BorderRadius("10px").Children(
				ui.Text("Selected metric: "+a.metric).FontWeight("bold"),
				ui.P().Text("Charts and values are deterministic fixtures. Refresh cycles sample values; click a card to change the selection.").
					Margin("8px 0 0").FontSize("12px").Color("#abbcd2"),
			),
		)
}

func (a *monitorApp) metricCard(label, value, note, action string) *ui.Element {
	bg := "#202d42"
	if (a.metric == "CPU" && action == "metric:cpu") ||
		(a.metric == "Memory" && action == "metric:memory") ||
		(a.metric == "Disk" && action == "metric:disk") {
		bg = "#2b4164"
	}
	return ui.Button("").Attr("data-action", action).
		CSS("flex", "1").CSS("text-align", "left").
		BackgroundColor(bg).Color("#ecf2fb").
		CSS("border", "1px solid #3e526e").
		Padding("16px").BorderRadius("11px").
		Children(
			ui.Text(label).FontSize("13px").Color("#b5c7df"),
			ui.P().Text(value).Margin("10px 0 5px").
				FontSize("29px").FontWeight("bold").Color("#f4f7ff"),
			ui.Text(note).FontSize("11px").Color("#a0b6d0"),
		)
}

func (a *monitorApp) chartRow() *ui.Element {
	return ui.Div().Display("flex").Gap("12px").Children(
		a.chartCard("CPU history", "metric:cpu", cpuSamples, "#6d9bff"),
		a.chartCard("Memory history", "metric:memory", memorySamples, "#59d4b8"),
	)
}

func (a *monitorApp) chartCard(label, action string, samples []int, color string) *ui.Element {
	bars := make([]*ui.Element, 0, len(samples))
	for i := range samples {
		value := samples[(i+a.sample)%len(samples)]
		bars = append(bars, ui.Div().Width("15px").
			Height(fmt.Sprintf("%dpx", 18+value)).
			BackgroundColor(color).BorderRadius("4px"))
	}
	return ui.Div().CSS("flex", "1").CSS("min-width", "0px").
		Padding("16px").BorderRadius("11px").BackgroundColor("#202d42").
		Children(
			ui.Button(label+"  ↗").Attr("data-action", action).
				BackgroundColor("#2d4261").Color("#e6efff").
				CSS("border", "1px solid #425b7d").
				Padding("8px 12px").BorderRadius("7px").FontWeight("bold"),
			ui.Div().Height("135px").Display("flex").Gap("8px").
				AlignItems("flex-end").Padding("10px 0 0").
				Children(bars...),
			ui.Text("12 simulated samples · click title to inspect").
				FontSize("11px").Color("#a8bad1"),
		)
}

func (a *monitorApp) chartsPanel() *ui.Element {
	return ui.Section().CSS("flex", "1").Display("flex").
		FlexDirection("column").Gap("18px").
		Children(
			a.panelHeader("Performance charts", "Click chart titles to select a metric; refresh rotates synthetic samples."),
			a.chartRow(),
			a.metricCard("Active selection", a.metric, "Metrics remain local, in-memory sample data.", "metric:cpu"),
		)
}

func (a *monitorApp) processesPanel() *ui.Element {
	processes := []struct {
		name string
		cpu  string
		mem  string
	}{
		{"ownframe-render", "12.1%", "98 MB"},
		{"blinkless-layout", "8.4%", "42 MB"},
		{"ui-event-loop", "2.6%", "17 MB"},
		{"demo-worker", "0.8%", "12 MB"},
	}
	rows := make([]*ui.Element, 0, len(processes))
	for _, process := range processes {
		rows = append(rows, ui.Div().Display("flex").
			CSS("justify-content", "space-between").Padding("16px").
			BackgroundColor("#26364e").BorderRadius("8px").
			Children(
				ui.Text(process.name).FontWeight("bold"),
				ui.Text(process.cpu+" CPU  ·  "+process.mem).Color("#b6c8df"),
			))
	}
	return ui.Section().CSS("flex", "1").Display("flex").
		FlexDirection("column").Gap("12px").
		Children(
			a.panelHeader("Example processes", "Synthetic rows demonstrating tables and layout in the Go builder."),
			ui.Div().Display("flex").FlexDirection("column").
				Gap("9px").Children(rows...),
		)
}
