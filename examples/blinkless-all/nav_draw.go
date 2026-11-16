package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	ringStyle = "box-shadow:0 0 0 3px #2563eb;background:#dbeafe"
)

//nolint:gochecknoglobals // pixel sizes, timings and markup of the live demo
var imgColors = [imgFrames]string{"#6366f1", "#16a34a", "#d97706", "#dc2626"}

// draw turns the state into the views of every element it drives.
func (s *navSession) draw(sc *navScreen, m navModel, now time.Time) {
	s.drawFocus(sc, m)
	s.drawCaret(sc, m, now)
	s.drawFields(sc, m)
	s.drawSelect(sc, m)
	s.drawAdjust(sc)
	s.drawSlider(sc, m)
	s.drawDrag(sc)
	s.drawLayer(sc, m, now)
	s.drawAppear(sc, m)
	s.drawImage(sc, m, now)
}

func (s *navSession) drawFocus(sc *navScreen, m navModel) {
	st := &s.st
	ids := map[string][]string{
		"focus":   {"n-a", "n-b", "n-c", "n-d", "n-e", "n-f"},
		"grid":    {"gn-a", "gn-c", "gg-a", "gg-c"},
		"contain": {"c-a", "c-b", "o-x"},
		"action":  {"s-a"},
	}

	for demo, list := range ids {
		for _, id := range list {
			switch {
			case demo == "action":
				sc.style("s-content", fmt.Sprintf("margin-left:%gpx;margin-top:%gpx", offsetOf(st.scroll[0]), offsetOf(st.scroll[1])))
				// A patch re-places only the element it names. Touching the box
				// that holds the sheet, with a value that changes whenever the
				// offset does, makes the sheet move on every key, in both
				// directions on both axes.
				sc.style("sbox", fmt.Sprintf("z-index:%d", int(st.scroll[0]/scrollStep)*sheetRows+int(st.scroll[1]/scrollStep)))
				sc.style(id, ringStyle)
			case st.focus[demo] == id:
				sc.style(id, ringStyle)
			default:
				sc.style(id, "")
			}
		}

		text := st.chip[demo]
		if text == "" {
			text = fmt.Sprintf("focus is on %s", st.focus[demo])
		}

		sc.text("chip-"+demo, text)
	}
}

func (s *navSession) drawCaret(sc *navScreen, m navModel, now time.Time) {
	st := &s.st
	ms := int(now.Sub(st.lastActivity) / time.Millisecond)

	for _, id := range []string{"cr-bar", "cr-block", "cr-under", "cr-autoblock"} {
		d := m.Items[max(m.ByID(id), 0)].Declared
		c := navCaretOf(d)
		text := st.texts[id]
		x, y, w, h := c.Rect(navTextLeft+float64(len([]rune(text)))*liveCharW, navTextTop, liveCharW, navLineH)

		color := c.Color
		if color == "" {
			color = "#1f2937"
		}

		if !c.Visible(ms) {
			color = "transparent"
		}

		ring := ""
		if st.caretFocus == id && st.activeDemo == "caret" {
			ring = "box-shadow:0 0 0 2px #2563eb;"
		}

		sc.style(id, ring)
		sc.text(id+"-t", text)
		sc.style(id+"-c", fmt.Sprintf("left:%gpx;top:%gpx;width:%gpx;height:%gpx;background:%s", x, y, w, h, color))
	}

	f := navCaretOf(m.Items[max(m.ByID(st.caretFocus), 0)].Declared)
	sc.text("chip-caret", fmt.Sprintf("%s: shape %s, %s, caret %s", st.caretFocus, f.Shape, map[bool]string{true: "manual (always on)", false: "blinks every 500ms"}[f.Manual],
		map[bool]string{true: "on", false: "off"}[f.Visible(ms)]))
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func (s *navSession) drawFields(sc *navScreen, m navModel) {
	st := &s.st

	for _, id := range []string{"fs-content", "fs-fixed"} {
		d := m.Items[max(m.ByID(id), 0)].Declared
		text := st.texts[id]
		focused := st.fieldFocus == id && st.activeDemo == "field"
		shown := text

		if focused {
			shown += "|"
		}

		// The width counts the caret, so the content-sized field grows to the
		// right instead of pushing its last character onto a second line.
		w := navFieldWidth(d["field-sizing"], shown, liveCharW, navFieldPad, 40, 150)
		extra := fmt.Sprintf("width:%gpx;white-space:nowrap;overflow:hidden", w)

		shown = navFit(shown, w, focused)

		if focused {
			extra += ";box-shadow:0 0 0 2px #2563eb"
		}

		sc.style(id, extra)
		sc.text(id, shown)
	}

	sc.text("chip-field", fmt.Sprintf("fs-content holds %d characters and is %gpx wide; fs-fixed holds %d and stays %gpx",
		len([]rune(st.texts["fs-content"])), navFieldWidth("content", st.texts["fs-content"], liveCharW, navFieldPad, 40, 150),
		len([]rune(st.texts["fs-fixed"])), 150.0))

	for _, id := range []string{"pw-auto", "pw-none"} {
		d := m.Items[max(m.ByID(id), 0)].Declared
		shown := navShownText("password", d["input-security"], st.texts[id])
		extra := "white-space:nowrap;overflow:hidden"
		focused := st.secFocus == id && st.activeDemo == "security"

		if focused {
			shown += "|"
			extra += ";box-shadow:0 0 0 2px #2563eb"
		}

		shown = navFit(shown, 150, focused)

		sc.style(id, extra)
		sc.text(id, shown)
	}

	sc.text("chip-security", fmt.Sprintf("pw-auto shows %q, pw-none shows %q",
		navShownText("password", "auto", st.texts["pw-auto"]), navShownText("password", "none", st.texts["pw-none"])))
}

func (s *navSession) drawSelect(sc *navScreen, m navModel) {
	for _, id := range []string{"us-text", "us-none", "us-all"} {
		r := s.st.sel[id]
		sc.style(id+"-h", fmt.Sprintf("left:%gpx;width:%gpx", navTextLeft+float64(r[0])*liveCharW, float64(r[1]-r[0])*liveCharW))
		sc.style(id, "")
	}

	text := s.st.chip["select"]
	if text == "" {
		text = "click a word"
	}

	sc.text("chip-select", text)
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func (s *navSession) drawAdjust(sc *navScreen) {
	var parts []string

	for _, c := range []struct{ id, adjust string }{{"ta-none", "none"}, {"ta-120", "120%"}, {"ta-wk", "150%"}, {"ta-auto", "auto"}} {
		size := 14.0
		if s.st.inflate {
			size = navScaleText(c.adjust, 14, 1.5)
		}

		sc.style(c.id, fmt.Sprintf("font-size:%gpx", size))
		parts = append(parts, fmt.Sprintf("%s %gpx", c.id, size))
	}

	state := "off"
	if s.st.inflate {
		state = "on"
	}

	sc.text("chip-adjust", "inflation "+state+": "+strings.Join(parts, ", "))
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func (s *navSession) drawSlider(sc *navScreen, m navModel) {
	var parts []string

	for _, id := range []string{"sl-h", "sl-rtl", "sl-v"} {
		it := m.Items[max(m.ByID(id), 0)]
		orient := navSliderOrient(it.Declared["slider-orientation"], it.W, it.H)
		p := s.st.slider[id]

		var fill, thumb string

		switch orient {
		case "left-to-right":
			fill = fmt.Sprintf("left:0;top:0;width:%gpx;height:%gpx", p*it.W, it.H)
		case "right-to-left":
			fill = fmt.Sprintf("left:%gpx;top:0;width:%gpx;height:%gpx", p*it.W, (1-p)*it.W, it.H)
		case "top-to-bottom":
			fill = fmt.Sprintf("left:0;top:0;width:%gpx;height:%gpx", it.W, p*it.H)
		default:
			fill = fmt.Sprintf("left:0;top:%gpx;width:%gpx;height:%gpx", p*it.H, it.W, (1-p)*it.H)
		}

		if orient == "left-to-right" || orient == "right-to-left" {
			thumb = fmt.Sprintf("left:%gpx;top:-3px;width:10px;height:18px;border-radius:3px", p*(it.W-10))
		} else {
			thumb = fmt.Sprintf("left:-3px;top:%gpx;width:18px;height:10px;border-radius:3px", p*(it.H-10))
		}

		sc.style(id+"-f", fill)
		sc.style(id+"-t", thumb)

		px, py := it.X+p*it.W, it.Y+p*it.H
		parts = append(parts, fmt.Sprintf("%s %.0f", id, navSliderValue(it.Declared["slider-orientation"], it.X, it.Y, it.W, it.H, px, py, 0, 100)))
	}

	sc.text("chip-slider", "values: "+strings.Join(parts, ", "))
}

func (s *navSession) drawDrag(sc *navScreen) {
	sc.style("drag-win", fmt.Sprintf("left:%gpx;top:%gpx", s.st.winX, s.st.winY))

	text := s.st.chip["drag"]
	if text == "" {
		text = "press on the title bar and move"
	}

	sc.text("chip-drag", text)
}

// layerProgress is how far the click animation of one box has run, 0 to 1,
// or -1 when that box is not animating.
func (s *navSession) layerProgress(id string, now time.Time) float64 {
	st := &s.st
	if st.layerID != id {
		return -1
	}

	p := float64(now.Sub(st.layerStart)) / float64(layerMs*time.Millisecond)
	if p >= 1 || p < 0 {
		return -1
	}

	return p
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func (s *navSession) drawLayer(sc *navScreen, m navModel, now time.Time) {
	st := &s.st

	for _, id := range []string{"wc", "wc2"} {
		_, promote := navLayerHints(m.Items[max(m.ByID(id), 0)].Declared["will-change"])
		p := s.layerProgress(id, now)

		if p >= 0 {
			// A box with a cached layer is painted once and then only moved. A
			// box without one is painted again on every frame it moves.
			st.layerFrames[id]++
			st.layerPaints[id] = st.layerFrames[id]

			if promote {
				st.layerPaints[id] = 1
			}

			sc.style(id, fmt.Sprintf("transform:translateX(%gpx);opacity:%g", p*120, 1-0.5*p))
		} else {
			sc.style(id, "")
		}

		sc.text(id+"-badge", layerNote(id, promote, st.layerFrames[id], st.layerPaints[id]))
	}

}

// layerNote is the line beside a will-change box: what the host did with it.
func layerNote(id string, promote bool, frames, paints int) string {
	if frames == 0 {
		return id + ": click the box to move it"
	}

	if promote {
		return fmt.Sprintf("%s: kept on a cached layer. Moved over %d frames, painted %d time", id, frames, paints)
	}

	return fmt.Sprintf("%s: no cached layer. Moved over %d frames, painted %d times (every frame)", id, frames, paints)
}

// navControlStyle is the look the host draws for a control kind.
func navControlStyle(kind string, pressed bool) string {
	switch {
	case kind == "plain" && pressed:
		return "border:0;background:#e5e7eb;border-radius:0"
	case kind == "plain":
		return "border:0;background:transparent;border-radius:0"
	case pressed:
		return "border:2px solid #111827;background:#cbd5e1;border-radius:5px"
	}

	return "border:2px solid #9ca3af;background:#f3f4f6;border-radius:5px"
}

func (s *navSession) drawAppear(sc *navScreen, m navModel) {
	var parts []string

	for _, id := range []string{"ap-none", "ap-btn", "ap-def"} {
		it := m.Items[max(m.ByID(id), 0)]
		kind := navControlKind(it.Tag, "", it.Declared["-webkit-appearance"])
		pressed := s.st.pressed == id
		sc.style(id, navControlStyle(kind, pressed))

		state := "up"
		if pressed {
			state = "pressed"
		}

		parts = append(parts, fmt.Sprintf("%s %s %s", id, kind, state))
	}

	sc.text("chip-appear", strings.Join(parts, ", "))
}

func (s *navSession) drawImage(sc *navScreen, m navModel, now time.Time) {
	st := &s.st
	el := s.imageElapsed(now)

	var parts []string

	for _, id := range []string{"ia-paused", "ia-stopped", "ia-normal"} {
		mode := m.Items[max(m.ByID(id), 0)].Declared["image-animation"]
		pausedAt := 450

		if id == "ia-normal" && st.imgPaused {
			mode, pausedAt = "paused", st.imgPauseAt
		}

		f := navImageFrame(mode, el, pausedAt, imgFrameMs, imgFrames)
		sc.style(id, "background:"+imgColors[f])
		sc.text(id, fmt.Sprint(f+1))
		parts = append(parts, fmt.Sprintf("%s frame %d", id, f+1))
	}

	label := "Pause normal"
	if st.imgPaused {
		label = "Resume normal"
	}

	sc.text("ia-btn", label)
	sc.text("chip-image", strings.Join(parts, ", "))
}

// navFit keeps only as much of text as fits a box w pixels wide, so a long
// value never draws outside its box. A field with focus shows the end of its
// text, which keeps the caret at the right edge as a browser does. A field
// without focus shows the start. The engine does not clip text that runs past
// a box, so the fit is made here instead of by overflow.
func navFit(text string, w float64, focused bool) string {
	room := int((w - navFieldPad) / liveCharW)
	runes := []rune(text)

	if room < 1 || len(runes) <= room {
		return text
	}

	if focused {
		return string(runes[len(runes)-room:])
	}

	return string(runes[:room])
}

// offsetOf is the CSS offset of a sheet scrolled by v: minus v, and a plain 0
// at the start, never "-0".
func offsetOf(v float64) float64 {
	if v == 0 {
		return 0
	}

	return -v
}

// sheetRows is more than the number of vertical scroll positions, so every
// offset maps to its own z-index value.
const sheetRows = 100
