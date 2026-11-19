package main

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

// Sizes the live demos share. liveCharW is the advance of one character of
// the 14px monospace font the editable cells use, measured from a laid out
// ten character run (84.01px). navFieldPad is the horizontal padding and
// border of an editable field.
const (
	liveCharW   = 8.4013671875
	navFieldPad = 14
	// navMaxChars is the longest value a field takes.
	navMaxChars = 64
	navTextLeft = 4.0
	navTextTop  = 2.0
	navLineH    = 16.0
	imgFrameMs  = 200
	imgFrames   = 4
	layerMs     = 1000
)

// navView is what the session wants one element to look like.
type navView struct {
	Style   string
	Text    string
	HasText bool
}

// navScreen collects the views of one frame.
type navScreen struct {
	base map[string]string
	out  map[string]navView
}

func (sc *navScreen) style(id, extra string) {
	v := sc.out[id]
	v.Style = strings.TrimRight(sc.base[id], "; ") + ";" + extra
	sc.out[id] = v
}

func (sc *navScreen) text(id, t string) {
	v := sc.out[id]
	if v.Style == "" {
		v.Style = sc.base[id]
	}

	v.Text, v.HasText = t, true
	sc.out[id] = v
}

// navDemoOf maps a clickable id to the demo that owns it.
func navDemoOf(id string) string {
	switch {
	case strings.HasPrefix(id, "arr-"):
		parts := strings.Split(id, "-")

		return parts[1]
	case strings.HasPrefix(id, "n-"):
		return "focus"
	case strings.HasPrefix(id, "gn-"), strings.HasPrefix(id, "gg-"):
		return "grid"
	case strings.HasPrefix(id, "c-"), id == "o-x":
		return "contain"
	case id == "s-a":
		return "action"
	case strings.HasPrefix(id, "cr-"):
		return "caret"
	case strings.HasPrefix(id, "fs-"):
		return "field"
	case strings.HasPrefix(id, "pw-"):
		return "security"
	case strings.HasPrefix(id, "us-"):
		return "select"
	case strings.HasPrefix(id, "ta-"):
		return "adjust"
	case strings.HasPrefix(id, "sl-"):
		return "slider"
	case strings.HasPrefix(id, "drag-"):
		return "drag"
	case strings.HasPrefix(id, "wc"):
		return "layer"
	case strings.HasPrefix(id, "ap-"):
		return "appear"
	case strings.HasPrefix(id, "ia-"):
		return "image"
	}

	return ""
}

// navHotIDs are the elements a click can land on. The parts inside a control
// (its text, caret, fill, thumb and highlight) are not listed.
//
//nolint:gochecknoglobals // pixel sizes, timings and markup of the live demo
var navHotIDs = map[string]bool{
	"n-a": true, "n-b": true, "n-c": true, "n-d": true, "n-e": true, "n-f": true,
	"gn-a": true, "gn-c": true, "gg-a": true, "gg-c": true,
	"c-a": true, "c-b": true, "o-x": true, "s-a": true,
	"cr-bar": true, "cr-block": true, "cr-under": true, "cr-autoblock": true,
	"fs-content": true, "fs-fixed": true, "pw-auto": true, "pw-none": true,
	"us-text": true, "us-none": true, "us-all": true, "ta-btn": true,
	"sl-h": true, "sl-rtl": true, "sl-v": true,
	"drag-bar": true, "drag-btn": true, "drag-title": true, "drag-body": true,
	"wc": true, "wc2": true, "ap-none": true, "ap-btn": true, "ap-def": true, "ia-btn": true,
}

// navHot reports whether id is a click target.
func navHot(id string) bool { return navHotIDs[id] || strings.HasPrefix(id, "arr-") }

// navState is everything the session remembers between frames.
type navState struct {
	start        time.Time
	activeDemo   string
	focus        map[string]string
	chip         map[string]string
	scroll       [2]float64
	texts        map[string]string
	caretFocus   string
	fieldFocus   string
	secFocus     string
	lastActivity time.Time
	sel          map[string][2]int
	slider       map[string]float64
	slDrag       string
	winX, winY   float64
	dragging     bool
	grabX, grabY float64
	layerID      string
	layerStart   time.Time
	// layerFrames is how many frames each box moved; layerPaints is how many
	// times it had to be painted in those frames.
	layerFrames, layerPaints map[string]int
	pressed                  string
	inflate                  bool
	imgPaused                bool
	imgPauseAt               int
	imgShift                 time.Duration
}

// navSession is the live behavior of the nav tab.
type navSession struct {
	base   map[string]string
	active bool
	st     navState
	last   map[string]navView
	// patches counts the elements written, for tests.
	patches int
}

func newNavSession() *navSession { return &navSession{} }

// TabID names the tab.
func (s *navSession) TabID() string { return "nav" }

var (
	reTag   = regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	reID    = regexp.MustCompile(`\sid="([^"]*)"`)
	reStyle = regexp.MustCompile(`\sstyle="([^"]*)"`)
)

// baseStyles reads the authored style attribute of every element with an id.
func baseStyles(doc string) map[string]string {
	out := map[string]string{}

	for _, tag := range reTag.FindAllString(doc, -1) {
		id := reID.FindStringSubmatch(tag)
		if id == nil {
			continue
		}

		style := ""
		if m := reStyle.FindStringSubmatch(tag); m != nil {
			style = m[1]
		}

		out[id[1]] = style
	}

	return out
}

// reset starts the demo state again, as when the tab is opened.
//
//nolint:mnd // pixel sizes, timings and markup of the live demo
func (s *navSession) reset(doc string, now time.Time) {
	s.base = baseStyles(doc)
	s.last = map[string]navView{}
	s.st = navState{
		start: now, lastActivity: now, activeDemo: "focus",
		focus:       map[string]string{"focus": "n-a", "grid": "gn-a", "contain": "c-a", "action": "s-a"},
		chip:        map[string]string{},
		layerFrames: map[string]int{}, layerPaints: map[string]int{},
		texts:  map[string]string{"cr-bar": "type here", "cr-block": "type here", "cr-under": "type here", "cr-autoblock": "type here", "fs-content": "hello", "fs-fixed": "hello", "pw-auto": "hunter2", "pw-none": "hunter2"},
		sel:    map[string][2]int{},
		slider: map[string]float64{"sl-h": 0.25, "sl-rtl": 0.25, "sl-v": 0.25},
		winX:   20, winY: 20,
		caretFocus: "cr-bar", fieldFocus: "fs-content", secFocus: "pw-auto",
		imgPauseAt: -1,
	}
}

// Frame runs one frame of input.
//
//nolint:mnd // pixel sizes, timings and markup of the live demo
func (s *navSession) Frame(ctx context.Context, page *ownframe.Page, in liveInput, now time.Time) (time.Time, error) {
	doc := page.HTML()
	if !strings.Contains(doc, `id="sec-nav"`) {
		s.active = false

		return time.Time{}, nil
	}

	display := page.Probe().Display()
	if display == nil {
		return time.Time{}, nil
	}

	m := navModelOf(display)

	// A fresh document (the tab was opened, or its button was pressed again)
	// shows the authored markup, so the session starts over.
	if !s.active || s.reloaded(display.Boxes) {
		s.active = true
		s.reset(doc, now)
	}

	s.input(m, in, now)

	sc := &navScreen{base: s.base, out: map[string]navView{}}
	s.draw(sc, m, now)

	return now.Add(50 * time.Millisecond), s.apply(ctx, page, sc)
}

// reloaded reports whether the page shows the authored chip text again after
// the session had written something else there.
func (s *navSession) reloaded(boxes []layout.Box) bool {
	want, ok := s.last["chip-focus"]
	if !ok || !want.HasText {
		return false
	}

	for _, b := range boxes {
		if b.ID == "chip-focus" {
			return b.Text != want.Text
		}
	}

	return false
}

// input applies one frame of input to the state.
func (s *navSession) input(m navModel, in liveInput, now time.Time) {
	st := &s.st

	if in.Click {
		if id := navHit(m, in.CursorX, in.CursorY); id != "" {
			s.click(m, id, in, now)
		}
	}

	if !in.Down {
		st.slDrag, st.dragging, st.pressed = "", false, ""
	}

	s.hold(m, in)

	for _, k := range in.Keys {
		s.key(m, k, now)
	}

	if text := navPrintable(in.Text); text != "" {
		s.typeText(text, now)
	}
}

// navPrintable drops control characters from typed text.
func navPrintable(text string) string {
	var b strings.Builder

	for _, r := range text {
		if r >= ' ' && r != 0x7f {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// navHit returns the id of the deepest clickable element under the point.
func navHit(m navModel, x, y float64) string {
	hit := ""

	for _, it := range m.Items {
		if it.ID != "" && navHot(it.ID) && x >= it.X && x < it.X+it.W && y >= it.Y && y < it.Y+it.H {
			hit = it.ID
		}
	}

	return hit
}

func (s *navSession) click(m navModel, id string, in liveInput, now time.Time) {
	st := &s.st
	demo := navDemoOf(id)
	st.activeDemo = demo

	switch demo {
	case "focus", "grid", "contain", "action":
		if strings.HasPrefix(id, "arr-") {
			s.arrow(m, demo, navDirOf(id[strings.LastIndex(id, "-")+1:]))
		} else if demo != "action" {
			from := st.focus[demo]
			st.focus[demo] = id
			st.chip[demo] = fmt.Sprintf("clicked %s (was %s)", id, from)
		}
	case "caret":
		st.caretFocus, st.lastActivity = id, now
	case "field":
		st.fieldFocus = id
	case "security":
		st.secFocus = id
	case "select":
		it := m.Items[m.ByID(id)]
		pos := int(math.Floor((in.CursorX - it.X - navTextLeft) / liveCharW))
		lo, hi := navSelect(it.Declared["-webkit-user-select"], "pick a word here", pos)
		st.sel[id] = [2]int{lo, hi}
		st.chip["select"] = fmt.Sprintf("%s: click at %d selects [%d,%d) %q", id, pos, lo, hi, string([]rune("pick a word here")[lo:hi]))
	case "adjust":
		if id == "ta-btn" {
			st.inflate = !st.inflate
		}
	case "slider":
		st.slDrag = id
	case "drag":
		s.dragPress(m, id, in)
	case "layer":
		if id == "wc" || id == "wc2" {
			st.layerID, st.layerStart = id, now
			st.layerFrames[id], st.layerPaints[id] = 0, 0
		}
	case "appear":
		st.pressed = id
	case "image":
		s.imageClick(id, now)
	}
}

func navDirOf(name string) navDir {
	switch name {
	case "up":
		return navUp
	case "down":
		return navDown
	case "left":
		return navLeft
	}

	return navRight
}

// arrow runs one direction key press through the focus demo's host function.
func (s *navSession) arrow(m navModel, demo string, dir navDir) {
	st := &s.st
	from := st.focus[demo]
	names := [...]string{"up", "down", "left", "right"}
	r := m.Move(m.ByID(from), dir, navFocusable)

	switch {
	case r.Action == "scroll":
		before := st.scroll
		st.scroll[0] = clampTo(st.scroll[0]+r.ScrollDX, 0, navSheetW-navViewW)
		st.scroll[1] = clampTo(st.scroll[1]+r.ScrollDY, 0, navSheetH-navViewH)

		edge := ""
		if st.scroll == before {
			edge = " (already at the end of the sheet)"
		}

		st.chip[demo] = fmt.Sprintf("%s %s: scroll step (%g,%g), offset now (%g,%g)%s, focus stays on %s",
			from, names[dir], r.ScrollDX, r.ScrollDY, st.scroll[0], st.scroll[1], edge, from)
	case r.Target >= 0:
		to := m.Items[r.Target].ID
		st.focus[demo] = to
		st.chip[demo] = fmt.Sprintf("moved %s -> %s via %s", from, to, r.Via)
	default:
		st.chip[demo] = fmt.Sprintf("%s %s: no target, focus stays on %s", from, names[dir], from)
	}
}

func (s *navSession) key(m navModel, key string, now time.Time) {
	st := &s.st

	switch key {
	case "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight":
		demo := st.activeDemo
		if _, ok := st.focus[demo]; !ok {
			return
		}

		s.arrow(m, demo, navDirOf(strings.ToLower(strings.TrimPrefix(key, "Arrow"))))
	case "Backspace":
		s.editFocused(func(t string) string {
			r := []rune(t)
			if len(r) == 0 {
				return t
			}

			return string(r[:len(r)-1])
		}, now)
	}
}

func (s *navSession) typeText(text string, now time.Time) {
	s.editFocused(func(t string) string {
		r := []rune(t + text)
		if len(r) > navMaxChars {
			r = r[:navMaxChars]
		}

		return string(r)
	}, now)
}

// editFocused applies edit to the text cell of the active typing demo.
func (s *navSession) editFocused(edit func(string) string, now time.Time) {
	st := &s.st

	var id string

	switch st.activeDemo {
	case "caret":
		id = st.caretFocus
	case "field":
		id = st.fieldFocus
	case "security":
		id = st.secFocus
	default:
		return
	}

	st.texts[id] = edit(st.texts[id])
	st.lastActivity = now
}

// hold runs the parts that follow the pointer while the button is down.
func (s *navSession) hold(m navModel, in liveInput) {
	st := &s.st

	if st.slDrag != "" && in.Down {
		it := m.Items[m.ByID(st.slDrag)]
		orient := navSliderOrient(it.Declared["slider-orientation"], it.W, it.H)
		along := clamp01((in.CursorX - it.X) / it.W)

		if orient == "top-to-bottom" || orient == "bottom-to-top" {
			along = clamp01((in.CursorY - it.Y) / it.H)
		}

		st.slider[st.slDrag] = along
	}

	if st.dragging && in.Down {
		area := m.Items[m.ByID("drag-area")]
		win := m.Items[m.ByID("drag-win")]
		st.winX = clampTo(in.CursorX-area.X-st.grabX, 0, area.W-win.W)
		st.winY = clampTo(in.CursorY-area.Y-st.grabY, 0, area.H-win.H)
	}
}

func (s *navSession) dragPress(m navModel, id string, in liveInput) {
	st := &s.st
	i := m.ByID(id)

	if i < 0 || !m.navDragRegion(i) {
		st.chip["drag"] = fmt.Sprintf("press on %s: no drag (window-drag is none there)", id)

		return
	}

	area := m.Items[m.ByID("drag-area")]
	st.dragging = true
	st.grabX, st.grabY = in.CursorX-area.X-st.winX, in.CursorY-area.Y-st.winY
	st.chip["drag"] = fmt.Sprintf("press on %s: dragging (window-drag move)", id)
}

// imageElapsed is the clock the frame cycle reads, in milliseconds, with the
// time spent paused taken out.
func (s *navSession) imageElapsed(now time.Time) int {
	return int((now.Sub(s.st.start) - s.st.imgShift) / time.Millisecond)
}

func (s *navSession) imageClick(id string, now time.Time) {
	if id != "ia-btn" {
		return
	}

	st := &s.st
	el := s.imageElapsed(now)

	if st.imgPaused {
		st.imgShift += time.Duration(el-st.imgPauseAt) * time.Millisecond
		st.imgPaused, st.imgPauseAt = false, -1

		return
	}

	st.imgPaused, st.imgPauseAt = true, el
}

func clamp01(v float64) float64 { return clampTo(v, 0, 1) }

func clampTo(v, lo, hi float64) float64 {
	if hi < lo {
		hi = lo
	}

	return math.Min(math.Max(v, lo), hi)
}

// navSliderOrient resolves auto the way navSliderValue does.
func navSliderOrient(orient string, w, h float64) string {
	if orient == "" || orient == "auto" {
		if h > w {
			return "bottom-to-top"
		}

		return "left-to-right"
	}

	return orient
}

// apply writes the views that changed since the last frame.
func (s *navSession) apply(ctx context.Context, page *ownframe.Page, sc *navScreen) error {
	var batch []ownframe.ElementPatch

	for id, v := range sc.out {
		prev, seen := s.last[id]

		oldStyle := prev.Style
		if !seen {
			oldStyle = s.base[id]
		}

		var edits []ownframe.ElementEdit

		if v.Style != oldStyle {
			edits = append(edits, ownframe.ElementEdit{ID: id, Attribute: "style", Value: v.Style})
		}

		if v.HasText && (!seen || !prev.HasText || prev.Text != v.Text) {
			t := v.Text
			edits = append(edits, ownframe.ElementEdit{ID: id, Text: &t})
		}

		s.last[id] = v

		if len(edits) > 0 {
			batch = append(batch, ownframe.ElementPatch{ID: id, Edits: edits})
		}
	}

	s.patches += len(batch)

	return page.PatchElements(ctx, batch)
}

// written returns the style the session last wrote to an element.
func (s *navSession) written(id string) navView { return s.last[id] }
