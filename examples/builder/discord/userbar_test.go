package main

import (
	"context"
	"strings"
	"testing"

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
