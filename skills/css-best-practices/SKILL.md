---
name: css-best-practices
description: >
  Write and review template CSS in this repo. Use when text alignment looks
  off, a row or bar is not centered, flex or grid wraps or overflows, spacing
  drifts, hover or dark mode breaks, or a template change needs a second pass
  before commit. Covers the left-aligned plus vertically centered pattern
  agents miss, with copy-paste rules and a verification checklist. Does not
  cover engine layout internals. For a visibly wrong render with unknown cause,
  use debug-html-template first.
---

# CSS best practices

Write CSS the template already honors. Most misses in this repo were one missing property in a known trio, not a new feature.

## How to use this skill

1. Pin the fragment. Open the `.html` and the `.css` that style it. Quote the used values for `display`, `width`, `height`, `line-height`, `text-align`, `margin`, `gap`, and flex properties.
2. Match it to one rule below. Apply the smallest change that satisfies the rule.
3. Verify with the checklist at the end. If the CSS is correct and the output is still wrong, stop and treat it as an engine gap.

## Hard boundaries

- Change one fragment at a time.
- Do not rewrite shared chrome to fix one row.
- Do not touch engine code because a template rule is incomplete. Fix the template first.
- Do not widen a selector or add `!important` to force a win. Raise specificity by one concrete class instead.

## Left plus vertical center

This is the miss that started this skill. Text should sit left with vertical center alignment. The agent left default alignment in place.

Apply all three. One property alone fails.

```css
.left-center {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  text-align: left;
}
```

For a single-line fixed-height row, set line height to match height minus borders.

```css
.badge { height: 32px; line-height: 30px; text-align: left; padding: 0 12px; }
```

 Concrete cases from code-view:

- Picker rows in `examples/code-view/components/chrome.css`. `.project-input` had `height: 28px` with no line height, so text sat off center. The fix adds `line-height: 28px; text-align: left`.
- Filter bar in `examples/code-view/components/sidebar.css`. The 28px bar holds a 12px icon next to an input with `line-height: 26px`. The icon needs `align-self: center; flex: 0 0 0 auto` or it keeps a baseline gap.
- Search meta showing "0 files / 0 results". Same trio missing. Add `line-height` plus `text-align: left` plus `align-items: center`.
- Markdown bar in `examples/code-view/components/editor.html`. One flex row held the label and the icon group with no split, so the path wrapped. Add a fill spacer between the left crumb and the right control cluster, the same pattern the preview buttons use.

Horizontal is `text-align: left` for inline content or `justify-content: flex-start` for flex. Vertical is `align-items: center` for flex, `line-height` equal to height for single lines, or `vertical-align: middle` for inline icons only. `vertical-align` does nothing on flex items.

## Flex rules

- Name the container. A child with `flex: 1` and no `display: flex` on its parent cannot center its own children. Give the inner wrapper its own `display: flex; flex-direction: column; align-items: center` when it owns the stack.
- Pair axis and inline alignment. `flex-direction: column; align-items: center` needs `text-align: center` on the text, or blocks center while words stay left. The reverse holds for left stacks.
- State wrap intent. Default `flex-shrink: 1` squeezes chips to min content and wraps labels to two lines. Use `flex-shrink: 0` on fixed chips, or add `flex-wrap: wrap` when wrap is the design.
- Keep full-width rows full width. A `width: 80%` row next to full-width siblings looks lopsided. Use `width: 100%` and control rhythm with the bottom margin.
- Center fixed overlays on both axes. `position: fixed; inset: 0` alone pins the box to the viewport without centering. Add `display: flex; align-items: center; justify-content: center; padding: 32px`.

```css
/* Before: short row packed left */
.flow-row { display: flex; justify-content: flex-start; }

/* After: short row centered */
.flow-row { display: flex; justify-content: center; }
```

```css
/* Before: icon above left-aligned text */
.flow-node { display: flex; align-items: flex-start; text-align: left; }

/* After: stacked and centered */
.flow-node { display: flex; flex-direction: column; align-items: center; text-align: center; }
```

## Grid and display rules

- Give grids an explicit collapse. `repeat(2, 1fr)` needs a `@media (max-width: 640px)` fallback to `1fr`.
- Protect fluid tracks. `grid-template-columns: 212px 1fr` overflows on long content. Write `212px minmax(0, 1fr)`.
- Center single children with grid. A fixed 28px icon box reads better as `display: grid; place-items: center` than a three-property flex.
- Size label plus value rows as fluid plus fixed. Replace fixed `130px 1fr 34px` with `1fr 72px 28px`, add `align-items: center`, and put ellipsis on the label.
- Prefer `repeat(auto-fill, minmax(240px, 1fr))` over manual breakpoints for card grids.

## Sizing and overflow rules

This engine honors `overflow: visible` and `hidden` only. There is no element scroller.

- Put `overflow: hidden` on the block row, never on an inline `b` or `span`. Inline clip never fires.
- Give ellipsis a definite width. `flex: 1; min-width: 0` plus the nowrap trio still paints past the box here when the engine sizes content to its own width. A fixed width such as `width: 200px` truncates. In code-view, long crumbs clip in Go (`clipDiffPath` in `diff.go`) because inline runs split on spaces regardless of `white-space: nowrap`.
- Start every stylesheet with `*, *::before, *::after { box-sizing: border-box; }`. `width: 100vw` plus padded children overflows without it.
- Bound lists with `max-height` instead of inner scroll. Let the window scroll the page. Keep `.palette-list` and `.search-list` bounded and let pre blocks use `max-width: 100%; overflow: auto`.
- Drop margins that fight shrink. A 5px left margin on a flex sibling plus extra flex basis defeats `min-width: 0`. Remove the margin on the shrinking row.

## Spacing rules

- One parent owns rhythm. Replace per-child `margin-top: 3px` stacks with one parent `gap: 3px` plus `align-items: center`.
- Reset lists before styling them. UA `ul` ships with 40px left padding. Write `list-style: none; padding: 0; margin: 0` and space with parent `gap`.
- Zero outer edges on divider columns. Uniform `padding: 0 1.1rem` double counts at the container edge. Add `:first-child { padding-left: 0 }` and `:last-child { padding-right: 0 }`.
- Space siblings only. Margin on every `p` doubles the parent gap above the first child. Use `p + p { margin-top: 16px }`.
- Adjacent vertical margins collapse to the larger one, not the sum. Resolve UA margins before reasoning about the gap.

## Typography and theme rules

- Expand `font` shorthand. The cascade needs size, line height, family, weight, and style as separate longhands or the declaration drops silently.
- Resolve weight to a real face. Weight 700 and above must select a bold face from the registry. Synthetic stroke is fallback only.
- Use fixed pixel `line-height` for vertical centering in fixed rows. Badge, nav, and avatar rows in telegram and spotify examples use 20px, 40px, and 160px lines for this reason.
- Paint the caret with `currentColor` and keep it above glyphs.

```css
.caret { border-left: 1px solid currentColor; z-index: 1; }
```

- Treat theme sheets as order-sensitive. `Config.Theme` and `Page.SetTheme` add one sheet walked last, but rule indices restart at zero per sheet, so an equal-specificity tie can still go to a late template rule. Theme `!important` beats normal template rules. Inline `style=` beats normal theme rules. `SetTheme` does not draw, so call `Redraw` after.
- Build dark mode as a variable swap under `body.dark` or `data-theme` (`--bg`, `--panel`, `--ink`, `--sub`, `--line`, `--accent`) plus a persisted bool and toggle. Keep status readable without relying on color alone.

## Positioning rules

- Anchor absolute children to a positioned parent with nonzero size. A zero-size relative parent drops its absolute descendants.
- Use sticky for in-flow pinning and fixed for viewport pinning. Sticky clamps where the containing block meets the page. Fixed stamps on every page. A chat top bar that must stay put while messages scroll needs the containing block, the `top` inset, and the right scrollport.
- Shift relative boxes and their paint. `position: relative` with `top` and `left` must move both the box and its emitted ops, or the box looks static and its absolute children anchor to the wrong rect.
- Order decorative layers explicitly. Give glows `position: absolute; z-index: 0; pointer-events: none` and content `position: relative; z-index: 1`.

```css
.hero-section > * { position: relative; z-index: 1; }
.hero-section::before {
  content: ""; position: absolute; z-index: 0; pointer-events: none;
}
```

- Give pinned chrome all four pieces: `sticky; top: 0; z-index` plus an opaque background and border. Without the background, scrolled text shows through.

## Border and background rules

- Keep radius and clip together. `border-radius: 12px; overflow: hidden` on a card clips a square `border-left` accent round. In article mode, strip background, border, radius, shadow, and overflow together, then re-add the spine with `border-radius: 0`.
- Use `background-color`, not `background` shorthand. The cascade handles the longhand only. Image and gradient layers stay ignored.
- Expect one `box-shadow` layer. The first non-inset layer paints. Inset and later layers drop. Blur renders as stacked fills, not a Gaussian, and spread grows the shadow box without changing layout.

## Hover and motion rules

- Guard disabled hover. Add `pointer-events: none; cursor: default` on `.disabled` or hover styles keep matching.
- Keep `::after` overlays out of the pointer path. An `inset: -1px` glow covering a button needs `pointer-events: none`.
- Scope transitions. `transition: all` animates layout and paint. List `color, border-color, transform, opacity` instead.
- Pair every hover with focus. Hover never fires for keyboards. Add `:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px }`.
- Transition nested icon opacity separately. A parent color transition does not un-mute a child with `opacity: 0.6`.
- Promise clicks honestly. Actionable controls get `cursor: pointer`. Static names and disabled rows keep `cursor: default` even with hover color.
- Gate motion. Reduce entrances under `prefers-reduced-motion` and only animate otherwise.

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; }
}
```

## Responsive rules

- Support print and screen only in `@media`. Other types and unknown features evaluate false. Size features and orientation compare against the viewport.
- Re-evaluate on resize. Relayout, not repaint. A frozen placement leaves text unwrapped and viewport units stale.
- Ship the 900px collapse for multi-pane layouts. Stack to one column and hide the sidebar.
- Use the `.phone` class for phone layout, not a media query. The mobile view applies `.phone` plus inset top and bottom.

```css
.phone .grid { grid-template-columns: 1fr; }
.phone #q { max-width: none; }
```

- Clamp circles in flex. `border-radius: 50%` alone stretches into an ellipse when flex hands it extra room. Write `width: 50px; height: 50px; min-width: 50px; max-width: 50px; flex: none`.

## Template fidelity

- Let the template drive. Telegram thread and chat pages render from their `<style>` and display list, not from a bitmap fallback. Login inputs render from `login.html` (`.field` base, focus and selected states, placeholder text, 320px card), not from hand-built Go chrome. Hand-rolled geometry only creates pixel-test churn the template already solves.

## Verification checklist

Answer yes to each line before commit:

- Left text uses `text-align: left` or `justify-content: flex-start`, and vertical center uses `align-items: center` or a matching `line-height`. No single-property centering.
- The wrapper that centers is itself a flex or grid container.
- No `width: 80%` row sits next to full-width siblings without reason.
- Lists reset UA margin and padding. Rhythm comes from one parent `gap` or `p + p`.
- Ellipsis targets a block with a definite width. Inline clip and `overflow: auto` inner scroll are gone.
- Disabled controls carry `pointer-events: none` and `cursor: default`. Overlays carry `pointer-events: none`.
- Hover has a matching `:focus-visible` rule. Transitions name properties, never `all`.
- Theme changes call `Redraw`. Dark surfaces use `currentColor` where inheritance matters.
- Pinned bars have position, inset, `z-index`, and opaque background.
- `gofmt` is clean on touched Go files. `make test` passes for the touched module.
