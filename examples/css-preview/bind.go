package main

import (
	"errors"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

func bind(d *ownframe.Display, b ownframe.Box, t target, base, to []paint, origin time.Time) (motion, error) {
	if d == nil {
		return motion{}, nil
	}
	duration, delay, endDelay := motionTiming(b, t)
	m := motion{duration: duration, delay: delay, endDelay: endDelay, start: origin}
	indices := cssmotion.OperationIndices(d, b)
	ops := cssmotion.Resolve(d, indices)
	if len(ops) != len(indices) || len(ops) != len(base) {
		return motion{}, errors.New("csspreview: captured target operations changed during redraw")
	}
	for i, op := range ops {
		dest := read(op)
		if to != nil && i < len(to) {
			dest = to[i]
		}
		m.pairs = append(m.pairs, pair{op, base[i], dest})
	}
	if m.duration <= 0 || len(m.pairs) == 0 {
		return m, nil
	}
	var err error
	if t.kind == "animation" {
		m.animation, err = cssanimation.Bind(d, b, indices, base, origin)
	} else {
		m.transition, err = csstransition.BindTo(d, b, t.property, indices, base, to, origin)
	}
	if err != nil {
		return motion{}, err
	}
	return m, nil
}
