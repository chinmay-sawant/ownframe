package main

import (
	"fmt"
	"strings"
)

// textSpec returns the demo of a text, svg or anim property, or false when
// another group file answers for it.
func textSpec(r genProp) (DemoSpec, bool) {
	var html string

	switch r.Group {
	case "text":
		html = textHTML(r.Name)
	case "svg":
		html = svgHTML(r.Name)
	case "anim":
		html = animHTML(r.Name)
	default:
		return DemoSpec{}, false
	}

	return DemoSpec{HTML: html, Look: textLook(r), Visible: textVisible[r.Name]}, true
}

const (
	textBody = `Sphinx of black quartz, judge my vow. Fifty 1234 words.`
	textBase = `width:160px;font-size:13px;line-height:1.3;background:#f1efe7;padding:4px;`
)

// tbox is the plain measured text box: base is fixed style that goes before the
// property, body is the content.
func tbox(base, body string) string {
	return `<div id="{{ID}}" style="` + textBase + base + `{{STYLE}}">` + body + `</div>`
}

func textHTML(name string) string {
	switch name {
	case "alignment-baseline", "dominant-baseline", "baseline-shift", "text-anchor", "glyph-orientation-vertical":
		target := `<text id="{{ID}}-1" x="60" y="40" font-size="16" fill="#1f2937" style="{{STYLE}}">Hag<tspan>ij</tspan></text>`
		if name == "baseline-shift" {
			target = `<text id="{{ID}}-1" x="20" y="40" font-size="16" fill="#1f2937">Hag<tspan id="{{ID}}-2" style="{{STYLE}}">ij</tspan>k</text>`
		}

		return `<div id="{{ID}}" style="width:120px;height:60px;background:#f1efe7"><svg width="120" height="60" viewBox="0 0 120 60">` +
			`<line x1="0" y1="40" x2="120" y2="40" stroke="#c0392b" stroke-width="1"/>` + target + `</svg></div>`
	case "vertical-align":
		return tbox("", `Text <span id="{{ID}}-1" style="{{STYLE}}">raised</span> and <small>small</small> inline.`)
	case "ruby-align", "ruby-merge", "ruby-overhang", "ruby-position":
		return tbox("", `A <ruby id="{{ID}}-1" style="{{STYLE}}">base<rt>top</rt></ruby> word <ruby>two<rt>x</rt></ruby>.`)
	case "font-variant-east-asian":
		return tbox("", `<ruby>base<rt>top</rt></ruby> ruby text 123`)
	case "text-combine-upright", "text-orientation":
		return tbox("", `Year 2026 upright digits`)
	case "initial-letter", "initial-letter-align", "initial-letter-wrap":
		return tbox("", `Drop cap paragraph that wraps around the first letter of the text.`)
	case "hyphens", "hyphenate-character", "hyphenate-limit-chars", "hyphenate-limit-last", "hyphenate-limit-lines", "hyphenate-limit-zone":
		return `<div id="{{ID}}" lang="en" style="width:90px;font-size:13px;background:#f1efe7;padding:4px;hyphens:auto;{{STYLE}}">` +
			`Extraordinarily internationalization of uncharacteristically long words.</div>`
	case "line-clamp":
		return tbox("overflow:hidden;", textBody+` `+textBody+` `+textBody)
	case "block-ellipsis":
		return tbox("overflow:hidden;line-clamp:2;", textBody+` `+textBody+` `+textBody)
	case "text-overflow":
		return tbox("white-space:nowrap;overflow:hidden;", `One very long line of text that does not fit the box.`)
	case "white-space", "white-space-collapse", "white-space-trim":
		return tbox("", "Spaced     words\n   and a new line\nthird line   ")
	case "tab-size":
		return tbox("white-space:pre;", "a\tb\n\tc\td")
	case "text-wrap", "text-wrap-mode", "text-wrap-style":
		return tbox("", `Wrapping text that fills a narrow box with several short words here.`)
	case "text-align", "text-align-all":
		return tbox("", `Aligned text across two lines of this box.`)
	case "text-align-last":
		return tbox("text-align:right;", `Aligned text across two lines of this box.`)
	case "text-justify":
		return tbox("text-align:justify;", textBody+` `+textBody)
	case "direction", "unicode-bidi":
		return tbox("", `ABC `+"אבג"+` 123 end.`)
	case "letter-spacing":
		return `<div style="letter-spacing:0">` + tbox("letter-spacing:2px;", textBody) + `</div>`
	case "word-spacing":
		return `<div style="word-spacing:0">` + tbox("word-spacing:6px;", textBody) + `</div>`
	case "text-decoration":
		return `<div style="text-decoration:none">` + tbox("text-decoration:underline;", textBody) + `</div>`
	case "quotes":
		return tbox("", `Say <q id="{{ID}}-1">hello <q>nested</q></q> now.`)
	case "content":
		return tbox("", `<span id="{{ID}}-1" style="{{STYLE}}">original</span> text`)
	case "bookmark-label", "bookmark-level", "bookmark-state", "copy-into", "footnote-display", "footnote-policy":
		return tbox("", `Heading text<span id="{{ID}}-1" style="display:block">and a second child</span>`)
	case "counter-reset":
		return counterHTML("counter-reset:k 5;", "")
	case "counter-increment", "counter-set":
		return counterHTML("counter-reset:k;", "{{STYLE}}")
	case "text-decoration-color", "text-decoration-thickness", "text-decoration-style", "text-decoration-inset",
		"text-underline-offset", "text-underline-position", "text-decoration-skip", "text-decoration-skip-box",
		"text-decoration-skip-ink", "text-decoration-skip-self", "text-decoration-skip-spaces", "text-decoration-line":
		return tbox("text-decoration:underline;", `Underlined gypsy jumping quickly over lazy dogs.`)
	case "text-emphasis", "text-emphasis-color", "text-emphasis-position", "text-emphasis-skip", "text-emphasis-style":
		return tbox("text-emphasis:filled #1f2937;", `Emphasis marks over words`)
	case "font-synthesis", "font-synthesis-position", "font-synthesis-small-caps", "font-synthesis-style", "font-synthesis-weight":
		return tbox("font-weight:700;font-style:italic;", `Synthesised bold italic text`)
	case "text-box", "text-box-edge", "text-box-trim":
		return `<div id="{{ID}}" style="width:160px;font-size:20px;line-height:2;background:#cfe3f7;{{STYLE}}">Trimmed Hxg</div>`
	case "font-size", "font-size-adjust", "line-height", "line-padding", "line-fit-edge":
		return tbox("", textBody)
	case "text-indent":
		return tbox("", textBody+` `+textBody)
	case "font-variant-emoji":
		return tbox("", "Emoji ❤ ☺ ★ text")
	default:
		return tbox("", `<b>Bold</b> and <i>italic</i> `+textBody+` <code>fi ffl 1234</code>`)
	}
}

func counterHTML(base, onFirst string) string {
	style := `<style>.cq::before{content:counter(k) ") "}</style>`

	first := onFirst
	if first == "" {
		first = "counter-increment:k;"
	}

	return style + `<div id="{{ID}}" style="width:160px;font-size:13px;background:#f1efe7;padding:4px;` + base + `{{STYLE}}">` +
		`<div id="{{ID}}-1" class="cq" style="counter-increment:k;` + strings.ReplaceAll(onFirst, "{{STYLE}}", "{{STYLE}}") + `">one</div>` +
		`<div id="{{ID}}-2" class="cq" style="counter-increment:k;">two</div></div>`
}

const svgWrap = `<div id="{{ID}}" style="width:120px;height:80px;background:#f1efe7"><svg width="120" height="80" viewBox="0 0 120 80"%s>%s</svg></div>`

const tri = `M15 70 L60 12 L105 70 Z`

func svgHTML(name string) string {
	wrap := func(svgStyle, inner string) string {
		attr := ""
		if svgStyle != "" {
			attr = ` style="` + svgStyle + `"`
		}

		return fmt.Sprintf(svgWrap, attr, inner)
	}
	path := func(attrs string) string {
		return `<path id="{{ID}}-1" d="` + tri + `" ` + attrs + `/>`
	}
	marker := `<defs><marker id="{{ID}}-m" markerWidth="8" markerHeight="8" refX="4" refY="4" markerUnits="userSpaceOnUse"><circle cx="4" cy="4" r="4" fill="#c0392b"/></marker></defs>`
	markerLine := func(base string) string {
		return marker + `<path id="{{ID}}-1" d="M15 60 L60 20 L105 60" fill="none" stroke="#1f2937" stroke-width="2" style="` + base + `{{STYLE}}"/>`
	}
	grad := func(stops string) string {
		return `<defs><linearGradient id="{{ID}}-g" x1="0" y1="0" x2="1" y2="0">` + stops + `</linearGradient></defs>` +
			`<rect id="{{ID}}-1" x="10" y="10" width="100" height="60" fill="url(#{{ID}}-g)"/>`
	}
	filter := func(prim, shapeAttrs string) string {
		return `<defs><filter id="{{ID}}-f" x="0" y="0" width="1" height="1">` + prim + `</filter></defs>` +
			`<rect id="{{ID}}-1" x="10" y="10" width="100" height="60" fill="#3b82f6" filter="url(#{{ID}}-f)" ` + shapeAttrs + `/>`
	}

	switch name {
	case "fill":
		return wrap("{{STYLE}}", path(`stroke="#1f2937" stroke-width="3"`))
	case "fill-opacity":
		return wrap("{{STYLE}}", path(`fill="#c0392b" stroke="#1f2937" stroke-width="3"`))
	case "fill-rule":
		return wrap("{{STYLE}}", `<path id="{{ID}}-1" d="M60 5 L95 75 L10 30 L110 30 L25 75 Z" fill="#c0392b"/>`)
	case "stroke", "stroke-opacity":
		return wrap("{{STYLE}}", path(`fill="#e5e7eb" stroke-width="6"`+map[bool]string{true: ` stroke="#1f2937"`}[name == "stroke-opacity"]))
	case "stroke-width":
		return wrap("{{STYLE}}", path(`fill="#e5e7eb" stroke="#1f2937"`))
	case "stroke-dasharray", "stroke-dashoffset":
		return wrap("{{STYLE}}", `<path id="{{ID}}-1" d="M10 40 L110 40" fill="none" stroke="#1f2937" stroke-width="6"`+map[bool]string{true: ` stroke-dasharray="8 4"`}[name == "stroke-dashoffset"]+`/>`)
	case "stroke-linecap":
		return wrap("{{STYLE}}", `<path id="{{ID}}-1" d="M20 40 L100 40" fill="none" stroke="#1f2937" stroke-width="12"/>`)
	case "stroke-linejoin", "stroke-miterlimit":
		return wrap("{{STYLE}}", `<path id="{{ID}}-1" d="M15 70 L60 12 L105 70" fill="none" stroke="#1f2937" stroke-width="10" stroke-linejoin="miter"/>`)
	case "paint-order":
		return wrap("{{STYLE}}", path(`fill="#c0392b" stroke="#1f2937" stroke-width="12"`))
	case "path-length":
		return wrap("", `<path id="{{ID}}-1" d="M10 40 L110 40" fill="none" stroke="#1f2937" stroke-width="6" stroke-dasharray="1" style="{{STYLE}}"/>`)
	case "vector-effect":
		return `<div id="{{ID}}" style="width:120px;height:80px;background:#f1efe7"><svg width="120" height="80" viewBox="0 0 60 40">` +
			`<path id="{{ID}}-1" d="M5 35 L30 5 L55 35" fill="none" stroke="#1f2937" stroke-width="4" style="{{STYLE}}"/></svg></div>`
	case "cx", "cy", "r":
		return wrap("", `<circle id="{{ID}}-1" cx="60" cy="40" r="25" fill="#3b82f6" style="{{STYLE}}"/>`)
	case "rx", "ry":
		return wrap("", `<rect id="{{ID}}-1" x="20" y="15" width="80" height="50" fill="#3b82f6" style="{{STYLE}}"/>`)
	case "x", "y":
		return wrap("", `<rect id="{{ID}}-1" x="5" y="5" width="50" height="40" fill="#3b82f6" style="{{STYLE}}"/>`)
	case "d":
		return wrap("", `<path id="{{ID}}-1" d="M10 70 L110 70 L110 40 Z" fill="#3b82f6" style="{{STYLE}}"/>`)
	case "marker", "marker-start", "marker-mid", "marker-end", "marker-side":
		base := "marker-start:url(#{{ID}}-m);marker-mid:url(#{{ID}}-m);marker-end:url(#{{ID}}-m);"
		return wrap("", markerLine(base))
	case "stop-color":
		return wrap("", grad(`<stop offset="0" stop-color="#3b82f6"/><stop offset="1" stop-color="#1f2937" style="{{STYLE}}"/>`))
	case "stop-opacity":
		return wrap("", grad(`<stop offset="0" stop-color="#3b82f6"/><stop offset="1" stop-color="#c0392b" style="{{STYLE}}"/>`))
	case "color-interpolation":
		return wrap("", `<defs><linearGradient id="{{ID}}-g" x1="0" y1="0" x2="1" y2="0" style="{{STYLE}}"><stop offset="0" stop-color="#c0392b"/><stop offset="1" stop-color="#3b82f6"/></linearGradient></defs>`+
			`<rect id="{{ID}}-1" x="10" y="10" width="100" height="60" fill="url(#{{ID}}-g)" style="{{STYLE}}"/>`)
	case "color-interpolation-filters":
		return wrap("", filter(`<feGaussianBlur stdDeviation="6" style="{{STYLE}}"/>`, ``))
	case "flood-color", "flood-opacity":
		return wrap("", filter(`<feFlood style="{{STYLE}}"/>`, ``))
	case "lighting-color":
		return wrap("", filter(`<feDiffuseLighting in="SourceGraphic" style="{{STYLE}}"><feDistantLight azimuth="45" elevation="60"/></feDiffuseLighting>`, ``))
	}

	return wrap("{{STYLE}}", path(`fill="#e5e7eb" stroke="#1f2937" stroke-width="3"`))
}

const animKeyframes = `<style>@keyframes slide{from{margin-left:20px;opacity:.5}to{margin-left:60px;opacity:1}}</style>`

func animBox(base string) string {
	return animKeyframes + `<div id="{{ID}}" style="width:120px;height:30px;background:#3b82f6;color:#fff;font-size:12px;` + base + `{{STYLE}}">box</div>`
}

const animRun = `animation:slide 2s linear infinite;`

func animHTML(name string) string {
	switch {
	case strings.HasPrefix(name, "animation-") && name != "animation-name" && name != "animation-timeline" && name != "animation-trigger":
		return animBox(animRun)
	case name == "animation-name":
		return animBox(animRun)
	case name == "animation":
		return animBox("")
	case strings.HasPrefix(name, "transition"):
		return animKeyframes + `<style>.tr{background:#3b82f6}.tr:hover{background:#c0392b}</style><div id="{{ID}}" class="tr" style="width:120px;height:30px;color:#fff;font-size:12px;{{STYLE}}">hover target</div>`
	case strings.HasPrefix(name, "offset"):
		return `<div style="width:200px;height:50px;background:#f1efe7"><div id="{{ID}}" style="width:30px;height:20px;background:#3b82f6;offset-path:path('M0 0 L100 0');offset-distance:0px;{{STYLE}}"></div></div>`
	case name == "will-change":
		return animBox("")
	case strings.Contains(name, "timeline") && !strings.Contains(name, "trigger") && name != "animation-timeline":
		return `<div id="{{ID}}" style="width:160px;height:60px;overflow:auto;background:#f1efe7;{{STYLE}}"><div id="{{ID}}-1" style="height:140px;background:#3b82f6;color:#fff;font-size:12px">tall content in a scroller</div></div>`
	case name == "animation-timeline" || name == "animation-trigger":
		return animBox(animRun)
	default:
		return animBox("")
	}
}

// textVisible lists the properties whose demo changes the engine's layout or
// paint operations (measured by the visible-spec test).
var textVisible = map[string]bool{
	"content-visibility":          true,
	"counter-increment":           true,
	"counter-reset":               true,
	"direction":                   true,
	"font-family":                 true,
	"font-size":                   true,
	"font-size-adjust":            true,
	"font-synthesis-small-caps":   true,
	"font-variant":                true,
	"font-variant-caps":           true,
	"font-variant-position":       true,
	"font-weight":                 true,
	"font-width":                  true,
	"initial-letter":              true,
	"letter-spacing":              true,
	"line-clamp":                  true,
	"line-height":                 true,
	"line-padding":                true,
	"offset-anchor":               true,
	"ruby-position":               true,
	"tab-size":                    true,
	"text-align":                  true,
	"text-align-all":              true,
	"text-box":                    true,
	"text-box-trim":               true,
	"text-decoration":             true,
	"text-decoration-color":       true,
	"text-decoration-inset":       true,
	"text-decoration-skip":        true,
	"text-decoration-skip-ink":    true,
	"text-decoration-skip-self":   true,
	"text-decoration-skip-spaces": true,
	"text-decoration-style":       true,
	"text-decoration-thickness":   true,
	"text-emphasis":               true,
	"text-emphasis-color":         true,
	"text-emphasis-position":      true,
	"text-indent":                 true,
	"text-shadow":                 true,
	"text-transform":              true,
	"text-underline-offset":       true,
	"text-underline-position":     true,
	"text-wrap":                   true,
	"text-wrap-mode":              true,
	"unicode-bidi":                true,
	"vertical-align":              true,
	"white-space":                 true,
	"word-break":                  true,
	"word-spacing":                true,
	"writing-mode":                true,
}

// textLookOf holds property-specific one-line notes.
var textLookOf = map[string]string{}

func textLook(r genProp) string {
	if s, ok := textLookOf[r.Name]; ok {
		return s
	}

	motion := ""
	if r.Group == "anim" {
		motion = " The motion itself is on the Motion tab."
	}

	if textVisible[r.Name] {
		return "The declaration changes the measured layout or the paint operations of the demo; compare the two boxes." + motion
	}

	return "The declaration is parsed and exported to the host; the demo's layout and paint operations stay the same as the control here." + motion
}
