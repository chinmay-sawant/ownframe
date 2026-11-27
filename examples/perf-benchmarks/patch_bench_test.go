package perfbench_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

const patchListHTML = `<!DOCTYPE html><html><head><style>
html,body{margin:0;font-family:sans-serif;color:#1c1915;background:#f4f1ea}
#head{height:40px;background:#1c1915;color:#fff}
.row{display:flex;box-sizing:border-box;height:40px;gap:12px;padding:8px 16px;border-bottom:1px solid #ddd6c7}
.row.selected{background:#e8e2d5}
.row[data-active="1"]{background:#dce8d5}
.num{width:64px}.title{width:200px;font-weight:bold}.body{color:#4a463c}
</style></head><body><div id="head">Patch vs Rebuild</div>
{{range .}}<div id="row-{{.Num}}" class="row"><span class="num">{{.Num}}</span><span class="title">{{.Title}}</span><span id="body-{{.Num}}" class="body">{{.Body}}</span></div>{{end}}
</body></html>`

const structuralFeedHTML = `<!DOCTYPE html><html><head><style>
html,body{margin:0;font-family:sans-serif;color:#1c1915;background:#f4f1ea}
#feed{width:760px;height:520px;overflow:hidden;background:#ffffff}
.item{height:24px;padding:2px 8px;box-sizing:border-box;border-bottom:1px solid #e5e0d5}
</style></head><body><div id="feed">{{range .}}<div id="item-{{.Num}}" class="item">{{.Title}}</div>{{end}}</div></body></html>`

type patchRow struct {
	Num   int
	Title string
	Body  string
}

func makePatchRows(n int) []patchRow {
	rows := make([]patchRow, n)
	for i := range rows {
		idx := i + 1
		rows[i] = patchRow{
			Num:   idx,
			Title: "Item " + strconv.Itoa(idx),
			Body:  "Initial body " + strconv.Itoa(idx),
		}
	}

	return rows
}

func newPatchPage(ctx context.Context, b *testing.B, rows []patchRow) *ownframe.Page {
	b.Helper()
	p, err := ownframe.New(ownframe.Config{
		Title:       "Patch benchmark",
		HTML:        patchListHTML,
		Width:       800,
		Height:      600,
		RetainState: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	p.SetData(rows)
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	return p
}

// BenchmarkListPatchVsRebuild compares updating text in a 1,000-row unwindowed
// list via a full template rebuild vs. in-place PatchElement / PatchElements.
func BenchmarkListPatchVsRebuild(b *testing.B) {
	ctx := b.Context()
	const rowCount = 1000
	textA, textB := "changed body A", "changed body B"

	b.Run("FullRebuild1Row", func(b *testing.B) {
		rows := makePatchRows(rowCount)
		p := newPatchPage(ctx, b, rows)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			if i%2 == 0 {
				rows[3].Body = textA
			} else {
				rows[3].Body = textB
			}
			p.SetData(rows)
			if err := p.Redraw(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("PatchElement1Row", func(b *testing.B) {
		rows := makePatchRows(rowCount)
		p := newPatchPage(ctx, b, rows)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			text := textA
			if i%2 != 0 {
				text = textB
			}
			if err := p.PatchElement(ctx, "body-4", ownframe.ElementEdit{Text: &text}); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("PatchElements10Rows", func(b *testing.B) {
		rows := makePatchRows(rowCount)
		p := newPatchPage(ctx, b, rows)
		patchesA := make([]ownframe.ElementPatch, 10)
		patchesB := make([]ownframe.ElementPatch, 10)
		for j := range 10 {
			id := fmt.Sprintf("body-%d", j+1)
			patchesA[j] = ownframe.ElementPatch{
				ID:    id,
				Edits: []ownframe.ElementEdit{{Text: &textA}},
			}
			patchesB[j] = ownframe.ElementPatch{
				ID:    id,
				Edits: []ownframe.ElementEdit{{Text: &textB}},
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			batch := patchesA
			if i%2 != 0 {
				batch = patchesB
			}
			if err := p.PatchElements(ctx, batch); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkRetainedMutations guards the class-list, attribute-removal, and
// structural child-fragment mutation paths on retained pages.
func BenchmarkRetainedMutations(b *testing.B) {
	ctx := b.Context()
	const rowCount = 1000

	b.Run("ClassToggle1Row", func(b *testing.B) {
		rows := makePatchRows(rowCount)
		p := newPatchPage(ctx, b, rows)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			edit := ownframe.ElementEdit{AddClass: "selected"}
			if i%2 != 0 {
				edit = ownframe.ElementEdit{RemoveClass: "selected"}
			}
			if err := p.PatchElement(ctx, "row-4", edit); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("AttributeSetRemove1Row", func(b *testing.B) {
		rows := makePatchRows(rowCount)
		p := newPatchPage(ctx, b, rows)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			edit := ownframe.ElementEdit{Attribute: "data-active", Value: "1"}
			if i%2 != 0 {
				edit = ownframe.ElementEdit{RemoveAttribute: "data-active"}
			}
			if err := p.PatchElement(ctx, "row-4", edit); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("StructuralAppendRemove", func(b *testing.B) {
		p, err := ownframe.New(ownframe.Config{
			Title:       "Structural benchmark",
			HTML:        structuralFeedHTML,
			Width:       800,
			Height:      600,
			RetainState: true,
		})
		if err != nil {
			b.Fatal(err)
		}
		p.SetData(makePatchRows(50))
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		const frag = `<div id="dyn-item" class="item">appended row</div>`
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			edit := ownframe.ElementEdit{AppendHTML: frag}
			if i%2 != 0 {
				edit = ownframe.ElementEdit{ID: "dyn-item", Remove: true}
			}
			if err := p.PatchElement(ctx, "feed", edit); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("StructuralSetInnerHTML", func(b *testing.B) {
		p, err := ownframe.New(ownframe.Config{
			Title:       "InnerHTML benchmark",
			HTML:        structuralFeedHTML,
			Width:       800,
			Height:      600,
			RetainState: true,
		})
		if err != nil {
			b.Fatal(err)
		}
		p.SetData(makePatchRows(10))
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		fragA := `<div id="a1" class="item">Alpha 1</div><div id="a2" class="item">Alpha 2</div>`
		fragB := `<div id="b1" class="item">Beta 1</div><div id="b2" class="item">Beta 2</div>`
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			frag := fragA
			if i%2 != 0 {
				frag = fragB
			}
			if err := p.PatchElement(ctx, "feed", ownframe.ElementEdit{InnerHTML: &frag}); err != nil {
				b.Fatal(err)
			}
		}
	})
}
