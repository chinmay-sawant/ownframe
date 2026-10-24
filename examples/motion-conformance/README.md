# CSS motion conformance

`motion-conformance` is both the manual comparison page and the page the tests
sample for the animation and transition property classes the automatic driver
ships. Every card is CSS in `index.html`; `page.go` builds the page and publishes
the three transition class toggles. The example registers no tick callback,
builds no player, uses no manual animation package, and imports no `internal`
package. OwnFrame binds the animations at the initial `Redraw`, captures each
transition when a toggle commits, and samples both from the window tick
([css-motion.md](../../documentation/css-motion.md)).

The three transition cards add one button each. A click flips a boolean and
publishes the card's class with one `SetData` and `Redraw`, the same class-toggle
shape the css-preview catalog uses. The transition itself is the stylesheet's
`transition` declaration; the handler writes no paint.

`Config.RetainState` is on from the first layout. The 3D and perspective cards
need the retained box tree for the re-projection pass; a page that starts
without it pays one owed redraw before the first 3D sample moves.

## Cards

| Card | Declarations | Sampled write |
|---|---|---|
| Timing functions | `animation-timing-function`: `linear`, `ease-in-out`, `cubic-bezier(...)`, `steps(5, jump-end)` | 2D matrix |
| Direction | `animation-direction`: `normal`, `reverse`, `alternate` | 2D matrix |
| Delay | `animation-delay`: `0s`, `600ms`, `-600ms` | 2D matrix |
| Fill mode | `animation-fill-mode`: `none`, `forwards`, `both` | 2D matrix |
| Iteration count | `animation-iteration-count`: `1`, `3`, `infinite` | 2D matrix |
| 2D translate | `transform: translateX()` | 2D matrix |
| 2D scale | `transform: scale()` | 2D matrix |
| 2D rotate | `transform: rotate()` | 2D matrix |
| 2D skew | `transform: skewX()` | 2D matrix |
| 2D matrix | `transform: matrix()` | 2D matrix |
| 3D rotateX | `transform: rotateX()` | re-projection |
| 3D rotateY | `transform: rotateY()` | re-projection |
| 3D translateZ | `transform: translateZ()` | re-projection |
| 3D matrix3d | `matrix3d()` animates its translation under an animated `perspective` | re-projection |
| Perspective | `perspective` | re-projection |
| Perspective origin | `perspective` and `perspective-origin` | re-projection |
| Opacity | `opacity` | per-operation element opacity |
| Colour and currentColor | `background-color` on a card with a label, and `color: currentColor` on a second card | fill channels, text channels |
| Border and outline colour | `border-color` and `outline-color` | stroke operations |
| Corner-shape path | `background-color` on a `corner-shape: bevel` box | path fill only |
| Paused and playing | `animation-play-state` | 2D matrix |
| Transition: background-color | class toggle with `transition: background-color 2s linear` | fill channels |
| Transition: opacity | class toggle on a wrapper with overlapping children | subtree paint opacity |
| Transition: transform | class toggle with `transition: transform 2s linear` | 2D matrix |

The 3D and perspective cards carry a perspective on their stage, so the
projection is visible rather than an edge-on or orthographic flattening. The
corner-shape card is the engine's bevel path: one fill op and one border stroke
op share the element, and a background keyframe must write the fill half only.

The opacity card puts two demo chips on top of each other inside the element
whose opacity moves, so a sample that fades each chip separately is visible as a
different overlap from one group fade. The transform card translates the tile
through the operation matrix, not through its laid-out position.

The matrix3d card animates its translation under an animated camera. The
canonical transform list writes all sixteen matrix3d arguments unitless, with
the translation components already in display points, so the restamp parser
accepts the list and reprojects the subtree.

## What the tests assert

`page_test.go` builds the page, advances two ticks through `Page.Testing().Tick`,
and checks:

- The page binds players with no tick callback, no bind error, and retained
  layout state from the 3D cards.
- The 2D translate card writes a matrix and moves again on the second tick.
- A 3D card restamps on both ticks (the applied request in `Restamped`), and
  its projection changes on the second tick. The `NeedsRestamp` sample counter
  is reset by the rebind that adopts the re-projected display.
- The background-colour card changes its fill and leaves its label's text
  operation untouched.
- The corner-path card recolours the path fill and keeps the border stroke and
  the path command list.
- The paused card keeps its translation across the ticks while the playing card
  advances.
- The three transition cards bind from their buttons alone: the page reports
  three bound transitions, and a sample part-way through the run has a colour,
  an opacity, and a matrix translation strictly between the committed endpoints.
  A snap would land on an endpoint.
- `Window().SetTransitions(false)` releases the live renderers and freezes the
  presented paint across a tick; `Transitions()` reports the switch.
  `SetTransitions(true)` lets the next card toggle capture again.

## Deliberately absent

These property classes are not on this page because they belong to another
phase or track:

| Class | Owner | Reason |
|---|---|---|
| Layout keyframes (`width`, `height`, insets, margin, padding, `font-size`, `line-height`, track lists) | Animation phase 04 | A size or inset change needs a relayout, not a retained paint write. |
| Composited layers (`filter`, `backdrop-filter`, `clip-path`, mask position and size, group opacity) | Animation phase 08 | The current path writes per-operation opacity and does not composite an animation layer. |
| SVG paint host (`fill`, `stroke`, `stroke-width`, `stroke-dasharray`, `stroke-dashoffset`) | Animation phase 10, engine phase 11 | Inline SVG serializes to an image operation; the engine needs `RestyleInlineSVG`. |
| Transitioned geometry (`width`, `height`, `top`, `left`) | Transition phase 02 | A layout length needs the relayout override, not a retained paint write. |
| Discrete transitions (`display`, `visibility`, `content-visibility`, `position`, `z-index`, `order`, grid lines) | Transition phases 02 and 04 | The endpoint swap and the outgoing layer are later phases. `@starting-style` is not implemented. |
| A `display` cross-fade | Transition phase 04 | The outgoing element has no operations in the new display, so the change snaps until the page keeps an outgoing layer. |

`transform-origin` is not a keyframed property. A keyframe that declares it
writes nothing, and the runtime uses the element's laid-out origin, so the 2D
cards here use the default centre origin.

## Running

```sh
go run ./examples/motion-conformance
```

Serve the same picture in a browser:

```sh
go run ./examples/motion-conformance -web
```

`Serve` does not tick, so `-web` shows the committed frame rather than moving
cards. Use the desktop window to watch the motion. Pass `-addr` to change the
listen address:

```sh
go run ./examples/motion-conformance -web -addr=127.0.0.1:8143
```

## Testing

```sh
cd examples && go test -count=1 ./motion-conformance/...
```
