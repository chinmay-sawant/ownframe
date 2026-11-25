package page_test

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
)

const scaleListHTML = `<!DOCTYPE html><html><head><style>` +
	`html,body{margin:0;font-family:sans-serif;color:#1c1915;background:#f4f1ea}` +
	`#head{height:40px;background:#1c1915;color:#ffffff}` +
	`.row{display:flex;box-sizing:border-box;height:40px;gap:12px;padding:8px 16px;border-bottom:1px solid #ddd6c7}` +
	`.num{width:64px}.title{width:200px;font-weight:bold}.body{color:#4a463c}` +
	`</style></head><body><div id="head">Scale benchmark</div>` +
	`{{range .}}<div id="row-{{.Num}}" class="row">` +
	`<span class="num">{{.Num}}</span>` +
	`<span class="title">{{.Title}}</span>` +
	`<span class="body">{{.Body}}</span>` +
	`</div>{{end}}</body></html>`

type scaleRow struct {
	Num   int
	Title string
	Body  string
}

func makeScaleRows(n int) []scaleRow {
	rows := make([]scaleRow, n)
	for i := range rows {
		idx := i + 1
		rows[i] = scaleRow{
			Num:   idx,
			Title: "Item " + strconv.Itoa(idx),
			Body:  "Row payload text for item " + strconv.Itoa(idx),
		}
	}

	return rows
}

// BenchmarkScaleStages measures each pipeline stage (TemplateExecute,
// HTMLParse, CSSApply, CSSRelayout, DisplayList) across 50, 200, and 1,000 rows
// to expose per-stage cost ratios and scaling linearity (ns/row).
func BenchmarkScaleStages(b *testing.B) {
	ctx := context.Background()
	tpl := template.Must(template.New("scale").Parse(scaleListHTML))
	const width, height = 800, 600

	for _, count := range []int{50, 200, 1000} {
		rows := makeScaleRows(count)

		var initial bytes.Buffer
		if err := tpl.Execute(&initial, rows); err != nil {
			b.Fatal(err)
		}
		source := initial.Bytes()

		b.Run(fmt.Sprintf("Rows%d", count), func(b *testing.B) {
			b.Run("TemplateExecute", func(b *testing.B) {
				b.ReportAllocs()
				var buf bytes.Buffer
				for b.Loop() {
					buf.Reset()
					if err := tpl.Execute(&buf, rows); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*count), "ns/row")
			})

			b.Run("HTMLParse", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := html.Parse(source); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*count), "ns/row")
			})

			doc, err := html.Parse(source)
			if err != nil {
				b.Fatal(err)
			}

			b.Run("CSSApply", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := css.Apply(ctx, doc, css.Options{
						WidthPx:  width,
						HeightPx: height,
						Media:    "screen",
					}); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*count), "ns/row")
			})

			styled, err := css.Apply(ctx, doc, css.Options{
				WidthPx:  width,
				HeightPx: height,
				Media:    "screen",
			})
			if err != nil {
				b.Fatal(err)
			}

			b.Run("CSSRelayout", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; b.Loop(); i++ {
					w := width + i%2
					if _, err := css.Relayout(ctx, styled, w, height, "", "", ""); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*count), "ns/row")
			})

			b.Run("DisplayList", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := layout.DisplayListOptions(ctx, styled, layout.Options{}); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*count), "ns/row")
			})
		})
	}
}

// BenchmarkLayoutModels compares Block, Flex, and Grid DisplayList costs on the
// same 200-row 3-column dataset to isolate layout-model overhead.
func BenchmarkLayoutModels(b *testing.B) {
	ctx := context.Background()
	const rowCount = 200
	const width, height = 800, 600
	rows := makeScaleRows(rowCount)

	models := []struct {
		name string
		css  string
	}{
		{
			name: "Block",
			css: `.row{display:block;box-sizing:border-box;height:40px;padding:8px 16px;border-bottom:1px solid #ddd6c7}` +
				`.num,.title,.body{display:inline-block}.num{width:64px}.title{width:200px;font-weight:bold}.body{color:#4a463c}`,
		},
		{
			name: "Flex",
			css: `.row{display:flex;box-sizing:border-box;height:40px;gap:12px;padding:8px 16px;border-bottom:1px solid #ddd6c7}` +
				`.num{width:64px}.title{width:200px;font-weight:bold}.body{color:#4a463c}`,
		},
		{
			name: "Grid",
			css: `.row{display:grid;grid-template-columns:64px 200px 1fr;box-sizing:border-box;height:40px;gap:12px;padding:8px 16px;border-bottom:1px solid #ddd6c7}` +
				`.title{font-weight:bold}.body{color:#4a463c}`,
		},
	}

	for _, m := range models {
		src := `<!DOCTYPE html><html><head><style>` +
			`html,body{margin:0;font-family:sans-serif;color:#1c1915;background:#f4f1ea}` +
			m.css +
			`</style></head><body>{{range .}}<div class="row">` +
			`<span class="num">{{.Num}}</span>` +
			`<span class="title">{{.Title}}</span>` +
			`<span class="body">{{.Body}}</span>` +
			`</div>{{end}}</body></html>`
		tpl := template.Must(template.New(m.name).Parse(src))
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, rows); err != nil {
			b.Fatal(err)
		}
		doc, err := html.Parse(buf.Bytes())
		if err != nil {
			b.Fatal(err)
		}
		styled, err := css.Apply(ctx, doc, css.Options{WidthPx: width, HeightPx: height, Media: "screen"})
		if err != nil {
			b.Fatal(err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := layout.DisplayListOptions(ctx, styled, layout.Options{}); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*rowCount), "ns/row")
		})
	}
}

// BenchmarkScaleStagesDenseGrid measures HTMLParse, CSSApply, and DisplayList
// on a 450-cell positioned grid under both a minimal stylesheet (~10 rules) and
// a realistic theme-sized stylesheet (~300 rules with :root custom properties),
// exposing how stylesheet size scales style resolution per element.
func BenchmarkScaleStagesDenseGrid(b *testing.B) {
	ctx := context.Background()
	const rows, cols = 30, 15
	const totalCells = rows * cols
	const width, height = 1440, 960

	var heavyCSS bytes.Buffer
	heavyCSS.WriteString(":root{")
	for i := range 60 {
		fmt.Fprintf(&heavyCSS, "--tok-%d:#%02x%02x%02x;", i, (i*37)%255, (i*73)%255, (i*113)%255)
	}
	heavyCSS.WriteString("}")
	for i := range 280 {
		fmt.Fprintf(&heavyCSS, ".comp-%d{display:flex;padding:4px 8px;color:var(--tok-%d);background:var(--tok-%d);border:1px solid var(--tok-%d)}", i, i%60, (i+1)%60, (i+2)%60)
	}

	baseCSS := `html,body{margin:0;font-family:sans-serif;font-size:13px}` +
		`.grid{position:relative;width:1400px;height:800px;overflow:hidden}` +
		`.cell{position:absolute;box-sizing:border-box;overflow:hidden;white-space:nowrap;padding:0 4px;line-height:22px;color:var(--tok-0,#1f2328);border-right:1px solid var(--tok-1,#d4d4d4);border-bottom:1px solid var(--tok-1,#d4d4d4)}`

	var body bytes.Buffer
	body.WriteString(`<div class="grid">`)
	for r := range rows {
		for c := range cols {
			fmt.Fprintf(&body, `<div id="c-%d-%d" class="cell" style="left:%dpx;top:%dpx;width:88px;height:24px"><span>%d</span></div>`, r, c, c*88, r*24, r*cols+c)
		}
	}
	body.WriteString(`</div>`)

	variants := []struct {
		name string
		css  string
	}{
		{"MinimalCSS", baseCSS},
		{"HeavyThemeCSS", heavyCSS.String() + baseCSS},
	}

	for _, v := range variants {
		src := []byte(`<!DOCTYPE html><html><head><style>` + v.css + `</style></head><body>` + body.String() + `</body></html>`)
		b.Run(v.name, func(b *testing.B) {
			b.Run("HTMLParse", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := html.Parse(src); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*totalCells), "ns/cell")
			})

			doc, err := html.Parse(src)
			if err != nil {
				b.Fatal(err)
			}

			b.Run("CSSApply", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := css.Apply(ctx, doc, css.Options{WidthPx: width, HeightPx: height, Media: "screen"}); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*totalCells), "ns/cell")
			})

			styled, err := css.Apply(ctx, doc, css.Options{WidthPx: width, HeightPx: height, Media: "screen"})
			if err != nil {
				b.Fatal(err)
			}

			b.Run("DisplayList", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := layout.DisplayListOptions(ctx, styled, layout.Options{RetainState: true}); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*totalCells), "ns/cell")
			})
		})
	}
}
