package main

import (
	"fmt"
	"strings"
	"unicode"
)

// navCaret is a caret as the host draws it.
type navCaret struct {
	// Shape is bar, block or underscore. auto resolves to bar.
	Shape string
	// Manual is true when the author drives the animation: no blink.
	Manual bool
	// Color is the declared color text, or "" for the text color.
	Color string
}

// navCaretOf reads the caret longhands and the caret shorthand from one
// element's declared map. A longhand wins over the same part of the shorthand.
func navCaretOf(d map[string]string) navCaret {
	c := navCaret{Shape: "bar"}

	for _, f := range strings.Fields(d["caret"]) {
		switch f {
		case "manual":
			c.Manual = true
		case "bar", "block", "underscore":
			c.Shape = f
		case "auto":
		default:
			c.Color = f
		}
	}

	switch d["caret-shape"] {
	case "bar", "block", "underscore":
		c.Shape = d["caret-shape"]
	case "auto":
		c.Shape = "bar"
	}

	switch d["caret-animation"] {
	case "manual":
		c.Manual = true
	case "auto":
		c.Manual = false
	}

	return c
}

// caretBlinkMs is the half period of the automatic blink.
const caretBlinkMs = 500

// caretBarWidth is the width of a bar caret and the height of an underscore.
const caretBarWidth = 2

// Rect gives the caret rectangle for a character cell at (x, y) that is
// cellW wide and cellH tall.
func (c navCaret) Rect(x, y, cellW, cellH float64) (rx, ry, rw, rh float64) {
	switch c.Shape {
	case "block":
		return x, y, cellW, cellH
	case "underscore":
		return x, y + cellH - caretBarWidth, cellW, caretBarWidth
	}

	return x, y, caretBarWidth, cellH
}

// Visible says whether the caret is drawn at ms milliseconds after the last
// key press. The automatic caret blinks every caretBlinkMs; a manual one stays
// on, because the author owns the animation.
func (c navCaret) Visible(ms int) bool {
	if c.Manual || ms < 0 {
		return true
	}

	return (ms/caretBlinkMs)%2 == 0
}

// navFieldWidth is the width of a text field. field-sizing: content grows the
// field with its text (never below minW); fixed, and a field with no
// declaration, keep fixedW.
func navFieldWidth(sizing, text string, charW, padding, minW, fixedW float64) float64 {
	if sizing != "content" {
		return fixedW
	}

	w := float64(len([]rune(text)))*charW + padding
	if w < minW {
		return minW
	}

	return w
}

// navShownText is what a text field shows for its value. A password field
// hides its characters unless input-security is none.
func navShownText(inputType, security, value string) string {
	if inputType == "password" && security != "none" {
		return strings.Repeat("•", len([]rune(value)))
	}

	return value
}

// navSelect is the range a click at rune index pos selects in text, given the
// -webkit-user-select value: none selects nothing, all selects everything,
// auto and text select the word under the pointer.
func navSelect(mode, text string, pos int) (lo, hi int) {
	runes := []rune(text)
	if pos < 0 || pos > len(runes) {
		pos = len(runes)
	}

	switch mode {
	case "none":
		return pos, pos
	case "all":
		return 0, len(runes)
	}

	lo, hi = pos, pos
	for lo > 0 && isWordRune(runes[lo-1]) {
		lo--
	}

	for hi < len(runes) && isWordRune(runes[hi]) {
		hi++
	}

	return lo, hi
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

// navScaleText is the font size after text-size-adjust. none keeps base, a
// percentage scales it, and auto uses hostAuto, the factor this host picks
// (1 means no inflation).
func navScaleText(adjust string, base, hostAuto float64) float64 {
	switch {
	case adjust == "none":
		return base
	case strings.HasSuffix(adjust, "%"):
		var pct float64
		if _, err := fmt.Sscan(strings.TrimSuffix(adjust, "%"), &pct); err == nil {
			return base * pct / navPercent
		}

		return base
	}

	return base * hostAuto
}

// navTextAdjustOf picks the standard property over the -webkit- one.
func navTextAdjustOf(d map[string]string) string {
	if v := d["text-size-adjust"]; v != "" {
		return v
	}

	return d["-webkit-text-size-adjust"]
}

// navSliderValue maps a pointer at (px, py) over a slider track to a value in
// [lo, hi]. auto is left-to-right on a wide track and bottom-to-top on a tall
// one.
func navSliderValue(orient string, x, y, w, h, px, py, lo, hi float64) float64 {
	if orient == "" || orient == "auto" {
		orient = "left-to-right"
		if h > w {
			orient = "bottom-to-top"
		}
	}

	var t float64

	switch orient {
	case "left-to-right":
		t = (px - x) / w
	case "right-to-left":
		t = 1 - (px-x)/w
	case "top-to-bottom":
		t = (py - y) / h
	case "bottom-to-top":
		t = 1 - (py-y)/h
	}

	t = min(max(t, 0), 1)

	return lo + t*(hi-lo)
}

// navDragRegion says whether a press at item i starts a window drag. The
// nearest ancestor-or-self that declares window-drag decides: move drags, none
// does not.
func (m navModel) navDragRegion(i int) bool {
	v, _ := m.declaredUp(i, "window-drag")

	return v == "move"
}

// navLayerHints lists the properties will-change names and whether the host
// should keep a cached layer for them. auto and no declaration mean no layer.
func navLayerHints(willChange string) (props []string, promote bool) {
	if willChange == "" || willChange == "auto" {
		return nil, false
	}

	for _, p := range strings.Split(willChange, ",") {
		p = strings.TrimSpace(p)
		props = append(props, p)

		switch p {
		case "transform", "opacity", "filter", "backdrop-filter", "scroll-position":
			promote = true
		}
	}

	return props, promote
}

// navControlKind is the control the host draws for an element. A -webkit-
// appearance of none turns a native control into a plain box.
func navControlKind(tag, inputType, appearance string) string {
	if appearance == "none" {
		return "plain"
	}

	if appearance != "" && appearance != "auto" {
		return appearance
	}

	switch {
	case tag == "button":
		return "button"
	case tag == "input" && (inputType == "checkbox" || inputType == "radio"):
		return inputType
	case tag == "input" || tag == "textarea":
		return "textfield"
	}

	return "plain"
}

// navImageFrame picks the frame of an animated image to draw. normal and
// running play, stopped shows frame 0, paused freezes at pausedAtMs.
func navImageFrame(mode string, elapsedMs, pausedAtMs, frameMs, frames int) int {
	if frames <= 1 || frameMs <= 0 {
		return 0
	}

	switch mode {
	case "stopped":
		return 0
	case "paused":
		elapsedMs = pausedAtMs
	}

	return (elapsedMs / frameMs) % frames
}

// navPercent turns a percentage into a fraction.
const navPercent = 100
