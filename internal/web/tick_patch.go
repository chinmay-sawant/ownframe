package web

import (
	"image"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// tickPatch is one id-anchored in-place update. Text replaces the element's
// textContent; Style replaces its inline style attribute. Both mirror the
// executed source form so a later full /dom morph converges to the same
// pixels. Only one of Text or Style is set per patch.
type tickPatch struct {
	ID    string  `json:"id"`
	Text  *string `json:"text,omitempty"`
	Style *string `json:"style,omitempty"`
}

// maxTickPatches caps one event. Steady in-place ticks emit a handful of
// id updates (clock, bar, a few flap tiles); past the cap the server omits
// patches and the client falls back to its current behavior.
const maxTickPatches = 128

// collectTickPatches resolves dirty rects to id-anchored text and fill-width
// patches. Only leaf elements are patched: replacing textContent or inline
// style on a container would destroy the child nodes the full document
// morph preserves. Leafness comes from two independent signals, and a box
// is skipped when either marks it a container: the layout parent index,
// and the executed-source nesting (an id whose element contains child
// elements). Boxes are walked deepest-first so a child claims its ops
// before a container can. Style patches additionally require the element to
// already carry an inline width in the executed source, so a fill that
// merely overlaps a leaf box can never restyle unrelated layout. Control
// ids are never patched. Only values that changed since the last event are
// emitted; cache records the last sent text and style per id.
func collectTickPatches(display *layout.Display, boxes []layout.Box, rects []image.Rectangle, cache map[string][2]string, styleWidthIDs, controlIDs, containers map[string]bool) []tickPatch {
	if display == nil || len(rects) == 0 {
		return nil
	}
	byOp := make(map[uint64]int, len(display.Ops))
	for i := range display.Ops {
		if id := display.Ops[i].ID; id != 0 {
			if _, ok := byOp[id]; !ok {
				byOp[id] = i
			}
		}
	}
	hasKids := make(map[int]bool, len(boxes))
	for i := range boxes {
		if p := boxes[i].ParentIndex; p >= 0 && p != i {
			hasKids[p] = true
		}
	}
	claimed := make(map[int]bool, len(display.Ops))
	var out []tickPatch
	for b := len(boxes) - 1; b >= 0; b-- {
		box := boxes[b]
		if box.ID == "" || !validPatchID(box.ID) {
			continue
		}
		if hasKids[b] {
			continue
		}
		if containers[box.ID] {
			continue
		}
		if controlIDs[box.ID] {
			continue
		}
		if !boxOverlapsDirty(box, rects) {
			continue
		}
		for _, opID := range box.OperationIDs {
			i, ok := byOp[opID]
			if !ok || claimed[i] {
				continue
			}
			claimed[i] = true
			op := &display.Ops[i]
			switch op.Kind {
			case layout.DisplayOpText:
				if cache != nil {
					if got := cache[box.ID]; got[0] == op.Text {
						continue
					}
				}
				if len(op.Text) > maxChangeValue {
					return nil
				}
				text := op.Text
				out = append(out, tickPatch{ID: box.ID, Text: &text})
				if cache != nil {
					got := cache[box.ID]
					got[0] = op.Text
					cache[box.ID] = got
				}
			case layout.DisplayOpFillRect:
				if box.W <= 0 || !styleWidthIDs[box.ID] {
					continue
				}
				style := fillWidthStyle(op.W, box.W)
				if cache != nil {
					if got := cache[box.ID]; got[1] == style {
						continue
					}
				}
				if len(style) > maxChangeValue {
					return nil
				}
				out = append(out, tickPatch{ID: box.ID, Style: &style})
				if cache != nil {
					got := cache[box.ID]
					got[1] = style
					cache[box.ID] = got
				}
			}
			if len(out) >= maxTickPatches {
				return nil
			}
		}
	}
	return out
}

// containerIDs returns the ids whose executed-source element contains a
// child element. Tick patches skip these even when the layout parent index
// does not link them: HTML nesting, not box parenting, decides what
// textContent replacement would destroy. Void elements never open a scope.
func containerIDs(source string) map[string]bool {
	out := map[string]bool{}
	var stack []string
	i := 0
	for i < len(source) {
		if source[i] != '<' {
			next := strings.IndexByte(source[i:], '<')
			if next < 0 {
				return out
			}
			i += next
			continue
		}
		tag, ok := parseTag(source, i)
		if !ok || tag.end <= i {
			i++
			continue
		}
		switch {
		case tag.comment || tag.other:
		case tag.closing:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		default:
			for _, id := range stack {
				if id != "" {
					out[id] = true
				}
			}
			if tag.selfClose || voidElement(tag.name) {
				break
			}
			stack = append(stack, attrValue(tag.attrs, "id"))
		}
		i = tag.end
	}
	return out
}

// voidElement reports the HTML void elements, which cannot contain children
// whether or not the source writes them self-closed.
func voidElement(name string) bool {
	switch name {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	default:
		return false
	}
}

// inlineWidthIDs returns the ids whose executed-source tag carries an inline
// width declaration, so tick style patches only ever override a width the
// template itself sets. It reuses the document tag scanner; a tag the
// scanner cannot parse is skipped, never patched.
func inlineWidthIDs(source string) map[string]bool {
	out := map[string]bool{}
	i := 0
	for i < len(source) {
		if source[i] != '<' {
			next := strings.IndexByte(source[i:], '<')
			if next < 0 {
				return out
			}
			i += next
			continue
		}
		tag, ok := parseTag(source, i)
		if !ok || tag.end <= i {
			i++
			continue
		}
		if !tag.closing && !tag.selfClose && tag.name != "" {
			id := attrValue(tag.attrs, "id")
			if id != "" {
				for _, a := range tag.attrs {
					if a.name == "style" && strings.Contains(strings.ToLower(a.value), "width") {
						out[id] = true
						break
					}
				}
			}
		}
		i = tag.end
	}
	return out
}

// executedSourceOf returns the page's last executed template source, or ""
// when the screen does not expose one.
func executedSourceOf(app any) string {
	if src, ok := app.(domSource); ok {
		return src.ExecutedSource()
	}
	return ""
}

// controlIDs returns the ids of live form controls, which tick patches must
// never rewrite.
func controlIDs(controls []page.WebControl) map[string]bool {
	if len(controls) == 0 {
		return nil
	}
	out := make(map[string]bool, len(controls))
	for _, c := range controls {
		if c.ID != "" {
			out[c.ID] = true
		}
	}
	return out
}

// boxOverlapsDirty reports whether the box intersects any dirty rect.
func boxOverlapsDirty(box layout.Box, rects []image.Rectangle) bool {
	x0 := int(box.X)
	y0 := int(box.Y)
	x1 := int(box.X + box.W + 0.5)
	y1 := int(box.Y + box.H + 0.5)
	for _, r := range rects {
		if r.Empty() {
			continue
		}
		if x0 < r.Max.X && x1 > r.Min.X && y0 < r.Max.Y && y1 > r.Min.Y {
			return true
		}
	}
	return false
}

// fillWidthStyle renders a fill width as the inline style the template
// prints (style="width:N%"). Out-of-range widths clamp to 0-100.
func fillWidthStyle(fillW, boxW float64) string {
	fraction := fillW / boxW
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return "width:" + percent1(fraction*100) + "%"
}

// percent1 formats one decimal place.
func percent1(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}

// validPatchID mirrors the /change id validator: bounded ids with no
// markup, quote, or control characters.
func validPatchID(id string) bool {
	if id == "" || len(id) > maxDOMIDLength {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c == '<' || c == '>' || c == '"' || c == '\'' || c == '\r' || c == '\n' || c == 0 {
			return false
		}
	}
	return true
}
