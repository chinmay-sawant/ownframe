package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The demo plays a scripted session with no window. Each step sets what the
// window would report (page scroll offset, cursor, a click, a key), advances a
// manual clock, and runs the page's registered tick through
// Page.Testing().Tick. That is the same callback and the same session.frame
// that the real window drives, so a state change in the log is a state change
// a person would see.

// manualClock is the demo's time source.
type manualClock struct{ t time.Time }

func (c *manualClock) Now() time.Time { return c.t }

const demoStepTime = 100 * time.Millisecond // one frame at 10 fps

// Demo geometry and clock values, named where the script uses them.
const (
	// unboundedViewport is taller than any page, so its ScrollY is the largest offset.
	unboundedViewport = 1e9
	// showTopPx is how far below the window top a shown element is placed.
	showTopPx = 300
	// scrollBackStep is the size of each step in the scroll-up pass, in CSS pixels.
	scrollBackStep = 800
	// pointerEdgeInset is how far the pointer sits inside a box edge, in CSS pixels.
	pointerEdgeInset = 4
	// innerScrollSteps is the number of inner-scroller positions after the start.
	innerScrollSteps = 4
	// demoEpochUnix is the manual clock's start, so every run logs the same times.
	demoEpochUnix = 1_700_000_000
	// cursorMarkRadius is the half-size of the marker drawn at the cursor.
	cursorMarkRadius = 4
)

type demoStep struct {
	Label   string
	Scroll  float64 // offset the window reports to the page
	CursorX float64 // cursor in window coordinates
	CursorY float64
	Click   bool
	Keys    []string
	// InnerY is how far the host reports inner-scroller scrolled. OwnFrame
	// gives no offset for an inner box, so only the demo moves it.
	InnerY float64
}

// StepLog is what one step changed.
type StepLog struct {
	N       int
	Label   string
	Scroll  int
	Changes []string
	Clicked bool
	Hit     string
}

func (l StepLog) String() string {
	line := fmt.Sprintf("step %03d %-34s scroll=%-5d", l.N, l.Label, l.Scroll)
	if l.Clicked {
		hit := l.Hit
		if hit == "" {
			hit = "nothing"
		}

		line += " click-hit=" + hit
	}

	if len(l.Changes) > 0 {
		line += " | " + strings.Join(l.Changes, "; ")
	}

	return line
}

// parkX is a cursor column that sits over no card in the fixture.
const parkX, parkY = 860, 8

func demoScript(s *session) []demoStep {
	var steps []demoStep

	add := func(label string, scroll, cx, cy float64) {
		steps = append(steps, demoStep{Label: label, Scroll: scroll, CursorX: cx, CursorY: cy})
	}
	clampScroll := func(v float64) float64 { return s.viewport(v).ScrollY }
	rect := func(id string) Rect {
		i, _ := s.host.IndexOf(id)

		return s.host.Nodes[i].Rect
	}
	// show scrolls so the element sits 300px down the window and returns the
	// scroll and the cursor point at the element's middle.
	show := func(id string) (scroll, cx, cy float64) {
		r := rect(id)
		scroll = clampScroll(r.Y - showTopPx)

		return scroll, r.X + r.W/2, r.Y + r.H/2 - scroll
	}

	max := s.viewport(unboundedViewport).ScrollY

	for y := 0.0; y < max; y += 400 {
		add("scroll down", y, parkX, parkY)
	}

	add("scroll bottom", max, parkX, parkY)

	for y := max - scrollBackStep; y > 0; y -= scrollBackStep {
		add("scroll up", y, parkX, parkY)
	}

	add("scroll top", 0, parkX, parkY)

	for _, id := range []string{"pointer-box", "pointer-box-longhand"} {
		sc, cx, cy := show(id)
		r := rect(id)
		add(id+" arrive", sc, parkX, parkY)
		add(id+" pointer left", sc, r.X+pointerEdgeInset, cy-(r.H/2-pointerEdgeInset))
		add(id+" pointer middle", sc, cx, cy)
		add(id+" pointer right", sc, r.X+r.W-pointerEdgeInset, cy+(r.H/2-pointerEdgeInset))
		add(id+" pointer away", sc, parkX, parkY)
	}

	for _, id := range []string{"interest-card", "interest-card-longhand"} {
		sc, cx, cy := show(id)
		add(id+" arrive", sc, parkX, parkY)

		for i := 0; i < 7; i++ {
			add(id+" hover", sc, cx, cy)
		}

		for i := 0; i < 7; i++ {
			add(id+" leave", sc, parkX, parkY)
		}
	}

	for _, id := range []string{"event-box", "event-box-longhand"} {
		sc, cx, cy := show(id)
		add(id+" arrive", sc, parkX, parkY)
		steps = append(steps, demoStep{Label: id + " click", Scroll: sc, CursorX: cx, CursorY: cy, Click: true})
		steps = append(steps, demoStep{Label: id + " press Enter", Scroll: sc, CursorX: cx, CursorY: cy, Keys: []string{"Enter"}})
	}

	// The inner scroller scrolls on its own: its scroll timeline and the view
	// timeline of the card inside it follow InnerY, not the page.
	if port, _, ok := s.host.scrollerViewport("inner-scroller"); ok {
		sc, _, _ := show("inner-scroller")
		add("inner-scroller arrive", sc, parkX, parkY)

		for k := 0; k <= 4; k++ {
			steps = append(steps, demoStep{Label: "inner-scroller scroll", Scroll: sc, CursorX: parkX, CursorY: parkY,
				InnerY: (port.ContentH - port.H) * float64(k) / innerScrollSteps})
		}

		add("inner-scroller back", sc, parkX, parkY)
	}

	for _, id := range []string{"inert-panel", "auto-panel"} {
		sc, cx, cy := show(id)
		add(id+" arrive", sc, parkX, parkY)
		steps = append(steps, demoStep{Label: id + " click", Scroll: sc, CursorX: cx, CursorY: cy, Click: true})
	}

	return steps
}

// describe lists the demo-visible state of a frame as sorted "key" -> "value"
// pairs, so two frames can be compared.
func (h *Host) describe(f Frame, interest map[int]bool) map[string]string {
	out := map[string]string{}

	for key, v := range f.Timelines {
		id := h.Nodes[key.Node].ID
		if v.Inactive {
			out[fmt.Sprintf("%s %s", id, key.Name)] = "inactive"

			continue
		}

		switch v.Kind {
		case KindPointer:
			if v.Inside {
				out[fmt.Sprintf("%s %s", id, key.Name)] = fmt.Sprintf("%.2f", v.Progress)
			} else {
				out[fmt.Sprintf("%s %s", id, key.Name)] = "idle"
			}
		default:
			out[fmt.Sprintf("%s %s", id, key.Name)] = fmt.Sprintf("%.2f", v.Progress)
		}
	}

	for i, on := range f.TriggerOn {
		out[fmt.Sprintf("%s trigger %s", h.Nodes[h.Triggers[i].Owner].ID, h.Triggers[i].Name)] = onOff(on, "on", "off")
	}

	for i, on := range f.EventOn {
		out[fmt.Sprintf("%s event %s", h.Nodes[h.Events[i].Owner].ID, h.Events[i].Name)] = onOff(on, "armed", "cleared")
	}

	for idx, on := range interest {
		out[fmt.Sprintf("%s interest", h.Nodes[idx].ID)] = onOff(on, "started", "cleared")
	}

	return out
}

func onOff(on bool, yes, no string) string {
	if on {
		return yes
	}

	return no
}

// diffState lists what changed from prev to cur, quiet values omitted on the
// first frame so the log does not open with a wall of zeros.
func diffState(prev, cur map[string]string) []string {
	var keys []string

	for k, v := range cur {
		old, seen := prev[k]
		if !seen {
			old = quiet(v)
		}

		if old != v {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = k + "=" + cur[k]
	}

	return out
}

// quiet is the value a fresh state is assumed to start at.
func quiet(v string) string {
	switch v {
	case "armed", "started", "on":
		return "cleared"
	case "idle":
		return "idle"
	}

	if strings.Contains(v, ".") {
		return "0.00"
	}

	return v
}

// runDemo plays the script. With frames true it writes one cropped PNG per
// step into dir. It returns one log entry per step.
func runDemo(ctx context.Context, html string, width, height int, dir string, frames bool, out io.Writer) ([]StepLog, error) {
	clock := &manualClock{t: time.Unix(demoEpochUnix, 0)}

	s, err := newSession(ctx, html, width, height, clock)
	if err != nil {
		return nil, err
	}

	var cur demoStep

	s.attach(inputSource{
		cursor: func() (float64, float64) { return cur.CursorX, cur.CursorY },
		click:  func() bool { return cur.Click },
		keys:   func() []string { return cur.Keys },
	})

	steps := demoScript(s)
	prev := map[string]string{}
	logs := make([]StepLog, 0, len(steps))

	for n, step := range steps {
		cur = step
		clock.t = clock.t.Add(demoStepTime)

		// The window reports the offset to the page after its wheel stage.
		s.page.Scroll().SetScrollOffset(0, int(step.Scroll))
		s.host.SetScrollerOffset("inner-scroller", 0, step.InnerY)

		if err = s.page.Testing().Tick(ctx); err != nil {
			return logs, fmt.Errorf("step %d: %w", n, err)
		}

		res := s.last
		state := s.host.describe(res.Frame, res.Frame.InterestOn)

		l := StepLog{N: n, Label: step.Label, Scroll: int(s.viewport(step.Scroll).ScrollY),
			Changes: diffState(prev, state), Clicked: res.Clicked, Hit: res.Hit}
		prev = state

		logs = append(logs, l)

		if out != nil {
			if _, err = fmt.Fprintln(out, l.String()); err != nil {
				return logs, err
			}
		}

		if frames {
			if err = writeFrame(ctx, s, dir, n, l.Scroll, step); err != nil {
				return logs, err
			}
		}
	}

	return logs, nil
}

// writeFrame saves what the window would show: the viewport slice of the full
// page picture, with a small marker at the cursor.
func writeFrame(ctx context.Context, s *session, dir string, n, scroll int, step demoStep) error {
	full, err := pngOf(ctx, s)
	if err != nil {
		return err
	}

	img, err := png.Decode(bytes.NewReader(full))
	if err != nil {
		return err
	}

	w, h := s.page.Size()
	view := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			view.Set(x, y, img.At(img.Bounds().Min.X+x, img.Bounds().Min.Y+scroll+y))
		}
	}

	for dy := -cursorMarkRadius; dy <= cursorMarkRadius; dy++ {
		for dx := -cursorMarkRadius; dx <= cursorMarkRadius; dx++ {
			c := color.RGBA{R: 220, G: 38, B: 38, A: 255} //nolint:mnd // marker color, RGBA table
			if dx == -4 || dx == 4 || dy == -4 || dy == 4 {
				c = color.RGBA{R: 255, G: 255, B: 255, A: 255} //nolint:mnd // marker outline color, RGBA table
			}

			x, y := int(step.CursorX)+dx, int(step.CursorY)+dy
			if image.Pt(x, y).In(view.Bounds()) {
				view.Set(x, y, c)
			}
		}
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, view); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("frame-%04d.png", n)), buf.Bytes(), pngFileMode)
}

// stitch joins the frames into an mp4 at 10 fps when ffmpeg is installed. It
// reports the mp4 path, or "" with a plain reason when it did not make one.
func stitch(ctx context.Context, dir string) (string, string) {
	ff, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", "ffmpeg not found on PATH; only PNG frames were written"
	}

	mp4 := filepath.Join(dir, "blinkless-scroll-demo.mp4")
	cmd := exec.CommandContext(ctx, ff, "-y", "-loglevel", "error", "-framerate", "10", //nolint:gosec // G204: ff comes from exec.LookPath and every argument is a fixed flag or a path this demo wrote
		"-i", filepath.Join(dir, "frame-%04d.png"),
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2", "-pix_fmt", "yuv420p", mp4)

	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Sprintf("ffmpeg failed (%v): %s; only PNG frames were written", err, strings.TrimSpace(string(out)))
	}

	return mp4, ""
}
