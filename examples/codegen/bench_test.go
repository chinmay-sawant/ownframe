package main

import (
	"bytes"
	"context"
	"html/template"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

const benchTemplateHTML = `<!DOCTYPE html>
<html>
<head>
<style>
html, body { margin: 0; padding: 0; background: #0f172a; color: #e2e8f0; font-family: sans-serif; }
.shell { padding: 16px; }
.header { display: flex; justify-content: space-between; align-items: center; height: 48px; padding: 0 12px; background: #1e293b; margin-bottom: 10px; }
.status-badge { width: 120px; height: 26px; overflow: hidden; white-space: nowrap; }
.metrics { display: flex; gap: 10px; margin-bottom: 10px; }
.metric-val { width: 80px; height: 22px; overflow: hidden; white-space: nowrap; }
.task-row { display: flex; align-items: center; box-sizing: border-box; height: 32px; overflow: hidden; padding: 0 8px; border-bottom: 1px solid #334155; }
.task-id { width: 56px; overflow: hidden; white-space: nowrap; }
.task-title { flex: 1; overflow: hidden; white-space: nowrap; }
.task-owner { width: 100px; overflow: hidden; white-space: nowrap; }
</style>
</head>
<body>
<div class="shell">
  <div class="header">
    <div>{{.Title}}</div>
    <div id="status-badge" class="status-badge{{if not .Online}} offline{{end}}">{{if .Online}}Online{{else}}Offline{{end}}</div>
  </div>
  <div class="metrics">
    <div id="open-count" class="metric-val">{{.OpenCount}}</div>
    <div id="done-count" class="metric-val">{{.DoneCount}}</div>
    <div id="total-count" class="metric-val">{{.TotalCount}}</div>
    <div id="uptime-pct" class="metric-val">{{.UptimePct}}</div>
  </div>
  <div id="last-event">Ready</div>
  <div id="task-list">
    {{- range .Tasks -}}
    <div id="task-{{.ID}}" class="task-row{{if .Done}} done{{end}}">
      <span class="task-id">#{{.ID}}</span>
      <span class="task-title">{{.Name}}</span>
      <span class="task-owner">{{.Owner}}</span>
    </div>
    {{- end -}}
  </div>
</div>
</body>
</html>`

func makeBenchView(rows int) *DashboardView {
	tasks := make([]DashboardViewTask, rows)
	for i := range tasks {
		id := i + 1
		s := strconv.Itoa(id)
		tasks[i] = DashboardViewTask{
			ID:    id,
			Name:  "Task item #" + s + " pipeline verification",
			Owner: "team-" + strconv.Itoa(id%8),
			Done:  id%3 == 0,
		}
	}

	return NewDashboardView().
		WithTitle("Benchmark Dashboard").
		WithOnline(true).
		WithOpenCount(rows - rows/3).
		WithDoneCount(rows / 3).
		WithTotalCount(rows).
		WithUptimePct(99.95).
		WithTasks(tasks...)
}

func benchmarkHTMLGenTemplate(b *testing.B, rows int) {
	tmpl := template.Must(template.New("bench").Parse(benchTemplateHTML))
	data := makeBenchView(rows)
	var buf bytes.Buffer

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkHTMLGenCodegen(b *testing.B, rows int) {
	view := makeBenchView(rows)
	var sb strings.Builder

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sb.Reset()
		view.RenderTo(&sb)
	}
}

func BenchmarkHTMLGen_Template_50(b *testing.B)   { benchmarkHTMLGenTemplate(b, 50) }
func BenchmarkHTMLGen_Codegen_50(b *testing.B)    { benchmarkHTMLGenCodegen(b, 50) }
func BenchmarkHTMLGen_Template_200(b *testing.B)  { benchmarkHTMLGenTemplate(b, 200) }
func BenchmarkHTMLGen_Codegen_200(b *testing.B)   { benchmarkHTMLGenCodegen(b, 200) }
func BenchmarkHTMLGen_Template_1000(b *testing.B) { benchmarkHTMLGenTemplate(b, 1000) }
func BenchmarkHTMLGen_Codegen_1000(b *testing.B)  { benchmarkHTMLGenCodegen(b, 1000) }

func BenchmarkRedraw_HTMLTemplate_200(b *testing.B) {
	ctx := context.Background()
	data := makeBenchView(200)
	p, err := ownframe.New(ownframe.Config{
		Title:  "bench-tmpl",
		HTML:   benchTemplateHTML,
		Width:  windowWidth,
		Height: windowHeight,
	})
	if err != nil {
		b.Fatal(err)
	}
	p.SetData(data)
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		data.DoneCount++
		p.SetData(data)
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRedraw_CodegenRender_200(b *testing.B) {
	ctx := context.Background()
	view := makeBenchView(200)
	p, err := ownframe.New(ownframe.Config{
		Title:  "bench-codegen",
		Render: view.Render,
		Width:  windowWidth,
		Height: windowHeight,
	})
	if err != nil {
		b.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		view.DoneCount++
		p.SetData(view)
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPatch_CodegenBuilder_1Row_200(b *testing.B) {
	ctx := context.Background()
	view := makeBenchView(200)
	p, err := ownframe.New(ownframe.Config{
		Title:       "bench-patch-1",
		Render:      view.Render,
		Width:       windowWidth,
		Height:      windowHeight,
		RetainState: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		view.DoneCount++
		err := view.Patch().
			DoneCountText(strconv.Itoa(view.DoneCount)).
			Apply(ctx, p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPatch_CodegenBuilder_10Rows_200(b *testing.B) {
	ctx := context.Background()
	view := makeBenchView(200)
	p, err := ownframe.New(ownframe.Config{
		Title:       "bench-patch-10",
		Render:      view.Render,
		Width:       windowWidth,
		Height:      windowHeight,
		RetainState: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		view.DoneCount++
		n := strconv.Itoa(view.DoneCount)
		pb := view.Patch().
			OpenCountText(n).
			DoneCountText(n).
			TotalCountText(n).
			UptimePctText("99.9" + n[len(n)-1:])
		for i := 1; i <= 6; i++ {
			txt := "Updated #" + strconv.Itoa(i) + " (" + n + ")"
			pb.ByID("task-"+strconv.Itoa(i), ownframe.ElementEdit{Text: &txt})
		}
		if err := pb.Apply(ctx, p); err != nil {
			b.Fatal(err)
		}
	}
}
