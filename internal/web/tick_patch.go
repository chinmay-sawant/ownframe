package web

import (
	"html"
	"image"
	"sort"
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

type tickTextPart struct {
	text string
	x    float64
	y    float64
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
// merely overlaps a leaf box can never restyle unrelated layout, and the
// percent resolves against the smallest strictly-containing source element
// with children (op points converted through Display.PixelPerPoint), never a
// sibling segment that overlaps a zero-width fill. With no container the
// style patch is skipped. Control
// ids are never patched. Only values that changed since the last event are
// emitted; cache records the last sent text and style per id.
func collectTickPatches(display *layout.Display, boxes []layout.Box, rects []image.Rectangle, cache map[string][2]string, styleWidthIDs, controlIDs, containers map[string]bool, sourceText map[string]string) []tickPatch {
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
	var touched []string
	// rollback drops the ids this call already recorded so a breached
	// event never poisons the cache with values the client never saw.
	rollback := func() []tickPatch {
		if cache != nil {
			for _, id := range touched {
				delete(cache, id)
			}
		}
		return nil
	}
	var out []tickPatch
	textParts := map[string][]tickTextPart{}
	textOrder := make([]string, 0)
	textDirty := map[string]bool{}
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
		boxDirty := boxOverlapsDirty(box, rects)
		for _, opID := range box.OperationIDs {
			i, ok := byOp[opID]
			if !ok || claimed[i] {
				continue
			}
			op := &display.Ops[i]
			dirty := boxDirty || displayOpOverlapsDirty(*op, display.PixelPerPoint, rects)
			claimed[i] = true
			if op.Kind == layout.DisplayOpText {
				if _, exists := textParts[box.ID]; !exists {
					textOrder = append(textOrder, box.ID)
				}
				textParts[box.ID] = append(textParts[box.ID], tickTextPart{text: op.Text, x: op.X, y: op.Y})
				if dirty {
					textDirty[box.ID] = true
				}
				continue
			}
			if !dirty {
				continue
			}
			switch op.Kind {
			case layout.DisplayOpFillRect:
				if !styleWidthIDs[box.ID] {
					continue
				}
				containerW, ok := tickContainerWidth(boxes, b, containers)
				if !ok || containerW <= 0 {
					continue
				}
				style := fillWidthStyle(op.W*display.PixelPerPoint, containerW)
				if cache != nil {
					if got := cache[box.ID]; got[1] == style {
						continue
					}
				}
				if len(style) > maxChangeValue {
					return rollback()
				}
				out = append(out, tickPatch{ID: box.ID, Style: &style})
				if cache != nil {
					got := cache[box.ID]
					got[1] = style
					cache[box.ID] = got
					touched = append(touched, box.ID)
				}
			}
			if len(out) >= maxTickPatches {
				return rollback()
			}
		}
	}
	for _, id := range textOrder {
		if !textDirty[id] {
			continue
		}
		text := mergeTickText(textParts[id], sourceText[id], hasSourceText(sourceText, id))
		if cache != nil {
			previous, cached := cache[id]
			if cached && previous[0] == text {
				continue
			}
			if !cached && hasSourceText(sourceText, id) && sourceText[id] == text {
				continue
			}
		}
		if len(text) > maxChangeValue {
			return rollback()
		}
		out = append(out, tickPatch{ID: id, Text: &text})
		if cache != nil {
			got := cache[id]
			got[0] = text
			cache[id] = got
			touched = append(touched, id)
		}
		if len(out) >= maxTickPatches {
			return rollback()
		}
	}
	return out
}

func hasSourceText(sourceText map[string]string, id string) bool {
	_, ok := sourceText[id]
	return ok
}

func mergeTickText(parts []tickTextPart, source string, hasSource bool) string {
	if hasSource && sourceMatchesTickText(source, parts) {
		return source
	}

	sort.SliceStable(parts, func(i, j int) bool {
		if parts[i].y == parts[j].y {
			return parts[i].x < parts[j].x
		}
		return parts[i].y < parts[j].y
	})
	var text strings.Builder
	var previousY float64
	for i, part := range parts {
		if i > 0 && part.y != previousY && text.Len() > 0 &&
			!endsWithSpace(text.String()) && !startsWithSpace(part.text) {
			text.WriteByte(' ')
		}
		text.WriteString(part.text)
		previousY = part.y
	}
	return text.String()
}

func sourceMatchesTickText(source string, parts []tickTextPart) bool {
	parts = append([]tickTextPart(nil), parts...)
	sort.SliceStable(parts, func(i, j int) bool {
		if parts[i].y == parts[j].y {
			return parts[i].x < parts[j].x
		}
		return parts[i].y < parts[j].y
	})
	position := 0
	hasText := false
	for _, part := range parts {
		if part.text == "" {
			continue
		}
		hasText = true
		at := strings.Index(source[position:], part.text)
		if at < 0 {
			return false
		}
		position += at + len(part.text)
	}
	return hasText || source == ""
}

func startsWithSpace(text string) bool {
	return len(text) > 0 && (text[0] == ' ' || text[0] == '\n' || text[0] == '\t')
}

func endsWithSpace(text string) bool {
	if len(text) == 0 {
		return false
	}
	last := text[len(text)-1]
	return last == ' ' || last == '\n' || last == '\t'
}

func sourceTextByID(source string, containers map[string]bool) map[string]string {
	texts := map[string]string{}
	for i := 0; i < len(source); {
		if source[i] != '<' {
			i++
			continue
		}
		tag, ok := parseTag(source, i)
		if !ok || tag.end <= i {
			i++
			continue
		}
		id := attrValue(tag.attrs, "id")
		if tag.closing || tag.comment || tag.other || tag.selfClose || voidElement(tag.name) ||
			id == "" || containers[id] {
			i = tag.end
			continue
		}
		if text, ok := leafText(source, tag); ok {
			texts[id] = html.UnescapeString(text)
		}
		i = tag.end
	}
	return texts
}

func leafText(source string, opening htmlTag) (string, bool) {
	for i := opening.end; i < len(source); {
		if source[i] != '<' {
			i++
			continue
		}
		tag, ok := parseTag(source, i)
		if !ok || tag.end <= i {
			i++
			continue
		}
		if tag.comment || tag.other {
			i = tag.end
			continue
		}
		if tag.closing {
			if tag.name == opening.name {
				return source[opening.end:i], true
			}
			return "", false
		}
		return "", false
	}
	return "", false
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

// displayOpOverlapsDirty checks the current retained geometry. A box can
// keep its committed width while a retained fill operation grows from zero.
func displayOpOverlapsDirty(op layout.DisplayOp, pixelPerPoint float64, rects []image.Rectangle) bool {
	if pixelPerPoint <= 0 {
		pixelPerPoint = 1
	}
	return boxOverlapsDirty(layout.Box{
		X: op.X * pixelPerPoint,
		Y: op.Y * pixelPerPoint,
		W: op.W * pixelPerPoint,
		H: op.H * pixelPerPoint,
	}, rects)
}

// fillWidthStyle renders a fill width in CSS pixels as the inline style the
// template prints (style="width:N%") relative to its container width in CSS
// pixels. The caller converts the op width from layout points through
// Display.PixelPerPoint and resolves the container with tickContainerWidth;
// a fill must never divide by its own leaf box, which hugs the fill.
// Out-of-range widths clamp to 0-100.
func fillWidthStyle(fillPx, containerPx float64) string {
	fraction := fillPx / containerPx
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return "width:" + percent1(fraction*100) + "%"
}

// tickContainerWidth returns the width of the smallest source container that
// geometrically contains boxes[leaf] and is strictly wider than it. Boxes
// are in CSS pixels, so no unit conversion applies here; the caller converts
// the op width through Display.PixelPerPoint before dividing by the result.
func tickContainerWidth(boxes []layout.Box, leaf int, containers map[string]bool) (float64, bool) {
	leafBox := boxes[leaf]
	best := -1
	for i := range boxes {
		if i == leaf {
			continue
		}
		c := boxes[i]
		if c.ID == "" || !containers[c.ID] {
			continue
		}
		if c.W <= leafBox.W {
			continue
		}
		if c.X <= leafBox.X && c.Y <= leafBox.Y &&
			c.X+c.W >= leafBox.X+leafBox.W &&
			c.Y+c.H >= leafBox.Y+leafBox.H {
			if best < 0 || c.W < boxes[best].W {
				best = i
			}
		}
	}
	if best < 0 {
		return 0, false
	}
	return boxes[best].W, true
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
