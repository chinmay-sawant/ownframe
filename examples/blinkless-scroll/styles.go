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

func baseWidth(style string) float64 {
	if m := widthRe.FindStringSubmatch(style); m != nil {
		if w, err := strconv.ParseFloat(m[1], 64); err == nil {
			return w
		}
	}

	return 600
}

// Colors a trigger or interest card changes to when it turns on.
const (
	colorOn       = "#bbf7d0"
	colorInterest = "#fde68a"
	// The box the latest click landed on gets this outline and fill.
	colorClick     = "#2563eb"
	colorClickFill = "#bfdbfe"
)

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

	for key, val := range f.Timelines {
		n := h.Nodes[key.Node]

		switch val.Kind {
		case KindScroll:
			w := baseWidth(base[n.ID]) * val.Progress
			extra[key.Node] = append(extra[key.Node], fmt.Sprintf("width:%.1fpx", w))
		case KindView:
			subject := key.Node
			for _, tl := range n.Timelines {
				if j, ok := h.byID[tl.SubjectID]; ok && tl.Name == key.Name && tl.SubjectID != "" {
					subject = j
				}
			}

			e := val.Ranges["entry"]
			extra[subject] = append(extra[subject],
				fmt.Sprintf("opacity:%.2f", 0.25+0.75*e),
				fmt.Sprintf("margin-left:%.1fpx", (1-e)*60))
		case KindPointer:
			dir := "90deg"
			for _, tl := range n.Timelines {
				if tl.Name == key.Name && tl.Axis == AxisY {
					dir = "180deg"
				}
			}

			p := val.Progress * 100
			extra[key.Node] = append(extra[key.Node],
				fmt.Sprintf("background:linear-gradient(%s,#fca5a5 %.0f%%,#e8e2d4 %.0f%%)", dir, p, p))
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
