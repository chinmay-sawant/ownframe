package replay

import (
	"reflect"

	"github.com/chinmay-sawant/blinkless/layout"
)

// sigGrid folds a grid run's segments, or a zero marker for every other op.
func sigGrid(h uint64, op *layout.DisplayOp) uint64 {
	if op.Grid == nil {
		return mix64(h, 0)
	}

	h = mix64(h, uint64(len(op.Grid.Segs))+1)

	for i := range op.Grid.Segs {
		s := &op.Grid.Segs[i]
		for _, v := range [8]float64{s.X, s.Y, s.W, s.H, s.Width, s.R, s.G, s.B} {
			h = mixF(h, v)
		}

		h = mix64(h, uint64(s.LineInset))
	}

	return h
}

// sigPath folds a corner path's command list, or a zero marker for every
// other op. The whole point triple and every verb are folded, so a mutated
// payload cannot keep a stale cached raster.
func sigPath(h uint64, op *layout.DisplayOp) uint64 {
	corner, ok := op.CornerPathValue()
	if !ok || corner == nil {
		return mix64(h, 0)
	}

	h = mix64(h, uint64(len(corner.Cmds))+1)
	h = mix64(h, bit(corner.Stroke))

	for i := range corner.Cmds {
		cmd := &corner.Cmds[i]
		h = mix64(h, uint64(cmd.Verb))

		for _, p := range cmd.Pts {
			h = mixF(h, p.X)
			h = mixF(h, p.Y)
		}
	}

	return h
}

// sigImage folds an image payload's identity, or a zero marker. The payload
// hash is the same one the decoded-image cache keys on.
func sigImage(h uint64, op *layout.DisplayOp) uint64 {
	if op.Kind != layout.DisplayOpImage {
		return mix64(h, 0)
	}

	return mix64(h, imageBytesKey(op))
}

// sigXform folds a baked transform, or a zero marker.
func sigXform(h uint64, op *layout.DisplayOp) uint64 {
	if !op.XformSet {
		return mix64(h, 0)
	}

	m := op.Transform()

	for _, v := range [6]float64{m.A, m.B, m.C, m.D, m.E, m.F} {
		h = mixF(h, v)
	}

	return h
}

// ptrBits is the address behind an interface holding a pointer, or zero.
func ptrBits(v any) uint64 {
	if v == nil {
		return 0
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return 0
	}

	return uint64(rv.Pointer())
}
