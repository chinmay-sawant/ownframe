package render

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestReplayableAcceptsRoundedStroke(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{{Kind: layout.DisplayOpStrokeRect, Radius: 4}},
	}

	if !Replayable(display) {
		t.Fatal("rounded stroke should replay")
	}
}

func TestReplayableAcceptsMaskedStroke(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpStrokeRect, Radius: 4, StrokeMask: 1},
		},
	}

	if !Replayable(display) {
		t.Fatal("masked stroke should replay")
	}
}

func TestReplayableAcceptsEllipticalStroke(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpStrokeRect, Radius: 4, RadiusY: 2},
		},
	}

	if !Replayable(display) {
		t.Fatal("elliptical stroke should replay")
	}
}

func TestReplayableRejectsUnknownStrokeMask(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpStrokeRect, Radius: 4, StrokeMask: 16},
		},
	}

	if Replayable(display) {
		t.Fatal("unknown mask should not replay")
	}
}

func TestReplayableRejectsImageWithoutBytes(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{{Kind: layout.DisplayOpImage}},
	}

	if Replayable(display) {
		t.Fatal("image without bytes should not replay")
	}
}
