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
	case "alignment-baseline", "dominant-baseline":
		return `<div id="{{ID}}" style="width:120px;height:50px;background:#f1efe7;padding:12px 8px;"><p id="{{ID}}-1" style="margin:0;font-size:16px;color:#1f2937;{{STYLE}}">Hag</p></div>`
	case "baseline-source":
		return `<div id="{{ID}}" style="width:160px;font-size:13px;background:#f1efe7;padding:4px;">Before <span id="{{ID}}-1" style="display:inline-block;width:54px;background:#dfe9f3;{{STYLE}}">line1 line2 line3</span> after</div>`
	case "baseline-shift", "text-anchor", "glyph-orientation-vertical":
		target := `<text id="{{ID}}-1" x="60" y="40" font-size="16" fill="#1f2937" style="{{STYLE}}">Hag<tspan>ij</tspan></text>`
		if name == "baseline-shift" {
			target = `<text id="{{ID}}-1" x="20" y="40" font-size="16" fill="#1f2937">Hag<tspan id="{{ID}}-2" style="{{STYLE}}">ij</tspan>k</text>`
		}

		return `<div id="{{ID}}" style="width:120px;height:60px;background:#f1efe7"><svg width="120" height="60" viewBox="0 0 120 60">` +
			`<line x1="0" y1="40" x2="120" y2="40" stroke="#c0392b" stroke-width="1"/>` + target + `</svg></div>`
	case "vertical-align":
		return tbox("padding-top:22px;", `Text<span id="{{ID}}-1" style="{{STYLE}}">raised</span> and more inline text.`)
	case "ruby-align":
		return tbox("padding:14px 4px;", `A <ruby id="{{ID}}-1" style="{{STYLE}}">longbaseword<rt>rt</rt></ruby> end.`)
	case "ruby-merge":
		return tbox("padding:14px 4px;", `A <ruby id="{{ID}}-1" style="{{STYLE}}">a<rt>1</rt>b<rt>2</rt></ruby> end.`)
	case "ruby-overhang":
		return tbox("padding:14px 4px;", `A <ruby id="{{ID}}-1" style="{{STYLE}}">b<rt>verylongrubytext</rt></ruby> end.`)
	case "ruby-position":
		return tbox("padding:14px 4px;", `A <ruby id="{{ID}}-1" style="{{STYLE}}">base<rt>top</rt></ruby> word <ruby>two<rt>x</rt></ruby>.`)
	case "font-variant-east-asian":
		return tbox("font-family:'Blinkless OT';font-size:22px;", "あいう")
	case "text-combine-upright":
		return `<div id="{{ID}}" style="writing-mode:vertical-rl;height:120px;font-size:14px;background:#f1efe7;">年<span id="{{ID}}-1" style="{{STYLE}}">26</span>日</div>`
	case "text-orientation":
		return `<div id="{{ID}}" style="writing-mode:vertical-rl;height:120px;font-size:14px;background:#f1efe7;{{STYLE}}">Year 2026</div>`
	case "initial-letter":
		// the first word is sized to two lines; the box is tall enough to hold it.
		return tbox("height:46px;padding-top:14px;", `Drop`)
	case "initial-letter-align", "initial-letter-wrap":
		return tbox("height:46px;padding-top:14px;", `<span style="initial-letter:2;{{STYLE}}">D</span>rop cap paragraph`)
	case "hyphens", "hyphenate-character", "hyphenate-limit-chars", "hyphenate-limit-lines", "hyphenate-limit-zone":
		return `<div id="{{ID}}" lang="en" style="width:90px;font-size:13px;background:#f1efe7;padding:4px;hyphens:auto;{{STYLE}}">` +
			`Extra&shy;ordinarily inter&shy;nationalization</div>`
	case "hyphenate-limit-last":
		return `<div id="{{ID}}" lang="en" style="width:110px;font-size:13px;background:#f1efe7;padding:4px;hyphens:auto;{{STYLE}}">See extra&shy;ordinarily</div>`
	case "line-break":
		return `<div id="{{ID}}" style="width:120px;font-size:13px;background:#f1efe7;padding:4px;word-break:keep-all;{{STYLE}}">WWWWWWWWWWWWWWWWWWWWWW</div>`
	case "line-clamp":
		return tbox("overflow:hidden;", textBody+` `+textBody+` `+textBody)
	case "block-ellipsis":
		return tbox("overflow:hidden;line-clamp:2;", textBody+` `+textBody+` `+textBody)
	case "text-overflow":
		return tbox("white-space:nowrap;overflow:hidden;", `One very long line of text that does not fit the box.`)
	case "white-space", "white-space-collapse":
		return tbox("", "Spaced     words\n   and a new line\nthird line   ")
	case "white-space-trim":
		return tbox("white-space:pre;", "   leading spaces")
	case "hanging-punctuation":
		return tbox("text-align:left;", `“Quoted text that starts with a quote mark.”`)
	case "text-group-align":
		return tbox("", "short\nlonger line")
	case "text-spacing", "text-spacing-trim":
		return tbox("", "「日本語」の（括弧）")
	case "word-space-transform":
		return tbox("", "word\u200Bword")
	case "tab-size":
		// real tab characters (U+0009) under white-space:pre, so tab-size moves the columns.
		return tbox("white-space:pre;", "a&#9;b&#9;c\n&#9;d&#9;e")
	case "text-shadow":
		return tbox("font-size:28px;font-weight:700;padding:10px 12px;", `Shadow`)
	case "text-transform":
		return tbox("", `Mixed case words wrap over two lines here.`)
	case "text-wrap", "text-wrap-mode":
		return tbox("", `Wrapping text that fills a narrow box with several short words here.`)
	case "text-wrap-style":
		return `<div id="{{ID}}" style="width:220px;font-size:13px;line-height:1.3;background:#f1efe7;padding:4px;{{STYLE}}">The quick brown fox jumps over the lazy dog</div>`
	case "text-align":
		return tbox("", `Aligned text across two lines of this box.`)
	case "text-align-all":
		// three lines show which lines `end` reaches.
		return tbox("", `Aligned text across three short lines of this wide box here.`)
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
		return `<style>#{{ID}}-1::before{content:open-quote}#{{ID}}-1::after{content:close-quote}</style>` + tbox("", `Say <q id="{{ID}}-1" style="{{STYLE}}">hello</q> now.`)
	case "content":
		return `<style>#{{ID}}-1::before{ {{RAWSTYLE}} }</style>` + tbox("", `<span id="{{ID}}-1">original</span> text`)
	case "bookmark-label", "bookmark-level", "bookmark-state", "copy-into", "footnote-display", "footnote-policy":
		return tbox("", `Heading text<span id="{{ID}}-1" style="display:block">and a second child</span>`)
	case "counter-reset":
		return counterHTML("counter-reset:k 5;", "")
	case "counter-increment", "counter-set":
		return counterHTML("counter-reset:k;", "{{STYLE}}")
	case "text-decoration-color", "text-decoration-thickness", "text-decoration-style", "text-decoration-inset",
		"text-underline-offset", "text-underline-position", "text-decoration-skip",
		"text-decoration-skip-ink", "text-decoration-skip-self", "text-decoration-skip-spaces":
		return tbox("text-decoration:underline;line-height:3.4;padding:4px 4px 24px;", `Underlined gypsy jumping quickly over lazy dogs.`)
	case "text-decoration-skip-box":
		return `<div id="{{ID}}" style="width:160px;font-size:14px;background:#f1efe7;padding:8px 4px;"><span style="text-decoration:underline">ab</span><span id="{{ID}}-1" style="text-decoration:underline;padding:0 6px;{{STYLE}}">cd</span><span style="text-decoration:underline">ef</span></div>`
	case "text-decoration-line":
		return tbox("line-height:2;", `Decorated text here.`)
	case "text-emphasis", "text-emphasis-color", "text-emphasis-position", "text-emphasis-style":
		return tbox("text-emphasis:filled #1f2937;padding:24px 4px 30px;line-height:3;", `Emphasis marks over words`)
	case "text-emphasis-skip":
		return tbox("text-emphasis:filled #1f2937;text-emphasis-skip:none;padding:24px 4px 30px;line-height:3;", `a b c d`)
	case "font-synthesis":
		return tbox("font-variant-caps:small-caps;font-size:16px;", `Synthesised small caps`)
	case "font-synthesis-weight":
		return tbox("font-family:'Blinkless OT';font-weight:700;font-size:20px;", `Bold text`)
	case "font-synthesis-position":
		return tbox("font-variant-position:super;font-size:18px;", `x2 + y3`)
	case "font-synthesis-small-caps":
		return tbox("font-variant-caps:small-caps;font-size:18px;", `Small Caps`)
	case "font-synthesis-style":
		return tbox("font-family:'Blinkless OT';font-style:italic;font-size:20px;", `Italic text`)
	case "font-variant-numeric":
		return tbox("font-family:'Blinkless OT';font-size:20px;", "1111<br>8888")
	case "font-optical-sizing":
		return tbox("font-family:'Blinkless Var';font-size:48px;", "AAAA")
	case "font-variation-settings":
		return tbox("font-family:'Blinkless Var';font-size:28px;", "AAAA")
	case "font-palette":
		return tbox("font-family:'Blinkless Palette';font-size:24px;color:#111111;", "Ag")
	case "text-box", "text-box-trim":
		return `<div id="{{ID}}" style="width:160px;font-size:20px;line-height:2;padding:8px 0;background:#cfe3f7;{{STYLE}}">Trimmed Hxg</div>`
	case "text-box-edge":
		return `<div id="{{ID}}" style="width:160px;font-size:20px;line-height:2;padding:8px 0;background:#cfe3f7;text-box-trim:trim-both;{{STYLE}}">Trimmed Hxg</div>`
	case "font-size-adjust":
		// aspect 2 scales the glyphs about four times: start from a 5px face.
		return tbox("font-size:5px;line-height:30px;height:30px;overflow:hidden;", `Sphinx`)
	case "font-size", "line-height", "line-padding":
		return tbox("", textBody)
	case "line-fit-edge":
		return `<div id="{{ID}}" style="width:160px;font-size:16px;background:#f1efe7;"><span id="{{ID}}-1" style="padding:8px;background:#dfe9f3;{{STYLE}}">Ag</span></div>`
	case "text-indent":
		return tbox("", textBody+` `+textBody)
	case "font-variant-emoji":
		return tbox("font-variant-emoji:emoji;", "Emoji ❤ ☺ ★ text")
	case "font-feature-settings", "font-variant-ligatures":
		return tbox("font-family:'DejaVu Sans';font-size:20px;", "fi fl ffi ffl")
	case "font-language-override":
		return tbox("font-family:'DejaVu Sans';font-size:24px;", "б бб")
	case "font-variant-alternates":
		return tbox("font-family:'DejaVu Sans';font-size:20px;", "I J a l")
	case "text-autospace":
		return tbox("", "日本語abc 123 漢字xyz")
	default:
		return tbox("", `<b>Bold</b> and <i>italic</i> `+textBody+` <code>fi ffl 1234</code>`)
	}
}

func counterHTML(base, onFirst string) string {
	style := `<style>.cq::before{content:counter(k) ") "}</style>`

	return style + `<div id="{{ID}}" style="width:160px;font-size:13px;background:#f1efe7;padding:4px;` + base + `{{STYLE}}">` +
		`<div id="{{ID}}-1" class="cq" style="counter-increment:k;` + onFirst + `">one</div>` +
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
	case "stroke-linejoin":
		return wrap("{{STYLE}}", `<path id="{{ID}}-1" d="M15 70 L60 12 L105 70" fill="none" stroke="#1f2937" stroke-width="10"/>`)
	case "stroke-miterlimit":
		return wrap("{{STYLE}}", `<path id="{{ID}}-1" d="M42 70 L60 10 L78 70" fill="none" stroke="#1f2937" stroke-width="10"/>`)
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
		return wrap("", filter(`<feDiffuseLighting in="SourceGraphic" color-interpolation-filters="sRGB" style="{{STYLE}}"><feDistantLight azimuth="45" elevation="45"/></feDiffuseLighting>`, ``))
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
//
//nolint:gochecknoglobals // read-only property lookup table
var textVisible = map[string]bool{
	"alignment-baseline":          true,
	"baseline-source":             true,
	"content":                     true,
	"content-visibility":          true,
	"counter-increment":           true,
	"counter-reset":               true,
	"counter-set":                 true,
	"direction":                   true,
	"dominant-baseline":           true,
	"font":                        true,
	"font-family":                 true,
	"font-optical-sizing":         true,
	"font-palette":                true,
	"font-size":                   true,
	"font-size-adjust":            true,
	"font-style":                  true,
	"font-synthesis-position":     true,
	"font-synthesis-small-caps":   true,
	"font-synthesis-style":        true,
	"font-variant":                true,
	"font-variant-caps":           true,
	"font-variant-position":       true,
	"font-variation-settings":     true,
	"font-weight":                 true,
	"hanging-punctuation":         true,
	"hyphenate-character":         true,
	"hyphenate-limit-chars":       true,
	"hyphenate-limit-last":        true,
	"hyphenate-limit-lines":       true,
	"hyphenate-limit-zone":        true,
	"hyphens":                     true,
	"initial-letter":              true,
	"initial-letter-align":        true,
	"initial-letter-wrap":         true,
	"letter-spacing":              true,
	"line-break":                  true,
	"line-clamp":                  true,
	"line-fit-edge":               true,
	"line-height":                 true,
	"line-padding":                true,
	"offset-anchor":               true,
	"quotes":                      true,
	"ruby-align":                  true,
	"ruby-merge":                  true,
	"ruby-overhang":               true,
	"ruby-position":               true,
	"text-align":                  true,
	"text-align-all":              true,
	"text-align-last":             true,
	"text-autospace":              true,
	"text-box":                    true,
	"text-box-edge":               true,
	"text-box-trim":               true,
	"text-combine-upright":        true,
	"text-decoration":             true,
	"text-decoration-color":       true,
	"text-decoration-inset":       true,
	"text-decoration-line":        true,
	"text-decoration-skip":        true,
	"text-decoration-skip-box":    true,
	"text-decoration-skip-ink":    true,
	"text-decoration-skip-self":   true,
	"text-decoration-skip-spaces": true,
	"text-decoration-style":       true,
	"text-decoration-thickness":   true,
	"text-emphasis":               true,
	"text-emphasis-color":         true,
	"text-emphasis-position":      true,
	"text-emphasis-skip":          true,
	"text-group-align":            true,
	"text-indent":                 true,
	"text-justify":                true,
	"text-orientation":            true,
	"text-shadow":                 true,
	"text-spacing":                true,
	"text-spacing-trim":           true,
	"text-transform":              true,
	"text-underline-offset":       true,
	"text-underline-position":     true,
	"text-wrap":                   true,
	"text-wrap-mode":              true,
	"text-wrap-style":             true,
	"unicode-bidi":                true,
	"vertical-align":              true,
	"white-space":                 true,
	"white-space-collapse":        true,
	"white-space-trim":            true,
	"word-break":                  true,
	"word-space-transform":        true,
	"word-spacing":                true,
	"writing-mode":                true,
}

// textLookOf holds property-specific one-line notes.
//
//nolint:gochecknoglobals // read-only property lookup table
var textLookOf = map[string]string{
	"text-transform":              "The demo text turns upper case, so the glyphs are wider and the box wraps onto more lines; the control keeps the mixed case.",
	"text-overflow":               "The one-line text truncates at the 160 px box edge and ends with the custom `x` string instead of running past the edge.",
	"text-wrap":                   "The control wraps the text inside the 160px box; with `nowrap` the demo keeps one line, which is wider than the box and runs past its right edge.",
	"text-wrap-mode":              "The control wraps the text inside the 160px box; with `nowrap` the demo keeps one line, which is wider than the box and runs past its right edge.",
	"text-wrap-style":             "With `balance` the two lines split more evenly (`The quick brown fox` / `jumps over the lazy dog`) instead of filling the first line greedily.",
	"alignment-baseline":          "With `middle` the text baseline shifts upward relative to the default alphabetic baseline.",
	"dominant-baseline":           "With `hanging` the text shifts upward to align its hanging baseline.",
	"baseline-source":             "With `first` the multi-line inline-block aligns to the surrounding text on its first line instead of its last line.",
	"content":                     "The `::before` pseudo-element inserts the generated string `x` immediately before `original`.",
	"quotes":                      "Custom `«` and `»` quote marks wrap `hello` via `open-quote`/`close-quote` instead of the default double quotes.",
	"text-decoration-skip-box":    "With `all` the continuous underline breaks across the padded `cd` inline span into separate segments.",
	"line-break":                  "With `anywhere` the unbreakable `W...` token wraps across multiple lines inside the 120 px box even under `word-break: keep-all`.",
	"line-fit-edge":               "With `text` the inline span drops its vertical padding from the line box height, shrinking the line from 34 px to 18 px.",
	"hyphenate-limit-last":        "With `always` the last line avoids hyphenating the long word at the soft hyphen.",
	"text-autospace":              "The painter replays text-autospace natively, opening extra spacing at ideograph and Latin boundaries.",
	"font-variant-numeric":        "In Blinkless OT, `tabular-nums` (`tnum`) widens the narrow `1` digits so `1111` aligns cleanly with `8888`.",
	"font-variant-ligatures":      "In DejaVu Sans, disabling ligatures breaks the joined `fi`, `fl`, `ffi`, `ffl` glyphs into separate characters.",
	"font-variant-east-asian":     "In Blinkless OT, `ruby` activates the OpenType `ruby` table to substitute compact ruby glyph forms for `あいう`.",
	"font-variant-alternates":     "In DejaVu Sans, `stylistic(1)` substitutes alternate glyph forms for `I`, `J`, `a`, and `l`.",
	"font-synthesis-weight":       "In regular-only Blinkless OT, `font-weight: 700` draws synthetic bold in the control; `font-synthesis-weight: none` disables it so the strokes stay regular.",
	"font-synthesis":              "In the control, lowercase letters are synthesised into small caps. With `font-synthesis: none`, synthesis is disabled and lowercase letters stay lowercase.",
	"font-feature-settings":       "In DejaVu Sans, `\"liga\" 0` breaks the joined `fi`, `fl`, `ffi`, `ffl` ligatures into separate characters.",
	"font-optical-sizing":         "In Blinkless Var at 48 px, `none` disables automatic `opsz` axis instancing so the glyphs keep their default optical proportions.",
	"font-palette":                "In Blinkless Palette (`COLR`+`CPAL`), `dark` switches the glyphs from the default light palette to the dark orange palette.",
	"font-variation-settings":     "In Blinkless Var, `\"wght\" 900` instances the variable weight axis to its heavy master, widening the glyphs.",
	"font-language-override":      "In DejaVu Sans, `\"SRB\"` activates the Serbian `locl` form of Cyrillic `б`.",
	"fill":                        "The triangle fill changes from the default black to #c0392b red. Judge this row by the picture.",
	"fill-opacity":                "The red triangle fill becomes half transparent, so the background shows through; the dark outline stays solid. Judge this row by the picture.",
	"stroke":                      "The control triangle has no outline; the demo draws a 6px red outline around it. Judge this row by the picture.",
	"stroke-opacity":              "The dark outline of the demo becomes half transparent and looks grey; the control outline is solid. Judge this row by the picture.",
	"stroke-width":                "The outline of the demo is 12px wide; the control's is hairline thin. Judge this row by the picture.",
	"stroke-dasharray":            "The control line is solid; the demo line breaks into short dashes. Judge this row by the picture.",
	"stroke-dashoffset":           "The dash pattern starts 12px later, so the dashes and gaps shift along the line. Judge this row by the picture.",
	"stroke-linecap":              "The line ends of the demo are rounded; the control's are flat. Judge this row by the picture.",
	"vector-effect":               "The demo stroke stays thin instead of scaling with the viewBox, so the line is narrower than the control's. Judge this row by the picture.",
	"rx":                          "The corners of the demo rectangle are rounded; the control's are square. Judge this row by the picture.",
	"ry":                          "The corners of the demo rectangle are rounded; the control's are square. Judge this row by the picture.",
	"d":                           "The demo triangle is the path from the declaration (a small triangle at the top left); the control draws a larger triangle at the bottom. Judge this row by the picture.",
	"path-length":                 "The control line is cut into short 1px dashes. With path-length 2 the engine scales the dash length to the path, so the dashes run together and the demo line is drawn as one unbroken solid line. Judge this row by the picture.",
	"flood-color":                 "The flood fill of the demo is red; the control floods black. Judge this row by the picture.",
	"flood-opacity":               "The flood fill of the demo is 40% opaque and looks grey; the control floods solid black. Judge this row by the picture.",
	"stop-opacity":                "The blue-to-red gradient fades to 40% opacity at its red right stop instead of staying solid red.",
	"fill-rule":                   "The centre pentagon of the five-point star becomes hollow under evenodd instead of staying filled.",
	"marker":                      "The control draws red dots at the path start, middle and end. With `none` all three red dots disappear.",
	"marker-start":                "The control draws red dots at the path start, middle and end. With `none` the start dot disappears while the middle and end dots remain.",
	"marker-mid":                  "The control draws red dots at the path start, middle and end. With `none` the middle dot disappears while the start and end dots remain.",
	"marker-end":                  "The control draws red dots at the path start, middle and end. With `none` the end dot disappears while the start and middle dots remain.",
	"paint-order":                 "The stroke paints beneath the red fill instead of over it, so the red triangle appears wider.",
	"stroke-linejoin":             "The peak of the V-shape rounds off instead of staying a pointed miter.",
	"stroke-miterlimit":           "The sharp peak exceeds miter limit 2 and clips to a flat bevel instead of extending to a point.",
	"stop-color":                  "The right stop of the horizontal gradient changes from dark slate to red.",
	"color-interpolation":         "The red-to-blue gradient interpolates in linearRGB instead of sRGB, brightening the mid-blend.",
	"color-interpolation-filters": "The diffuse lighting filter interpolates in linearRGB instead of sRGB, shifting the lit shading.",
	"cx":                          "The blue circle shifts left so its centre moves from x=60 to x=12.",
	"cy":                          "The blue circle shifts up so its centre moves from y=40 to y=12.",
	"r":                           "The blue circle shrinks from radius 25 to radius 12.",
	"x":                           "The blue rectangle shifts right from x=5 to x=12.",
	"y":                           "The blue rectangle shifts down from y=5 to y=12.",
	"font-size-adjust":            "The control is 5px text; aspect 2 scales the glyphs to about four times that, so the demo word is far larger than the control word inside its fixed-height box.",
	"initial-letter":              "The demo word is drawn about three times larger than the control's word; the engine sizes the whole first word, not just a letter, and does not wrap later text around it, so the box holds this one word.",
	"font-width":                  "The glyphs condense horizontally under ultra-condensed so the lines become narrower.",
	"font-style":                  "The demo text turns italic, so the glyph shapes and widths change; compare the two boxes.",
	"text-align-all":              "In this engine `end` right-aligns every line except the last: the first two lines of the demo sit flush to the right edge and the last line stays at the left edge. The standard says the last line moves too, so this is an engine deviation. `text-align: right` moves all lines.",
	"text-align-last":             "The control is right-aligned. With `start` the first line stays on the right and the last line moves to the left edge.",
	"tab-size":                    "The demo text holds real tab characters under `white-space: pre`. Setting `tab-size: 12px` narrows the tab stops from the default 8 spaces to 12 px, pulling columns `b`, `c`, `d`, and `e` sharply left.",
	"text-anchor":                 "The declaration is set on the SVG text at x=60; with middle the word shifts left so it is horizontally centred on x=60 instead of starting there.",
	"word-spacing":                "The control has 6px of word spacing and the demo inherits the parent's 0, so the word gaps narrow and the text re-wraps.",
	"letter-spacing":              "The control has 2px of letter spacing and the demo inherits the parent's 0, so the glyphs sit closer and the text re-wraps.",
	"text-decoration":             "The control is underlined and the demo inherits the parent's `none`, so the underline disappears.",
	"text-shadow":                 "The demo text gains a hard 2px offset dark shadow (zero blur) behind the bold 28px word. The box keeps the same size as the control, because a shadow does not change layout; the readout sizes match.",
	"text-underline-offset":       "The underline of the demo drops 12px below the text instead of the default offset, so each line's rule sits lower, closer to the next line; the lines are spaced wide and the box has bottom padding so the last rule stays inside.",
	"vertical-align":              "The word `raised` in the demo is lifted 12px above the baseline. The engine moves the whole line with the raised word, so the other words rise too. The markup has no literal space before the span, because the engine already draws one gap there; the box has top padding so the raised line stays inside.",
	"text-emphasis-style":         "The emphasis marks above each character change from filled dots to the custom string `x`.",
	"text-decoration-line":        "A horizontal line-through strike is drawn across the middle of the text instead of plain undecorated text.",
	"text-box":                    "The demo box trims to the cap height and alphabetic baseline of its text, so the blue box is shorter than the control's; the glyphs touch or cross its edges, and the 8px padding keeps the box from collapsing.",
	"text-box-edge":               "With `text-box-trim: trim-both` active, `cap alphabetic` trims the box tighter to the cap height and baseline.",
	"text-box-trim":               "The demo box loses the half-leading at its start edge, so it is shorter than the control's box.",
	"text-emphasis-skip":          "With `spaces` instead of `none`, emphasis marks are skipped over spaces between words.",
	"text-combine-upright":        "In vertical-rl text, the two digits `26` combine horizontally into one upright character cell.",
	"text-orientation":            "In vertical-rl text, `upright` turns Latin letters and digits upright instead of sideways.",
	"font-synthesis-position":     "With `font-variant-position: super`, `none` disables synthetic superscript scaling so the text stays full size.",
	"font-synthesis-small-caps":   "With `font-variant-caps: small-caps`, `none` disables synthetic small caps so lowercase letters stay lowercase.",
	"font-synthesis-style":        "In regular-only Blinkless OT, `font-style: italic` draws synthetic oblique in the control; `none` disables it so the glyphs stay upright.",
	"writing-mode":                "The demo is laid out vertically (the readout shows a different size) and the glyphs are quarter-turned: `vertical-rl` and `vertical-lr` replay on the display list like horizontal text. A rotation that is not a whole quarter turn, or a rotated run that is also skewed or projected, still falls back to the software painter.",
	"text-emphasis":               "The demo marks are red dots instead of the control's dark filled marks, drawn above each word; line-height 3 and padding leave room for the marks.",
	"text-emphasis-color":         "The demo marks turn red; the control marks are dark. Line-height 3 and padding leave room for the marks.",
	"text-emphasis-position":      "The demo marks move below the text; the control marks are above. Line-height 3 and padding leave room for the marks on both sides.",
}

func textLook(r genProp) string {
	if s, ok := textLookOf[r.Name]; ok {
		return s
	}

	motion := ""
	if r.Group == "anim" {
		motion = " The motion itself is on the Motion tab."
	}

	if declines(r.Name) {
		if textVisible[r.Name] {
			return "The engine lays the demo out differently (the readout reports a difference), but the window painter declines this property, so the picture is drawn without it and looks the same as the control." + motion
		}

		return "The engine reports no layout or paint change, and the window painter declines this property, so the picture looks the same as the control." + motion
	}

	if textVisible[r.Name] {
		return "The declaration changes the measured layout or the paint operations of the demo; compare the two boxes." + motion
	}

	return "The declaration is parsed and exported to the host; the demo's layout and paint operations stay the same as the control here." + motion
}
