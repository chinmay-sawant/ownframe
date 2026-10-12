package board

import (
	"math/rand"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

const glyphs = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "

type flap struct {
	op          *ownframe.DisplayOp
	want, cur   string
	delay, spin int
	done        bool
}

type flipper struct {
	gen        uint64
	cells      map[string]*flap
	order      []string
	rnd        *rand.Rand
	clock, bar *ownframe.DisplayOp
	clockText  string
	barFullW   float64
}

// rebind reacquires the tick's ops after a Redraw.
func (f *flipper) rebind(d *ownframe.Display, boxes []ownframe.Box, cells []Cell) {
	if d == nil {
		return
	}
	byID := make(map[string]ownframe.Box)
	for _, b := range boxes {
		byID[b.ID] = b
	}
	var order []string
	next := make(map[string]*flap)
	for i, c := range cells {
		b, ok := byID[c.ID]
		if !ok {
			continue
		}
		op := frame.Text(d, b)
		if op == nil {
			continue
		}
		fl := f.cells[c.ID]
		if fl != nil && fl.want == c.Ch {
			fl.op = op
		} else {
			cur := op.Text
			if fl != nil {
				cur = fl.op.Text
			}
			fl = &flap{op: op, want: c.Ch, cur: cur, delay: i % 8, spin: 8 + i%6, done: cur == c.Ch}
		}
		next[c.ID] = fl
		order = append(order, c.ID)
	}
	f.cells, f.order = next, order
	f.clock = nil
	if b, ok := byID["clock"]; ok {
		f.clock = frame.Text(d, b)
	}
}

func (f *flipper) bindBar(d *ownframe.Display, boxes []ownframe.Box, accent [3]float64) {
	f.bar, f.barFullW = nil, 0
	if d == nil {
		return
	}
	b, ok := boxByID(boxes, "refresh-fill")
	if !ok {
		return
	}
	f.bar = frame.Fill(d, b, accent)
	_, _, f.barFullW, _ = frame.BoxUnits(d, b)
}

func (f *flipper) step() {
	if f.rnd == nil {
		f.rnd = rand.New(rand.NewSource(1))
	}
	for _, id := range f.order {
		fl := f.cells[id]
		if fl == nil || fl.done {
			continue
		}
		if fl.delay > 0 {
			fl.delay--
			continue
		}
		if fl.spin > 0 {
			fl.spin--
			fl.op.Text = string(glyphs[f.rnd.Intn(len(glyphs))])
			continue
		}
		fl.op.Text = fl.want
		fl.done = true
	}
}
