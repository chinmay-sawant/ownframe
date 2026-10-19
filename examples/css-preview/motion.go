package main

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

type pair struct {
	op       *ownframe.DisplayOp
	from, to paint
}

type motion struct {
	animation  *animation.Player
	transition *transition.Renderer
	pairs      []pair
	start      time.Time
	duration   time.Duration
	delay      time.Duration
	endDelay   time.Duration
}

func (m *motion) step(now time.Time) bool {
	if m.animation != nil {
		return m.animation.Step(now)
	}
	if m.transition != nil {
		return m.transition.Step(now)
	}
	return false
}
