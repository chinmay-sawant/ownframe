package telegram_test

import (
	"context"
	"testing"
)

func TestTickOptsIntoCachedFrames(t *testing.T) {
	app := newApp(t, context.Background())
	if err := app.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !app.Page().FrameDirty() {
		t.Fatal("tick did not opt into cached frame drawing")
	}
}
