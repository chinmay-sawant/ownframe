package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// Bottom-left user bar opens the self profile, and mute/deafen pressed
// states tint red so the two toggles read apart from the icons alone.
func TestUserBarOpensSelfProfile(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := pagetest.BoxByID(a.page, "dc-user-profile"); !ok {
		t.Fatal("missing dc-user-profile")
	}
	if err := pagetest.Click(ctx, a.page, "dc-user-profile"); err != nil {
		t.Fatalf("click user profile: %v", err)
	}
	if a.profileOpen == "" {
		t.Fatal("user bar click did not open profile")
	}
}

// Voice icons use explicit strokes per theme: light on dark backgrounds,
// dark on light ones, red while pressed. Standalone img rasterization
// drops currentColor to black, so inheriting is not an option.
func TestVoiceIconsMatchTheme(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); !strings.Contains(html, `src="dc-mic-light"`) {
		t.Fatal("dark mode user bar misses the light-stroke mic")
	}
	a.appearance = dcLight
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); !strings.Contains(html, `src="dc-mic-dark"`) {
		t.Fatal("light mode user bar misses the dark-stroke mic")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "voice:mute"}); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); !strings.Contains(html, `src="dc-mic-off-red"`) {
		t.Fatal("muted mic misses the red-stroke icon")
	}
}

// Light paint pins the mention pill and badge rules the theme switch owns.
func TestLightPaintPinsBadgeAndMention(t *testing.T) {
	a := &app{appearance: dcLight}
	sheet := a.themeSheet()
	for _, rule := range []string{
		".of-scope--light .dc-chan-badge",
		".of-scope--light .dc-msg-mention",
		".of-scope--light .dc-member:hover",
		".of-scope--light .dc-member-name",
		".of-scope--light .dc-typing",
		".of-scope--light .dc-userbar",
		".of-scope--light .dc-userbar-name",
		".of-scope--light .dc-voice",
		".of-scope--light .dc-voice-name",
		".of-scope--light .dc-vc-name",
		".of-scope--light .dc-rail .of-tooltip-content",
		".of-scope--light .dc-st-card .dc-st-row div span:first-child",
		".of-scope--light .dc-st-card .dc-st-row div span:last-child",
	} {
		if !strings.Contains(sheet, rule) {
			t.Fatalf("theme sheet misses %q", rule)
		}
	}
}

// Rail hover never changes geometry: tiles stay circles until selected.
func TestRailHoverKeepsShape(t *testing.T) {
	start := strings.Index(appCSS, ".dc-server:hover {")
	if start < 0 {
		t.Fatal("missing .dc-server:hover rule")
	}
	end := strings.Index(appCSS[start:], "}")
	if end < 0 {
		t.Fatal("unterminated .dc-server:hover rule")
	}
	block := appCSS[start : start+end]
	if strings.Contains(block, "border-radius") {
		t.Fatalf("hover morphs shape: %q", block)
	}
}
