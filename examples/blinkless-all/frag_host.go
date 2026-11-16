package main

import (
	"context"
	"fmt"
	"html"
	"strings"

	bhtml "github.com/chinmay-sawant/blinkless/html"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
)

// The engine keeps bookmark, copy-into and footnote values and hands them to
// the host. The host here builds an outline list, a running header and a list
// of footnote settings from them. block-ellipsis is different: the engine
// itself ends a clamped paragraph with the marker.

const fragBookmarkPage = `<!DOCTYPE html><html><body>
<h1 style="bookmark-level:1">Guide</h1>
<h2 style="bookmark-level:2">Install</h2>
<h2 style="bookmark-level:2; bookmark-label:'Usage: ' attr(data-topic); bookmark-state:closed" data-topic="CLI">Using it</h2>
<h3 style="bookmark-level:3">Flags</h3>
<h2 style="bookmark-level:none">Skipped by level</h2>
<h1 style="bookmark-level:1">Reference</h1>
</body></html>`

const fragCopyPage = `<!DOCTYPE html><html><body>
<h1 style="copy-into: chapter-title content, chapter-id attr(data-id)" data-id="ch-7">Chapter Seven</h1>
<p>Body text.</p><h2 style="copy-into: none">No copy</h2>
</body></html>`

const fragFootnotePage = `<!DOCTYPE html><html><body>
<p id="fn-note" style="footnote-display: compact; footnote-policy: block">A note.</p>
<p id="fn-default">Another.</p>
</body></html>`

// fragDocument parses and styles one sample page.
func fragDocument(ctx context.Context, page string) (*css.Document, error) {
	doc, err := bhtml.Parse([]byte(page))
	if err != nil {
		return nil, err
	}

	return css.Apply(ctx, doc, css.Options{WidthPx: fragPageW, HeightPx: fragPageH, Media: "screen"})
}

// fragOutlineHTML turns layout.Outline into an indented list. A closed entry
// is marked, and its children are shown dimmed the way a collapsed tree would
// hide them.
func fragOutlineHTML(ctx context.Context) (string, error) {
	doc, err := fragDocument(ctx, fragBookmarkPage)
	if err != nil {
		return "", err
	}

	entries, err := layout.Outline(ctx, doc)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString(`<div id="frag-outline">`)

	for i, e := range entries {
		state := "open"
		if !e.Open {
			state = "closed"
		}

		fmt.Fprintf(&b, `<div id="frag-outline-%d" style="margin-left:%dpx">level %d, %s: %s</div>`,
			i, (e.Level-1)*fragOutlineIndent, e.Level, state, html.EscapeString(e.Label))
	}

	b.WriteString(`</div>`)

	return b.String(), nil
}

// fragRunningHeaderHTML shows the copies as a header bar above the page.
func fragRunningHeaderHTML(ctx context.Context) (string, error) {
	doc, err := fragDocument(ctx, fragCopyPage)
	if err != nil {
		return "", err
	}

	copies, err := layout.CopiedFragments(ctx, doc)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString(`<div id="frag-running-header" style="background:#e5e7eb;padding:4px 8px">`)

	for _, c := range copies {
		fmt.Fprintf(&b, `<span id="frag-copy-%s">%s = %s</span> `, c.Name, c.Name, html.EscapeString(c.Text))
	}

	b.WriteString(`</div>`)

	return b.String(), nil
}

// fragFootnoteHTML prints the footnote values the engine exports on each box.
func fragFootnoteHTML(ctx context.Context) (string, error) {
	doc, err := fragDocument(ctx, fragFootnotePage)
	if err != nil {
		return "", err
	}

	display, err := layout.DisplayList(ctx, doc)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString(`<div id="frag-footnotes" style="font-family:monospace;font-size:12px">`)

	for _, box := range display.Boxes {
		if !strings.HasPrefix(box.ID, "fn-") {
			continue
		}

		fmt.Fprintf(&b, `<div id="%s-read">%s: footnote-display=%q footnote-policy=%q</div>`,
			box.ID, box.ID, box.Declared["footnote-display"], box.Declared["footnote-policy"])
	}

	b.WriteString(`</div>`)

	return b.String(), nil
}

const fragEllipsisProse = "The quick brown fox jumps over the lazy dog near the river bank " +
	"while the red fox watches the slow green turtle cross the road again."

// fragEllipsisHTML shows one clamped paragraph per marker.
func fragEllipsisHTML() string {
	cases := []struct{ id, decl string }{
		{"frag-ell-default", "line-clamp:2"},
		{"frag-ell-string", "line-clamp:2; block-ellipsis:' [more]'"},
		{"frag-ell-none", "line-clamp:2; block-ellipsis:no-ellipsis"},
	}

	var b strings.Builder

	for _, c := range cases {
		fmt.Fprintf(&b, `<p id="%s" style="width:300px;margin:0 0 6px 0;%s">%s</p>`, c.id, c.decl, fragEllipsisProse)
	}

	return b.String()
}

// Sizes of the fragmentation demo page and the outline indent, in CSS pixels.
const (
	fragPageW         = 500
	fragPageH         = 400
	fragOutlineIndent = 18
)
