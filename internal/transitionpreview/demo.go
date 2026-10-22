package transitionpreview

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

type sample struct {
	base    []cssmotion.Paint
	indices []int
	start   time.Time
	player  *transition.Renderer
	active  bool
}

type demo struct {
	page    *ownframe.Page
	data    *view
	gen     uint64
	samples map[string]*sample
	pending string
}

func newDemo(page *ownframe.Page, data *view) *demo {
	return &demo{page: page, data: data, gen: page.Testing().Generation(), samples: map[string]*sample{}}
}

func (d *demo) click(_ context.Context, box ownframe.Box) error {
	item, ok := byButton(box.ID)
	if !ok {
		return nil
	}
	if d.samples[item.key] == nil {
		targetBox, found := findBox(d.page.Boxes(), item.id)
		if !found {
			return nil
		}
		d.samples[item.key] = &sample{
			base:    cssmotion.Capture(d.page.Testing().Display(), targetBox),
			indices: cssmotion.OperationIndices(d.page.Testing().Display(), targetBox),
		}
	}
	d.samples[item.key].start = time.Now()
	d.data.Started[item.key] = true
	d.pending = item.key
	d.page.SetData(d.data)
	d.page.SetTick(d.tick)
	return nil
}
