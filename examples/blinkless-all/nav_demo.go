package main

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

// Authored markup, one snippet per example. navDemoHTML joins them so the
// tests can lay out every example on one page. The ids are what the live
// session (nav_session.go) drives: it patches the style and the text of these
// elements while the page is open.

// navArrows is the row of four on-screen arrow buttons of one focus demo. A
// click on one does what the arrow key does.
func navArrows(demo string) string {
	const btn = `display:inline-block;width:46px;height:24px;line-height:24px;text-align:center;border:1px solid #94a3b8;` +
		`background:#f8fafc;margin-right:4px;border-radius:4px;font-size:13px`

	var b strings.Builder

	b.WriteString(`<div style="margin-top:6px">`)

	for _, d := range []string{"up", "down", "left", "right"} {
		fmt.Fprintf(&b, `<div id="arr-%s-%s" style="%s">%s</div>`, demo, d, btn, strings.ToUpper(d[:1])+d[1:])
	}

	b.WriteString(`</div>`)

	return b.String()
}

// navLiveChip is the result line a live demo rewrites.
func navLiveChip(id, text string) string {
	return `<div id="` + id + `" style="margin:6px 0;padding:3px 8px;background:#e0f2fe;border-radius:10px;font-family:monospace;font-size:12px;display:block;box-sizing:border-box;width:700px;height:40px;overflow:hidden">` + text + `</div>`
}

const navFieldLook = `position:relative;white-space:nowrap;overflow:hidden;height:22px;background:#fff;border:1px solid #cbd5e1;font-family:monospace;font-size:14px;margin:3px 0;box-sizing:border-box;`

// navCaretCell is one editable caret cell: a frame, its text and its caret.
func navCaretCell(id, decl string) string {
	return `<div id="` + id + `" style="` + decl + `;` + navFieldLook + `width:300px">` +
		`<div id="` + id + `-t" style="position:absolute;left:4px;top:2px">type here</div>` +
		`<div id="` + id + `-c" style="position:absolute;left:4px;top:2px;width:2px;height:16px;background:#1f2937"></div></div>`
}

// navSelectLine is one line of selectable text with its highlight behind it.
func navSelectLine(id, decl string) string {
	return `<div id="` + id + `" style="` + decl + `;position:relative;height:20px;width:300px;font-family:monospace;font-size:14px;margin:2px 0">` +
		`<div id="` + id + `-h" style="position:absolute;left:4px;top:2px;width:0;height:16px;background:#93c5fd"></div>` +
		`<div id="` + id + `-t" style="position:absolute;left:4px;top:2px">pick a word here</div></div>`
}

// navSlider is one slider track with its fill and thumb.
func navSlider(id, orient string, w, h int) string {
	return fmt.Sprintf(`<div id="%s" style="slider-orientation:%s;position:relative;width:%dpx;height:%dpx;background:#ddd;margin:8px 0">`+
		`<div id="%s-f" style="position:absolute;left:0;top:0;width:0;height:0;background:#93c5fd"></div>`+
		`<div id="%s-t" style="position:absolute;left:0;top:0;width:10px;height:10px;background:#1d4ed8"></div></div>`, id, orient, w, h, id, id)
}

const navAppearBox = `box-sizing:border-box;width:120px;height:30px;margin-top:6px;padding:0;line-height:24px;text-align:center;font-size:13px`

//nolint:gochecknoglobals,mnd // pixel sizes, timings and markup of the live demo
var (
	navSnipFocus = `<div id="pad" style="position:relative;height:130px;width:330px;background:#f1efe7">
  <button id="n-a" style="nav-down:#n-f;position:absolute;left:0;top:0;width:80px;height:36px">n-a</button>
  <button id="n-b" style="nav-left:#n-e current;position:absolute;left:110px;top:0;width:80px;height:36px">n-b</button>
  <button id="n-c" style="position:absolute;left:220px;top:0;width:80px;height:36px">n-c</button>
  <button id="n-d" style="position:absolute;left:0;top:80px;width:80px;height:36px">n-d</button>
  <button id="n-e" style="position:absolute;left:110px;top:80px;width:80px;height:36px">n-e</button>
  <button id="n-f" style="position:absolute;left:220px;top:80px;width:80px;height:36px">n-f</button>
 </div>`
	navSnipGrid = `<div id="gn" style="spatial-navigation-contain:contain;position:relative;height:100px;width:330px;background:#eef2f7">
  <button id="gn-a" style="position:absolute;left:0;top:0;width:80px;height:30px">gn-a</button>
  <button id="gn-c" style="position:absolute;left:100px;top:60px;width:80px;height:30px">gn-c</button>
 </div>
 <div id="gg" style="spatial-navigation-contain:contain;spatial-navigation-function:grid;position:relative;height:100px;width:330px;background:#eef7ee;margin-top:6px">
  <button id="gg-a" style="position:absolute;left:0;top:0;width:80px;height:30px">gg-a</button>
  <button id="gg-c" style="position:absolute;left:100px;top:60px;width:80px;height:30px">gg-c</button>
 </div>`
	navSnipContain = `<div id="cbox" style="spatial-navigation-contain:contain;position:relative;height:40px;width:330px;background:#f7eeee">
  <button id="c-a" style="position:absolute;left:0;top:0;width:80px;height:30px">c-a</button>
  <button id="c-b" style="position:absolute;left:100px;top:0;width:80px;height:30px">c-b</button>
 </div>
 <div style="position:relative;height:40px;width:330px;margin-top:4px">
  <button id="o-x" style="position:absolute;left:100px;top:0;width:80px;height:30px">o-x</button>
 </div>`
	navSnipAction   = navActionHTML()
	navSnipField    = navFieldBox("fs-content", "field-sizing:content", 90) + navFieldBox("fs-fixed", "field-sizing:fixed", 150)
	navSnipSecurity = navFieldBox("pw-auto", "input-security:auto", 150) + navFieldBox("pw-none", "input-security:none", 150)
	navSnipSelect   = navSelectLine("us-text", "-webkit-user-select:text") + navSelectLine("us-none", "-webkit-user-select:none") +
		navSelectLine("us-all", "-webkit-user-select:all")
	navSnipAdjust = `<div id="ta-none" style="text-size-adjust:none;font-size:14px;height:30px;width:300px;overflow:hidden">ta-none: sample text</div>
 <div id="ta-120" style="text-size-adjust:120%;font-size:14px;height:30px;width:300px;overflow:hidden">ta-120: sample text</div>
 <div id="ta-wk" style="-webkit-text-size-adjust:150%;font-size:14px;height:30px;width:300px;overflow:hidden">ta-wk: sample text</div>
 <div id="ta-auto" style="text-size-adjust:auto;font-size:14px;height:30px;width:300px;overflow:hidden">ta-auto: sample text</div>
 <div id="ta-btn" style="display:inline-block;margin-top:6px;padding:4px 10px;border:1px solid #94a3b8;background:#f8fafc;border-radius:4px;font-size:13px">toggle text inflation</div>`
	navSnipDrag = `<div id="drag-area" style="position:relative;width:420px;height:150px;background:#eef2f7">
  <div id="drag-win" style="position:absolute;left:20px;top:20px;width:200px;height:90px;background:#fff;border:1px solid #64748b">
   <div id="drag-bar" style="window-drag:move;height:24px;background:#cde"><div id="drag-btn" style="window-drag:none;display:inline-block;width:50px;background:#fca5a5">close</div><div id="drag-title" style="display:inline-block">title</div></div>
   <div id="drag-body" style="padding:6px">window body</div>
  </div>
 </div>`
	navSnipLayer = `<div id="wc-area" style="position:relative;height:90px;width:420px;background:#f1efe7">
  <div id="wc" style="will-change:transform, opacity;position:absolute;left:10px;top:8px;width:90px;height:30px;background:#6366f1;color:#fff">will-change</div>
  <div id="wc2" style="will-change:margin;position:absolute;left:10px;top:52px;width:90px;height:30px;background:#f59e0b;color:#fff">margin hint</div>
  <div id="wc-badge" style="position:absolute;left:120px;top:6px;width:290px;height:36px;padding:2px 4px;background:#fff;font-size:12px">wc: click the box to move it</div>
  <div id="wc2-badge" style="position:absolute;left:120px;top:50px;width:290px;height:36px;padding:2px 4px;background:#fff;font-size:12px">wc2: click the box to move it</div>
 </div>`
	navSnipAppear = `<button id="ap-none" style="-webkit-appearance:none;margin-right:10px;` + navAppearBox + `">plain button</button>
 <div id="ap-btn" style="-webkit-appearance:button;display:inline-block;margin-right:10px;` + navAppearBox + `">div as button</div>
 <button id="ap-def" style="` + navAppearBox + `">default button</button>`
	navSnipImage = `<div id="ia-paused" style="image-animation:paused;display:inline-block;width:70px;height:46px;line-height:46px;text-align:center;color:#fff;margin-right:8px;background:#6366f1">1</div>` +
		`<div id="ia-stopped" style="image-animation:stopped;display:inline-block;width:70px;height:46px;line-height:46px;text-align:center;color:#fff;margin-right:8px;background:#6366f1">1</div>` +
		`<div id="ia-normal" style="image-animation:normal;display:inline-block;width:70px;height:46px;line-height:46px;text-align:center;color:#fff;margin-right:8px;background:#6366f1">1</div>` +
		`<div id="ia-btn" style="display:inline-block;padding:4px 10px;border:1px solid #94a3b8;background:#f8fafc;border-radius:4px;font-size:13px">Pause normal</div>`
)

// navFieldBox is one single-line text field whose text the session edits.
func navFieldBox(id, decl string, width int) string {
	return fmt.Sprintf(`<div id="%s" style="%s;%swidth:%dpx;padding:2px 6px">hello</div>`, id, decl, navFieldLook, width)
}

func navSnipCaretAll() string {
	return navCaretCell("cr-bar", "caret-shape:bar") + navCaretCell("cr-block", "caret:block manual #d00000") +
		navCaretCell("cr-under", "caret-shape:underscore;caret-animation:auto") + navCaretCell("cr-autoblock", "caret-shape:block")
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func navSnipSliderAll() string {
	return navSlider("sl-h", "left-to-right", 200, 12) + navSlider("sl-rtl", "right-to-left", 200, 12) + navSlider("sl-v", "bottom-to-top", 12, 60)
}

// navDemoHTML is every example on one page, for the tests and the live
// session's id inventory.
//
//nolint:gochecknoglobals // pixel sizes, timings and markup of the live demo
var navDemoHTML = navSnipFocus + navArrows("focus") + navLiveChip("chip-focus", "") +
	navSnipGrid + navArrows("grid") + navLiveChip("chip-grid", "") +
	navSnipContain + navArrows("contain") + navLiveChip("chip-contain", "") +
	navSnipAction + navArrows("action") + navLiveChip("chip-action", "") +
	navSnipCaretAll() + navSnipField + navSnipSecurity + navSnipSelect + navSnipAdjust +
	navSnipSliderAll() + navSnipDrag + navSnipLayer + navSnipAppear + navSnipImage

const navHeadCSS = `<style>.res{font-family:monospace;font-size:12px;color:#0f5132;margin:3px 0}` +
	`.chip{display:inline-block;padding:2px 8px;margin:2px 4px 2px 0;border-radius:10px;background:#dbeafe;font-family:monospace;font-size:12px}` +
	`.chip.hit{background:#bbf7d0}.chip.no{background:#fecaca}</style>`

// navFocusable marks the buttons as the items that can hold focus.
func navFocusable(it navItem) bool { return it.Tag == "button" }

func resLine(format string, args ...any) string {
	return `<div class="res">` + html.EscapeString(fmt.Sprintf(format, args...)) + `</div>`
}

func chip(class, text string) string {
	return `<span class="chip ` + class + `">` + html.EscapeString(text) + `</span>`
}

// navShow lays the snippet out alone and hands its model to show. The snippet
// is drawn first, then whatever show returns.
func navShow(snippet string, show func(m navModel) string) string {
	m, err := navLayout(context.Background(), `<!DOCTYPE html><html><body style="margin:0;font-size:14px">`+snippet+`</body></html>`, navDemoWidth)
	if err != nil {
		return "<p>layout failed: " + html.EscapeString(err.Error()) + "</p>"
	}

	return navInitial(snippet, m) + show(m)
}

// navInitial writes the state the live session starts in into the authored
// markup: the password bullets, the slider thumbs, the first image frames, the
// carets at the end of their text, the flat or drawn buttons and each demo's
// status line. The still picture then shows what the session shows when the
// tab opens, and the readouts under it describe that same state.
func navInitial(snippet string, m navModel) string {
	now := time.Unix(0, 0)
	s := newNavSession()
	s.reset(snippet, now)

	sc := &navScreen{base: s.base, out: map[string]navView{}}
	s.draw(sc, m, now)

	for id, v := range sc.out {
		// A focus ring is not authored style: the session builds styles on the
		// markup's own, so a ring written into it would never come off.
		if strings.HasSuffix(v.Style, ringStyle) {
			v.Style = ""
		}

		snippet = navPatchMarkup(snippet, id, v)
	}

	return snippet
}

// navPatchMarkup sets the style attribute, and for a text leaf the text, of
// the element with this id.
func navPatchMarkup(doc, id string, v navView) string {
	re := regexp.MustCompile(`(<[a-zA-Z]+[^>]*\sid="` + regexp.QuoteMeta(id) + `"[^>]*>)([^<]*)`)

	return re.ReplaceAllStringFunc(doc, func(match string) string {
		sub := re.FindStringSubmatch(match)
		tag, text := sub[1], sub[2]

		if v.Style != "" {
			tag = reStyle.ReplaceAllString(tag, ` style="`+v.Style+`"`)
		}

		if v.HasText {
			text = html.EscapeString(v.Text)
		}

		return tag + text
	})
}

type navKey struct {
	from string
	dir  navDir
	name string
}

// moveChips draws one chip per key press: "from key -> target". A target that
// exists is green, no target is red.
func moveChips(m navModel, keys []navKey) string {
	var b strings.Builder

	for _, k := range keys {
		r := m.Move(m.ByID(k.from), k.dir, navFocusable)

		switch {
		case r.Action == "scroll":
			b.WriteString(chip("hit", fmt.Sprintf("%s %s: scroll %g,%g", k.from, k.name, r.ScrollDX, r.ScrollDY)))
		case r.Target >= 0:
			b.WriteString(chip("hit", fmt.Sprintf("%s %s -> %s (%s)", k.from, k.name, m.Items[r.Target].ID, r.Via)))
		default:
			b.WriteString(chip("no", fmt.Sprintf("%s %s -> nothing", k.from, k.name)))
		}
	}

	return `<div>` + b.String() + `</div>`
}

// navHint is the line that tells a person what to try.
func navHint(text string) string {
	return `<div style="color:#475569;font-size:13px;margin:0 0 6px 0">Try it: ` + html.EscapeString(text) + `</div>`
}

func navDemoFocus() string {
	return navHint("click a button to focus it, then press the arrow keys or click Up, Down, Left, Right.") +
		navShow(navSnipFocus+navArrows("focus")+navLiveChip("chip-focus", ""), func(m navModel) string {
			return moveChips(m, []navKey{{"n-a", navRight, "right"}, {"n-a", navDown, "down"}, {"n-b", navLeft, "left"}, {"n-e", navUp, "up"}}) +
				resLine("At rest: n-a down jumps to #n-f (nav-down), over the nearer n-d. n-a right has no value, so the nearest button wins.")
		})
}

func navDemoGrid() string {
	return navHint("focus gn-a or gg-a, then press Down. Normal (top box) reaches gn-c; grid (bottom box) needs the same column.") +
		navShow(navSnipGrid+navArrows("grid")+navLiveChip("chip-grid", ""), func(m navModel) string {
			return moveChips(m, []navKey{{"gn-a", navDown, "down"}, {"gg-a", navDown, "down"}}) +
				resLine("At rest: gg-c is not in the column of gg-a, so grid finds nothing.")
		})
}

func navDemoContain() string {
	return navHint("focus c-b and press Down: o-x is right below but outside the container, so focus stays.") +
		navShow(navSnipContain+navArrows("contain")+navLiveChip("chip-contain", ""), func(m navModel) string {
			return moveChips(m, []navKey{{"c-a", navRight, "right"}, {"c-b", navDown, "down"}}) +
				resLine("At rest: o-x sits right below c-b but outside the container, so down stops.")
		})
}

func navDemoAction() string {
	return navHint("press an arrow key or a button: the box scrolls its sheet by 40px. Up and Left bring it back, the ends hold, and focus stays on s-a.") +
		navShow(navSnipAction+navArrows("action")+navLiveChip("chip-action", ""), func(m navModel) string {
			return moveChips(m, []navKey{{"s-a", navDown, "down"}, {"s-a", navRight, "right"}}) +
				resLine("At rest: action scroll turns a key into a %d px scroll and focus stays on s-a.", scrollStep)
		})
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func navDemoCaret() string {
	return navHint("click a cell, type letters, press Backspace. The three automatic carets blink every 500ms; the red manual one stays on.") +
		navShow(navSnipCaretAll()+navLiveChip("chip-caret", ""), func(m navModel) string {
			var b strings.Builder

			for _, id := range []string{"cr-bar", "cr-block", "cr-under", "cr-autoblock"} {
				c := navCaretOf(m.Items[m.ByID(id)].Declared)
				x, y, w, h := c.Rect(0, 0, 9, 16)
				b.WriteString(resLine("%s: %s, manual=%v, rect=%g,%g,%g,%g", id, c.Shape, c.Manual, x, y, w, h))
			}

			return b.String()
		})
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func navDemoField() string {
	return navHint("click a field, type letters, Backspace deletes. The content field grows; the fixed one keeps 150px.") +
		navShow(navSnipField+navLiveChip("chip-field", ""), func(m navModel) string {
			var b strings.Builder

			for _, id := range []string{"fs-content", "fs-fixed"} {
				d := m.Items[m.ByID(id)].Declared
				b.WriteString(resLine("%s: field-sizing=%s, width of 'hello world' = %.0f", id, d["field-sizing"],
					navFieldWidth(d["field-sizing"], "hello world", liveCharW, navFieldPad, 40, 150)))
			}

			return b.String()
		})
}

func navDemoSecurity() string {
	return navHint("click a field and type: auto shows bullets, none shows the letters.") +
		navShow(navSnipSecurity+navLiveChip("chip-security", ""), func(m navModel) string {
			var b strings.Builder

			for _, id := range []string{"pw-auto", "pw-none"} {
				d := m.Items[m.ByID(id)].Declared
				b.WriteString(chip("hit", navShownText("password", d["input-security"], "hunter2")))
				b.WriteString(resLine("%s as a password field shows the text above", id))
			}

			return b.String()
		})
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func navDemoSelect() string {
	return navHint("click on the word 'word' in each line: text and auto select the word, all selects the line, none selects nothing.") +
		navShow(navSnipSelect+navLiveChip("chip-select", ""), func(m navModel) string {
			var b strings.Builder

			for _, id := range []string{"us-text", "us-none", "us-all"} {
				lo, hi := navSelect(m.Items[m.ByID(id)].Declared["-webkit-user-select"], "pick a word here", 7)
				b.WriteString(resLine("%s: a click on 'word' selects [%d,%d)", id, lo, hi))
			}

			return b.String()
		})
}

func navDemoAdjust() string {
	return navHint("click 'toggle text inflation': none stays 14px, a percentage scales, auto uses the host factor 1.5.") +
		navShow(navSnipAdjust+navLiveChip("chip-adjust", ""), func(m navModel) string {
			return resLine("At rest every line is 14px; the host auto factor is 1.5.")
		})
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func navDemoSlider() string {
	return navHint("press on a track and drag: the thumb follows and the value counts up in the direction the property names.") +
		navShow(navSnipSliderAll()+navLiveChip("chip-slider", ""), func(m navModel) string {
			var b strings.Builder

			for _, id := range []string{"sl-h", "sl-rtl", "sl-v"} {
				it := m.Items[m.ByID(id)]
				v := navSliderValue(it.Declared["slider-orientation"], it.X, it.Y, it.W, it.H, it.X+it.W*0.25, it.Y+it.H*0.25, 0, 100)
				b.WriteString(chip("hit", fmt.Sprintf("%s: %.0f", id, v)))
			}

			return b.String() + resLine("At rest: pointer at a quarter of the width and a quarter of the height")
		})
}

func navDemoDrag() string {
	return navHint("press on the title bar and move the pointer: the window follows. The red close box and the body do not drag.") +
		navShow(navSnipDrag+navLiveChip("chip-drag", ""), func(m navModel) string {
			return chip("hit", fmt.Sprintf("press on title drags: %v", m.navDragRegion(m.ByID("drag-title")))) +
				chip("no", fmt.Sprintf("press on close drags: %v", m.navDragRegion(m.ByID("drag-btn"))))
		})
}

func navDemoLayer() string {
	return navHint("will-change tells the host what is about to change. For transform or opacity the host keeps the box as a cached layer: it paints the box once and only slides that picture. Click each box: it moves for one second, and the note beside it counts how many times the box had to be painted.") +
		navShow(navSnipLayer, func(m navModel) string {
			props, promote := navLayerHints(m.Items[m.ByID("wc")].Declared["will-change"])

			return chip("hit", strings.Join(props, " + ")) + resLine("keep a cached layer: %v", promote)
		})
}

func navDemoAppear() string {
	return navHint("press and hold each control: none stays flat, button draws the control look and a pressed look.") +
		navShow(navSnipAppear+navLiveChip("chip-appear", ""), func(m navModel) string {
			return chip("hit", "ap-none: "+navControlKind("button", "", m.Items[m.ByID("ap-none")].Declared["-webkit-appearance"])) +
				chip("hit", "ap-btn: "+navControlKind("div", "", m.Items[m.ByID("ap-btn")].Declared["-webkit-appearance"]))
		})
}

//nolint:mnd // pixel sizes, timings and markup of the live demo
func navDemoImage() string {
	return navHint("watch the frames cycle every 200ms. Click 'Pause normal' to freeze the last box, again to resume.") +
		navShow(navSnipImage+navLiveChip("chip-image", ""), func(m navModel) string {
			var b strings.Builder

			for _, id := range []string{"ia-paused", "ia-stopped", "ia-normal"} {
				mode := m.Items[m.ByID(id)].Declared["image-animation"]
				b.WriteString(chip("hit", fmt.Sprintf("%s: frame %d", id, navImageFrame(mode, 0, 450, imgFrameMs, imgFrames)+1)))
			}

			return b.String() + resLine("At rest: a 4 frame gif, 200ms a frame, at 0ms. The paused box holds the frame it was paused on (450ms, frame 3); the stopped and normal boxes show frame 1, and normal then plays.")
		})
}

// navDemoWidth is the width the examples are laid out at when the host reads
// them alone. It is close to the width the demo box has in a 1000px window.
const navDemoWidth = 900

// Sizes of the scroll-action demo: a viewport that shows part of a larger
// sheet. The sheet is what scrolls, so Up undoes Down and Left undoes Right.
const (
	navViewW  = 316
	navViewH  = 100
	navSheetW = 700
	navSheetH = 400
	navCellW  = 80
	navCellH  = 50
)

// navActionHTML is a clipped viewport onto a sheet of labelled cells. Each
// cell names its own x,y position on the sheet, so a scroll is easy to see.
func navActionHTML() string {
	var b strings.Builder

	fmt.Fprintf(&b, `<div id="sbox" style="spatial-navigation-action:scroll;position:relative;margin-bottom:20px;height:%dpx;width:%dpx;`+
		`background:#f7f3e3;overflow:hidden"><div id="s-content" style="position:relative;width:%dpx;height:%dpx">`,
		navViewH, navViewW, navSheetW, navSheetH)

	for gy := 0; gy < navSheetH/navCellH; gy++ {
		for gx := 0; gx < (navSheetW-navCellW)/navCellW+1; gx++ {
			if gx == 0 && gy == 0 {
				continue
			}

			fmt.Fprintf(&b, `<div style="position:absolute;left:%dpx;top:%dpx;width:62px;height:26px;padding:5px 7px;background:#e7e0c8;`+
				`font-family:monospace;font-size:11px">%d,%d</div>`, gx*navCellW, gy*navCellH, gx*navCellW, gy*navCellH)
		}
	}

	b.WriteString(`<button id="s-a" style="position:absolute;left:0;top:0;width:76px;height:36px;box-sizing:border-box">s-a 0,0</button></div></div>`)

	return b.String()
}
