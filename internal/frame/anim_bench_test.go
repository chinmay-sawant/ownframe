package frame

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// cardHTML is one absolutely placed card with a label. The data field moves
// the card's left edge.
const cardHTML = `<style>` +
	`#card{position:absolute;left:{{.X}}px;top:40px;width:120px;height:80px;background:#3366cc}` +
	`#card span{position:absolute;left:12px;top:28px;color:#ffffff;font-size:16px}` +
	`</style><div id="card"><span>Board</span></div>`

const (
	benchFrames = 60
	benchStep   = 16 * time.Millisecond
)

// benchCard builds the card page and returns the card's fill and label ops.
func benchCard(b *testing.B) (*page.Page, []*ownframe.DisplayOp) {
	b.Helper()

	screen, err := page.New(page.Config{HTML: cardHTML, Width: 640, Height: 480})
	if err != nil {
		b.Fatal(err)
	}

	screen.SetData(struct{ X int }{X: 40})

	if err := screen.Redraw(context.Background()); err != nil {
		b.Fatal(err)
	}

	d := screen.Display()
	if d == nil {
		b.Fatal("no display list")
	}

	box := ownframe.Box{X: 20, Y: 20, W: 300, H: 200}
	fill := Fill(d, box, [3]float64{0.2, 0.4, 0.8})
	label := Text(d, box)
	if fill == nil || label == nil {
		b.Fatal("card ops not found")
	}

	return screen, []*ownframe.DisplayOp{fill, label}
}

// BenchmarkAnimCardRedraw animates the card with a Redraw per frame: one
// template execution, parse, cascade, and layout.
func BenchmarkAnimCardRedraw(b *testing.B) {
	screen, _ := benchCard(b)
	ctx := context.Background()
	data := struct{ X int }{}

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		data.X = 40 + i%benchFrames
		screen.SetData(data)

		if err := screen.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
