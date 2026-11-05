package click

import (
	"context"
	"html"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

const maxLabelWalk = 32

// labelFocuser is implemented by the page. A surface that does not
// implement it keeps the click sequence unchanged.
type labelFocuser interface {
	LabelControl(box layout.Box) string
	FocusControl(ctx context.Context, id string) error
}

func focusFromLabel(ctx context.Context, s Surface, hit layout.Box) (bool, error) {
	finder, ok := s.(labelFocuser)
	if !ok {
		return false, nil
	}

	id := finder.LabelControl(hit)
	if id == "" {
		return false, nil
	}

	return true, finder.FocusControl(ctx, id)
}

// LabelControl returns the control id named by the label under hit.
// A label with no for attribute returns "".
func LabelControl(source string, boxes []layout.Box, hit layout.Box) string {
	label, ok := labelBox(boxes, hit)
	if !ok {
		return ""
	}

	return LabelFor(source, label.ID, label.Text)
}

// LabelFor returns the for value of the label identified by id, or by its
// visible text when id is empty.
func LabelFor(source, id, text string) string {
	links := labelLinks(source)
	if id != "" {
		for _, link := range links {
			if link.id == id && link.forID != "" {
				return link.forID
			}
		}
	}

	folded := foldSpace(text)
	if folded == "" {
		return ""
	}

	for _, link := range links {
		if link.forID != "" && foldSpace(link.text) == folded {
			return link.forID
		}
	}

	return ""
}

type labelLink struct {
	id    string
	forID string
	text  string
}

func labelBox(boxes []layout.Box, hit layout.Box) (layout.Box, bool) {
	current := hit
	for range maxLabelWalk {
		if strings.EqualFold(current.Tag, "label") {
			return current, true
		}

		parent := current.ParentIndex
		if parent < 0 || parent >= len(boxes) {
			return layout.Box{}, false
		}

		current = boxes[parent]
	}

	return layout.Box{}, false
}

func labelLinks(source string) []labelLink {
	links := []labelLink{}
	i := 0
	for i < len(source) {
		next := strings.IndexByte(source[i:], '<')
		if next < 0 {
			break
		}

		i += next
		if strings.HasPrefix(source[i:], "<!--") {
			end := strings.Index(source[i+form.CommentOpenLen:], "-->")
			if end < 0 {
				break
			}

			i += end + form.CommentTotalLen

			continue
		}

		name, raw, end, ok := form.ReadOpen(source, i)
		if !ok {
			i++

			continue
		}

		switch strings.ToLower(name) {
		case "script", "style":
			_, i = form.CloseSpan(source, end, strings.ToLower(name))
		case "label":
			closeAt, after := form.CloseSpan(source, end, "label")
			if closeAt < end {
				closeAt = end
			}

			links = append(links, labelLink{
				id:    form.AttrValue(raw, "id"),
				forID: form.AttrValue(raw, "for"),
				text:  labelText(source[end:closeAt]),
			})
			i = after
		default:
			i = end
		}
	}

	return links
}

func labelText(fragment string) string {
	var b strings.Builder
	i := 0
	for i < len(fragment) {
		if fragment[i] != '<' {
			b.WriteByte(fragment[i])
			i++

			continue
		}

		_, _, end, ok := form.ReadOpen(fragment, i)
		if !ok {
			rel := strings.IndexByte(fragment[i:], '>')
			if rel < 0 {
				break
			}

			i += rel + 1

			continue
		}

		i = end
	}

	return html.UnescapeString(b.String())
}

func foldSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
