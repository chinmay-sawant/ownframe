package main

import "context"

// fragTab is the "Fragmentation and generated content" tab. The engine paints
// no page, footnote area or outline, so the host reads what the engine hands
// over (layout.Outline, layout.CopiedFragments, Box.Declared) and shows it.
// block-ellipsis is the one the engine acts on in the drawing list.
func fragTab() Tab {
	return Tab{
		ID:    "frag",
		Title: "Fragmentation and generated content",
		Props: []Prop{
			{
				Name:   "bookmark-label",
				Demo:   fragDemo(fragOutlineHTML),
				Covers: []string{"bookmark-level", "bookmark-state"},
				Note: "The engine lists each element with a bookmark-level as an outline entry " +
					"(label from text, a string or attr(), open or closed); the host prints the list below.",
			},
			{
				Name: "copy-into",
				Demo: fragDemo(fragRunningHeaderHTML),
				Note: "The engine hands the host the copied text under each name; the host shows it as a running header. " +
					"The heading itself still paints in the page.",
			},
			{
				Name: "block-ellipsis",
				Demo: fragEllipsisHTML,
				Note: "A paragraph cut by line-clamp:2 ends with the marker: the default, a string, or nothing.",
			},
			{
				Name:   "footnote-display",
				Demo:   fragDemo(fragFootnoteHTML),
				Covers: []string{"footnote-policy"},
				Note: "Stored and exported on the box and printed below. The engine has no footnote area, " +
					"so the notes do not move.",
			},
		},
	}
}

// fragDemo wraps a host helper that can fail into a Demo.
func fragDemo(fn func(context.Context) (string, error)) func() string {
	return func() string {
		out, err := fn(context.Background())
		if err != nil {
			return "<p>" + err.Error() + "</p>"
		}

		return out
	}
}
