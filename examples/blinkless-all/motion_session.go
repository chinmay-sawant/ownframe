package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// The motion session moves one box along its path with the clock. The page
// holds a box (#mp-live) whose style names a 200px line. Each frame the
// session writes a new offset-distance for it, and the engine turns that
// distance into the box's matrix when the page is laid out again.
const (
	motionLoop = 4 * time.Second
	// Each write re-lays out the page, so the box moves in steps and the session
	// asks for a frame about ten times a second. At 33 ms and a pixel at a time
	// the page used more than a core while idle.
	motionFrame     = 100 * time.Millisecond
	motionStepPx    = 4.0
	motionLiveLen   = 200.0
	motionLiveID    = "mp-live"
	motionLiveRead  = "mp-live-read"
	motionLiveStyle = "offset-path: path('M 0 0 L 200 0'); offset-rotate: 0deg; offset-anchor: left top; offset-distance: %.1fpx"
)

// motionLiveHTML is the moving lane: the box starts at distance 0.
func motionLiveHTML() string {
	return `<div class="lane"><div id="` + motionLiveID + `" class="dot" style="` +
		fmt.Sprintf(motionLiveStyle, 0.0) + `">m</div><span class="lbl">live: offset-distance follows the clock</span></div>` +
		`<div id="` + motionLiveRead + `" class="rd">offset-distance: 0px of 200px</div>`
}

// motionDistance is how far along the line the box is at time now: it runs the
// full length once per loop, then starts over.
func motionDistance(now time.Time) float64 {
	into := time.Duration(now.UnixNano()) % motionLoop
	dist := float64(into) / float64(motionLoop) * motionLiveLen

	return math.Floor(dist/motionStepPx) * motionStepPx
}

type motionSession struct {
	lastStyle string
	lastText  string
	patches   int
}

func newMotionSession() *motionSession { return &motionSession{} }

// TabID is the tab the session belongs to.
func (s *motionSession) TabID() string { return "motion" }

// Frame writes the box's distance for this instant. It sits idle while the
// motion tab is not the open document.
func (s *motionSession) Frame(ctx context.Context, page *ownframe.Page, _ liveInput, now time.Time) (time.Time, error) {
	if !strings.Contains(page.HTML(), `id="`+motionLiveID+`"`) {
		s.lastStyle, s.lastText = "", ""

		return time.Time{}, nil
	}

	dist := motionDistance(now)
	style := fmt.Sprintf(motionLiveStyle, dist)
	text := fmt.Sprintf("offset-distance: %.0fpx of %.0fpx", dist, motionLiveLen)

	var batch []ownframe.ElementPatch

	if style != s.lastStyle {
		batch = append(batch, ownframe.ElementPatch{ID: motionLiveID,
			Edits: []ownframe.ElementEdit{{ID: motionLiveID, Attribute: "style", Value: style}}})
	}

	if text != s.lastText {
		batch = append(batch, ownframe.ElementPatch{ID: motionLiveRead,
			Edits: []ownframe.ElementEdit{{ID: motionLiveRead, Text: &text}}})
	}

	s.lastStyle, s.lastText = style, text
	s.patches += len(batch)

	return now.Add(motionFrame), page.PatchElements(ctx, batch)
}
