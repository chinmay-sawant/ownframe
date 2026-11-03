package page

import (
	"context"
	"testing"
	"time"
)

func TestStaticPageFrameEvents(t *testing.T) {
	p := tickBarPage(t)
	events := p.FrameEvents()
	select {
	case <-events:
	default:
	}
	if p.FramePolling() {
		t.Fatal("static page requested polling")
	}
	p.WantFrameIn(time.Second)
	select {
	case <-events:
	default:
		t.Fatal("external deadline did not notify host")
	}
	if p.NextFrameAt().IsZero() {
		t.Fatal("notification consumed deadline")
	}
	p.SetTick(func(context.Context) error { p.WantFrameIn(time.Second); return nil })
	if !p.FramePolling() {
		t.Fatal("legacy callback lost keep-alive")
	}
	select {
	case <-events:
	default:
	}
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-events:
		t.Fatal("tick notified itself instead of scheduling deadline")
	default:
	}
	p.SetTick(nil)
	if p.FramePolling() {
		t.Fatal("removed callback still polls")
	}
}

func TestCaretSuppliesIdleDeadline(t *testing.T) {
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	p.setCaret(2, false)
	at := time.Now()
	p.formState().BlinkAt = at
	if p.FramePolling() {
		t.Fatal("caret unnecessarily enabled compatibility polling")
	}
	if got := p.NextFrameAt(); !got.Equal(at.Add(caretBlinkInterval)) {
		t.Fatalf("caret deadline = %v, want %v", got, at.Add(caretBlinkInterval))
	}
	earlier := at.Add(time.Millisecond)
	p.WantFrameAt(earlier)
	if got := p.NextFrameAt(); !got.Equal(earlier) {
		t.Fatalf("caret delayed earlier request: %v", got)
	}
	p.formState().All = true
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("selection scheduled a caret: %v", got)
	}
}
