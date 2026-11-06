package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// Joining voice reports a transient toast instead of a persistent strip in
// the message column; the sidebar panel keeps the mute/deafen/leave
// controls.
func TestVoiceJoinToastsAndLeavesNoStrip(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "voice:join:ownframe-lobby"}); err != nil {
		t.Fatal(err)
	}
	if a.voiceID != "ownframe-lobby" {
		t.Fatalf("voiceID = %q, want ownframe-lobby", a.voiceID)
	}
	if !a.toastOpen || a.toastTitle != "Voice connected" {
		t.Fatalf("toast = %v %q, want open Voice connected", a.toastOpen, a.toastTitle)
	}
	if strings.Contains(a.page.HTML(), "dc-voicebar-leave") {
		t.Fatal("message column keeps a voice strip")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "voice:leave"}); err != nil {
		t.Fatal(err)
	}
	if a.voiceID != "" {
		t.Fatalf("voiceID = %q, want cleared", a.voiceID)
	}
	if !a.toastOpen || a.toastTitle != "Voice disconnected" {
		t.Fatalf("leave toast = %v %q", a.toastOpen, a.toastTitle)
	}
}

// Toast alerts clear on the typing tick once past their lifetime.
func TestToastExpiresOnTick(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.showToast(ctx, "Voice connected", "Lobby"); err != nil {
		t.Fatal(err)
	}
	if !a.toastOpen {
		t.Fatal("toast did not open")
	}
	// A fresh toast survives a tick.
	if err := a.tickTyping(ctx); err != nil {
		t.Fatal(err)
	}
	if !a.toastOpen {
		t.Fatal("fresh toast cleared early")
	}
	// An expired toast clears on the next tick.
	a.toastUntil = a.toastUntil.Add(-2 * toastLifetime)
	if err := a.tickTyping(ctx); err != nil {
		t.Fatal(err)
	}
	if a.toastOpen {
		t.Fatal("expired toast did not clear")
	}
}
