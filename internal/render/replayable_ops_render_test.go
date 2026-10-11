package render

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestReplayableAcceptsTransformedImage(t *testing.T) {
	t.Parallel()
	source := `<html><body><div style="width:20px;height:20px;` +
		`background:linear-gradient(red,blue);transform:rotate(20deg)"></div></body></html>`
	display, err := DisplayList(context.Background(), source, 320, 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range display.Ops {
		if op.Kind != layout.DisplayOpImage || !op.XformSet {
			continue
		}
		if !Replayable(&layout.Display{Ops: []layout.DisplayOp{op}}) {
			t.Fatal("transformed image should replay")
		}
		return
	}
	t.Fatal("transformed svg produced no transformed image op")
}

func TestReplayableAcceptsImageWithBytes(t *testing.T) {
	t.Parallel()
	source := `<html><body><svg width="10" height="10">` +
		`<rect width="10" height="10" fill="red"/></svg></body></html>`
	display, err := DisplayList(context.Background(), source, 320, 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range display.Ops {
		if op.Kind != layout.DisplayOpImage {
			continue
		}
		if !Replayable(&layout.Display{Ops: []layout.DisplayOp{op}}) {
			t.Fatal("image with bytes should replay")
		}
		return
	}
	t.Fatal("inline svg produced no image op")
}

func TestReplayableAcceptsLetterSpacing(t *testing.T) {
	t.Parallel()
	display, err := DisplayList(context.Background(), `<html><body><p style="letter-spacing:2px">hi</p></body></html>`, 320, 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range display.Ops {
		if op.Kind != layout.DisplayOpText || op.LetterSpacing == 0 {
			continue
		}
		if !Replayable(&layout.Display{Ops: []layout.DisplayOp{op}}) {
			t.Fatal("letter-spaced text should replay")
		}
		return
	}
	t.Fatal("no letter-spaced text op")
}
