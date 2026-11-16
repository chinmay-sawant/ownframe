package window

import (
	"runtime"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func TestTitlebarRespectsAppOwnedChrome(t *testing.T) {
	if runtime.GOOS == "js" || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		t.Skip("desktop titlebar")
	}
	if !usesTitlebar(Options{}) {
		t.Fatal("default desktop window omitted shared titlebar")
	}
	for _, o := range []Options{
		{CustomHeader: true}, {Borderless: true}, {Transparent: true},
		{MousePassthrough: true}, {Interactive: func(int, int) bool { return true }}, {NativeHeader: true},
	} {
		if usesTitlebar(o) {
			t.Fatalf("titlebar replaced explicitly configured chrome: %+v", o)
		}
	}
}

func TestHeaderTitleUsesProportionalFontAndFits(t *testing.T) {
	face, err := newHeaderFont()
	if err != nil {
		t.Fatal(err)
	}
	wide, _ := text.Measure("WWW", face, 0)
	narrow, _ := text.Measure("III", face, 0)
	if wide <= narrow {
		t.Fatal("header title still uses a fixed-width face")
	}
	const title = "Workspace — a long window title"
	got := fitHeaderTitle(title, face, 90)
	width, _ := text.Measure(got, face, 0)
	if !strings.HasSuffix(got, "…") || width > 90 {
		t.Fatalf("fitted title=%q width=%f", got, width)
	}
	if got := fitHeaderTitle("Workspace", face, 200); got != "Workspace" {
		t.Fatalf("short title changed: %q", got)
	}
}

func TestTitlebarKeepsPageViewport(t *testing.T) {
	s := &shell{titlebar: titlebarState{enabled: true}}
	w, h := s.Layout(1000, 836)
	if w != 1000 || h != 836 || s.pendingH != 800 || s.screenH != 800 {
		t.Fatalf("canvas=%dx%d content=%dx%d screenH=%d", w, h, s.pendingW, s.pendingH, s.screenH)
	}
	s.titlebar.enabled = false
	s.Layout(1000, 800)
	if s.pendingH != 800 {
		t.Fatalf("custom chrome viewport changed: %d", s.pendingH)
	}
}

func TestCustomHeaderHeightKeepsPageViewport(t *testing.T) {
	s := &shell{titlebar: titlebarState{enabled: true, style: headerStyle{height: headerHeight(Options{HeaderHeight: 48})}}}
	w, h := s.Layout(900, 648)
	if w != 900 || h != 648 || s.pendingH != 600 || s.titlebarInset() != 48 {
		t.Fatalf("canvas=%dx%d contentHeight=%d headerHeight=%d", w, h, s.pendingH, s.titlebarInset())
	}
	if headerHeight(Options{}) != 36 || headerHeight(Options{HeaderHeight: 1}) != 24 {
		t.Fatal("header default or control minimum changed")
	}
}

func TestTitlebarControlHitBounds(t *testing.T) {
	for _, tc := range []struct{ x, want int }{{0, 0}, {861, 0}, {862, 1}, {907, 1}, {908, 2}, {953, 2}, {954, 3}, {999, 3}, {1000, 0}} {
		if got := titlebarButton(tc.x, 1000); got != tc.want {
			t.Fatalf("x=%d: control=%d want=%d", tc.x, got, tc.want)
		}
	}
}
