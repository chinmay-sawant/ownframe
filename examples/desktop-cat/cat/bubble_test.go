package cat

import (
	"context"
	"strings"
	"testing"
)

func TestNotificationAndClickableBubble(t *testing.T) {
	inbox := NewInbox()
	c, err := NewWithInbox(1, inbox)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := c.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := inbox.Submit(Notification{Expression: "sad", Message: "Have you had enough water?", Source: "Open Code"}); err != nil {
		t.Fatal(err)
	}
	if err := c.tick(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if c.animation.current != 18 || c.animation.cycle || !c.shown {
		t.Fatal("notification did not change cat and bubble")
	}
	if !c.Interactive(200, 200) || !c.Interactive(50, 50) || c.Interactive(5, 5) || c.Interactive(30, 200) {
		t.Fatal("incorrect input regions")
	}
	if err := c.Page.Click(ctx, 100, 60); err != nil {
		t.Fatal(err)
	}
	if c.shown || c.Interactive(50, 50) {
		t.Fatal("bubble did not dismiss")
	}
	if err := c.Page.Click(ctx, 200, 220); err != nil {
		t.Fatal(err)
	}
	if !c.shown || c.view.Message != "Have you had enough water?" {
		t.Fatal("cat did not reopen last message")
	}
	if err := inbox.Submit(Notification{Expression: "angry", Message: strings.Repeat("x", 161)}); err == nil {
		t.Fatal("long text accepted")
	}
	if c.Page.Display() == nil {
		t.Fatal("bubble forced bitmap fallback")
	}
}
