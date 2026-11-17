package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The page keeps its look in inline style attributes. Patching the style
// attribute replaces all of it, so the host keeps each driven element's
// original style text and appends its own declarations after it. A later
// declaration wins, so the original look survives and the driven values
// override only what they name.

var (
	tagRe   = regexp.MustCompile(`(?is)<[a-z][^>]*\bid="([^"]+)"[^>]*>`)
	styleRe = regexp.MustCompile(`(?is)\bstyle="([^"]*)"`)
	widthRe = regexp.MustCompile(`(?i)(?:^|;)\s*width\s*:\s*([0-9.]+)px`)
)

// BaseStyles maps element id to its inline style text.
func BaseStyles(html string) map[string]string {
	out := map[string]string{}

	for _, m := range tagRe.FindAllStringSubmatch(html, -1) {
		if s := styleRe.FindStringSubmatch(m[0]); s != nil {
			out[m[1]] = strings.TrimRight(strings.TrimSpace(s[1]), ";")
		}
	}

	return out
}

// Fallback width in CSS pixels for a box with no width in its style text, and
// the opacity and slide the view-timeline cards use as they enter.
const (
	defaultBaseWidthPx = 600
	cardOpacityStart   = 0.25
	cardOpacityRange   = 0.75
	cardSlideOffsetPx  = 60
)

func baseWidth(style string) float64 {
	if m := widthRe.FindStringSubmatch(style); m != nil {
		if w, err := strconv.ParseFloat(m[1], 64); err == nil {
			return w
		}
	}

	return defaultBaseWidthPx
}

// Colors a trigger or interest card changes to when it turns on.
const (
	colorOn       = "#bbf7d0"
	colorInterest = "#fde68a"
	// The box the latest click landed on gets this outline and fill.
	colorClick     = "#2563eb"
	colorClickFill = "#bfdbfe"
)

// Fills for a pointer box and for a box that scrolls on its own.
const (
	colorPointer = "#fca5a5"
	colorNested  = "#86efac"
	colorView    = "#93c5fd"
	colorTrack   = "#e8e2d4"
)

// progressFill is a background that is filled up to progress along dir.
func progressFill(dir, fill string, progress float64) string {
	p := progress * percentScale

	return fmt.Sprintf("background:linear-gradient(%s,%s %.0f%%,%s %.0f%%)", dir, fill, p, colorTrack, p)
}

// StylePatch is one element's new style attribute.
type StylePatch struct{ ID, Style string }

// Styles turns a frame into style patches for the elements it drives:
//
//   - a scroll timeline owner gets its width set to progress * its own width
//     (scroll-bar, scroll-bar-longhand)
//   - a view timeline subject fades and slides in with its entry progress
//     (view-card-1, view-card-2, view-card-3)
//   - a pointer timeline owner fills a gradient up to the progress
//     (pointer-box, pointer-box-longhand)
//   - a trigger owner turns green while its trigger is on (trigger-box,
//     trigger-longhand, trigger-range-pair, event-box, event-box-longhand)
//   - an interest owner turns amber while interest holds (interest-card,
//     interest-card-longhand)
//
// Only ids that exist in the host come back. The result is sorted by node
// order so tests can compare it.
func (h *Host) Styles(f Frame, base map[string]string) []StylePatch {
	extra := map[int][]string{}

	// The clicked box goes first, so a trigger, event or interest colour
	// declared after it still wins the background and only the outline stays.
	// Only a box that declares interactivity itself gets it. Any other box,
	// a button for example, is left alone.
	if h.hasClicked && h.clicked < len(h.Nodes) && strings.Contains(base[h.Nodes[h.clicked].ID], "interactivity") {
		extra[h.clicked] = append(extra[h.clicked], "outline:3px solid "+colorClick, "background:"+colorClickFill)
	}

	// A box can declare a list of timelines. Only the first of each kind is
	// drawn, in the order the names are written, so the result does not
	// depend on map order. Every timeline is still read and logged. An
	// inactive timeline has no scroller to read and draws nothing.
	for node := range h.Nodes {
		seen := map[string]bool{}

		for _, tl := range h.Nodes[node].Timelines {
			val, ok := f.Timelines[tlKey{node, tl.Name}]
			if !ok || val.Inactive || seen[tl.Kind] {
				continue
			}

			seen[tl.Kind] = true

			switch tl.Kind {
			case KindScroll:
				if tl.ScrollerID != "" {
					// A box that scrolls on its own would vanish if its width
					// followed its progress, so its background fills instead.
					extra[node] = append(extra[node], progressFill("90deg", colorNested, val.Progress))

					continue
				}

				// A progress bar fixed to the window runs to the far edge of the
				// window, whatever its size. Any other box keeps the width its
				// own style gives it.
				full := baseWidth(base[h.Nodes[node].ID])
				if strings.Contains(base[h.Nodes[node].ID], "position:fixed") {
					full = h.viewW
				}

				extra[node] = append(extra[node], fmt.Sprintf("width:%.1fpx", full*val.Progress))
			case KindView:
				subject := node
				if j, found := h.byID[tl.SubjectID]; found && tl.SubjectID != "" {
					subject = j
				}

				// The fill is how far the card has crossed the window, cover
				// progress: empty as it enters, full as it leaves. The fade and
				// the slide show only the entry part of that trip.
				e := val.Ranges["entry"]
				extra[subject] = append(extra[subject],
					progressFill("90deg", colorView, val.Progress),
					fmt.Sprintf("opacity:%.2f", cardOpacityStart+cardOpacityRange*e),
					fmt.Sprintf("margin-left:%.1fpx", (1-e)*cardSlideOffsetPx))
			case KindPointer:
				dir := "90deg"
				if tl.Axis == AxisY {
					dir = "180deg"
				}

				extra[node] = append(extra[node], progressFill(dir, colorPointer, val.Progress))
			}
		}
	}

	// A scroll container the host scrolled shows it: the box named
	// "<id>-content" slides up by the offset, and the container clips it.
	for id, off := range h.scrollers {
		if idx, ok := h.byID[id+"-content"]; ok && off[1] != 0 {
			extra[idx] = append(extra[idx], fmt.Sprintf("margin-top:%.0fpx", -off[1]))
		}
	}

	for i, on := range f.TriggerOn {
		if on {
			extra[h.Triggers[i].Owner] = append(extra[h.Triggers[i].Owner], "background:"+colorOn)
		}
	}

	for i, on := range f.EventOn {
		if on {
			extra[h.Events[i].Owner] = append(extra[h.Events[i].Owner], "background:"+colorOn)
		}
	}

	for idx, on := range f.InterestOn {
		if on {
			extra[idx] = append(extra[idx], "background:"+colorInterest)
		}
	}

	var out []StylePatch

	for i, n := range h.Nodes {
		decl, ok := extra[i]
		if !ok || n.ID == "" {
			continue
		}

		style := base[n.ID]
		if style != "" {
			style += ";"
		}

		out = append(out, StylePatch{ID: n.ID, Style: style + strings.Join(decl, ";")})
	}

	return out
}
