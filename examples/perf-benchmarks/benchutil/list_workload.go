package benchutil

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/list"
)

const (
	ListRows       = 1000
	ListRowHeight  = 40
	ListHeadHeight = 40
	listOverscan   = 4
)

// ListOptions selects instrumentation and row windowing for the same dataset.
type ListOptions struct {
	Perf     bool
	Windowed bool
}

// ListWorkload is a fixed-height list for comparable page-only workloads.
// Scroll reports the host offset; it does not run or measure the native host.
type ListWorkload struct {
	Page     *ownframe.Page
	rows     []Row
	view     listView
	windowed bool
	start    int
}

type listView struct {
	Rows        []Row
	Top, Bottom int
}

func NewListWorkload(options ListOptions) (*ListWorkload, error) {
	p, err := ownframe.New(ownframe.Config{
		Title: "List workload", HTML: listWorkloadHTML,
		Width: largeWidth, Height: largeHeight,
		Perf: options.Perf,
	})
	if err != nil {
		return nil, err
	}
	w := &ListWorkload{Page: p, rows: makeView(ListRows).Rows, windowed: options.Windowed, start: -1}
	w.view.Rows = w.rows
	p.SetData(&w.view)
	if options.Windowed {
		p.Scroll().SetWindowing(true)
		p.SetScrollWindow(w.slice)
	}
	return w, nil
}

// SetText changes a dataset row; Redraw publishes the changed template.
func (w *ListWorkload) SetText(index int, text string) error {
	if index < 0 || index >= len(w.rows) {
		return fmt.Errorf("list workload: row %d out of range", index)
	}
	w.rows[index].Body = text
	w.Page.SetData(&w.view)
	return nil
}

// Scroll submits an already-clamped host offset. Only a windowed page relayouts.
func (w *ListWorkload) Scroll(ctx context.Context, offset int) error {
	w.Page.Scroll().SetScrollOffset(0, offset)
	if w.windowed {
		return w.Page.Redraw(ctx)
	}
	return nil
}

func (w *ListWorkload) slice(offset, height int) bool {
	rows := list.Window(offset, height, ListHeadHeight, ListRowHeight, len(w.rows), listOverscan)
	if rows.Start == w.start && rows.Count == len(w.view.Rows) {
		return false
	}
	w.start = rows.Start
	w.view.Rows = w.rows[rows.Start : rows.Start+rows.Count]
	w.view.Top, w.view.Bottom = rows.Top, rows.Bottom
	w.Page.SetData(&w.view)
	return true
}

const listWorkloadHTML = `<!DOCTYPE html><html><head><style>
html,body{margin:0;font-family:sans-serif;color:#1c1915;background:#f4f1ea}
#head{height:40px;background:#1c1915;color:#fff}
.row{display:flex;box-sizing:border-box;height:40px;gap:12px;padding:8px 16px;border-bottom:1px solid #ddd6c7}
.num{width:64px}.title{width:200px;font-weight:bold}.body{color:#4a463c}
</style></head><body><div id="head">List workload</div>
<div style="height:{{.Top}}px"></div>
{{range .Rows}}<div id="{{.ID}}" class="row"><span class="num">{{.Num}}</span><span class="title">{{.Title}}</span><span class="body">{{.Body}}</span></div>{{end}}
<div style="height:{{.Bottom}}px"></div></body></html>`
