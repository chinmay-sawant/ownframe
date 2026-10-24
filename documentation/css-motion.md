# CSS animations and transitions

Blinkless reads CSS animation and transition declarations and writes their
values and `@keyframes` to the drawing list. OwnFrame binds that data to retained
paint operations and samples it from the window tick. A page that declares an
animation or a transition and changes a style gets motion with no Go callback.

## Automatic by default

`New` starts both drivers on. Every display-list commit binds the animations of
the display it just installed, after the generation bump: `Redraw`, the display
a `Page.PatchElement` adopts, and a re-projected display all call the same hook
(`internal/page/motion_wire.go`, `bindMotion`; `internal/page/page_draw.go`,
`commitDisplay`; `internal/page/patch_wire.go`, `spliceElement`). A commit
that falls back to a bitmap releases the players instead. A zero `Config` keeps
the driver on, and no option is needed.

The bind resolves one player per element and paint class. It keys each player by
the element id and its ordinal, so a later commit matches the same box even when
a run of anonymous boxes shifts. A matched player keeps its original start time,
so a redraw, patch, hover, reload, or resize samples the same phase instead of
restarting (`internal/motion/bind.go`, `Bind`). The paint captured as the base
is the fresh layout's declared paint.

A page needs a named `@keyframes` rule, an `animation-name` or `animation`
declaration on the element (the `-webkit-animation` aliases normalize to the
same records), and the display-list paint path. A page that fell back to a
bitmap binds nothing and reports `Bound` zero.

`Page.Window().SetAnimations(false)` turns the driver off; `Animations()`
reports the state. The zero value is on. Turning it off releases the bound
players at once, so nothing writes through an operation a later publish moves;
the display keeps the paint of the last sample until the next layout, which
paints the declared values. Turning it back on lets the next publish or tick
bind again. There is no per-element switch.

The automatic driver runs keyframe animations, scroll- and view-driven
timelines, CSS animation triggers, and the transitions described below. A paused
or triggered player does not keep the page asking for frames by itself; a
timeline player changes when the page scrolls, and a triggered player waits for
an activation. A running transition asks for a frame until it finishes.

### Tick order

The window tick runs the driver before the `SetTick` callback:

1. step the live transition renderers, then the bound animation players, at
   `time.Now()`, write their values, and collect the damage and the 3D requests
   (`internal/page/motion_wire.go`, `stepMotion`; `internal/motion/step.go`,
   `Step`);
2. run the application callback, with the operation snapshot taken after step 1;
3. blink the caret;
4. diff the callback's operations into dirty rects;
5. apply the 3D or perspective requests in one engine pass, and pay an owed
   retained-state redraw if one is due (`restampMotion`);
6. set the frame-dirty flag.

Step 1 marks its damage before the callback snapshot, so the callback's diff
never covers an animation change. A tick with no animation, no callback, and no
focused field leaves the page clean.

While a time-driven player reports work, the driver asks for the next frame at
once (`WantFrameIn(0)`). A player driven only by a scroll, view, or named
timeline never asks: its pose is a function of the page state, so a still page
keeps its last sample. The desktop window therefore does not enter idle park
while a time player is active. `Serve` does not tick, so a served page shows its
committed frame and does not animate.

## Transitions

A transition runs from a style change the page already makes: a `SetData`
redraw, a `Page.PatchElement` class edit, a theme change, a hover, or a reload.
Before the commit swaps the old display away, the page reads the transition
records from the new boxes and compares the paint of each element with the same
id in the old display (`internal/page/motion_wire.go`, `captureTransitions`;
`internal/motion/transition.go`, `CaptureTransitions`). The bind installs one
renderer per element and property after the commit, with the commit time as the
start (`bindTransitions`; `BindTransitions`). The old display holds the paint the
runtime last sampled, so the start is the value on screen, not the declared
base.

The element id and its ordinal identify the old box. An element with no id, or
one that was not in the previous display, has no start paint and no transition;
the capture counts it. `transition-property: all` covers every property the
capture reads; the table below is the shipped part. A later declaration overrides
an earlier one.

### Shipped property set

| Declaration | Sampled write |
|---|---|
| `background-color`, `accent-color` | the element's own fill operations (`R`, `G`, `B`) |
| `color`, `text-decoration-color`, `text-emphasis-color` | the element's own text operations |
| border, outline, and rule colours | the element's own stroke and line operations |
| `opacity` | the sampled opacity on each captured subtree operation |
| `transform` with 2D functions | the operation's baked matrix |

Colour selection uses the same nearest-owner rule as a keyframe colour, and a
transition blends colour in sRGB. A colour with alpha writes the RGB channels
only; the alpha stays at the element's base paint. `transform` takes the 2D
matrix path, so a 3D list, a perspective, or an element whose static transform
already participates in 3D does not transition on this path.

`width`, `height`, `top`, and `left` belong to the next phase. So do the
discrete properties: `display`, `visibility`, `content-visibility`, `position`,
`z-index`, `order`, and the grid-line integers. An appearing or disappearing
`display` change cannot cross-fade in this phase: the outgoing element has no
operations in the new display and the page keeps no outgoing layer, so the
change snaps and `Probe().Motion().TransitionSkipReason` names it. Treat a
property outside the table as unclaimed; do not promise its motion in an app
until its phase lands.

### Reversal and precedence

A second style change while a run is live starts from the paint on screen. When
the new destination is the value the live run started from, the new run is a
reversal: the declared duration and a negative delay are scaled by the CSS
reversing shortening factor, so re-crossing part of the range takes the part of
the duration that remains. Any other new value starts a normal run with the full
declared duration. A change back to the value already on screen cancels the run
without a replacement, because the commit already painted the answer.

Transitions step before animations in the tick. An animation on the same field
writes second and wins, so a field an active, paused, or filling animation owns
is not bound, and a live entry whose field an animation starts to own is
cancelled. Do not drive one property with both a transition and a manual player.

A zero or negative duration has no frame to sample, so the committed paint
stands. A changed operation count or kind and a bind error snap to the committed
value; the capture counts the skip, or `TransitionError` names the failed bind.
The other properties still bind.

### Pacing and damage

A running transition asks for the next frame at once, a delayed run asks for the
delay end, and a finished or released run asks for nothing. Each sample marks
the union of the target operations' painted bounds before and after the write,
and the page feeds that into its ordinary dirty set. A page with only
transitions therefore parks like a static page once the run ends. `Serve` does
not tick, so a served page shows its committed frame.

### Switch

`Page.Window().SetTransitions(false)` turns the transition driver off;
`Transitions()` reports the state. The zero value is on and `Config` needs no
option. Turning it off releases every live renderer without writing, so the
paint the last sample presented stays on screen; turning it back on lets the
next style change capture. A style change made while the driver is off commits
directly. The animation switch is separate: either driver can run with the
other off, and a page with only transitions still asks for frames while a run
is active.

A page that steps `csstransition.Bind` or `BindTo` by hand turns this switch off
for the elements it drives. Both paths write the same retained operations.

## What a keyframe may write

Two paths carry the sampled values.

| Declaration | Sampled write | Path |
|---|---|---|
| `transform` with 2D functions | the operation's baked matrix (`SetXform`) | in place |
| `opacity` | the operation's element opacity (`SetPaintOpacity`) | in place |
| `background-color`, `accent-color` | the element's own fill operations (`R`, `G`, `B`) | in place |
| `color`, `text-decoration-color`, `text-emphasis-color` | the element's own text operations | in place |
| border, outline, and rule colours | the element's own stroke and line operations | in place |
| 3D `transform` lists, `perspective`, `perspective-origin` | the element's projection | engine re-projection |

The property map is `internal/cssanimation/keyframes.go`, `propertyClass`. A
target carries a field mask, so a transform player never writes colour and a
colour player never writes the matrix. A keyframed transform writes the matrix,
not the operation's laid-out X and Y. W, H, and alpha are not keyframe write
fields: a size or inset change is layout state, and a colour's own alpha stays
at the element's base paint. A keyframed colour with alpha is counted in `Gaps`
rather than painted as if the alpha moved (`internal/animation/compose.go`,
`applyTransform`; `internal/cssmotion/paint.go`, `Mask`).

A declaration the binder cannot place stays at the element's base paint and is
reported once per element and reason in `Gaps`. That covers `filter`,
`backdrop-filter`, `clip-path`, the mask position and size properties,
`mix-blend-mode`, and `isolation`; SVG paint properties such as `fill`, `stroke`,
and their width and dash properties; the individual `translate`, `rotate`, and
`scale` properties; and layout properties such as `width`, `height`, insets,
`font-size`, `line-height`, letter and word spacing, shadows, background
position and size, outline geometry, and radii
(`internal/cssanimation/keyframes.go`, `deferredPhase` and `layoutProperty`).

### Nearest owner

A colour writes only the operations whose nearest owner is this element
(`Display.OpOwners`). A parent's background colour cannot recolour a child's
card or its own label text; a text colour writes only the element's text
operations, and a border colour only its stroke and line operations. Transform
and opacity take the element's whole captured subtree, because a moved parent
moves its children. A box with no display index, or a display without owner IDs,
falls back to the box's subtree, and the bind reports that fallback
(`internal/cssmotion/ownership.go`, `Select`; `internal/cssanimation/bind.go`,
`BindBox`).

### Colour forms

A keyframe colour may be `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`, `rgb()`,
`rgba()`, `hsl()`, `hsla()`, a CSS named colour, `transparent`, or
`currentColor` (`internal/cssanimation/values.go`, `parseColor`). Every form
resolves to sRGB channels and blends in sRGB. `currentColor` resolves against
the element's own text operation; an element that paints no text makes that bind
error for its box, and the other boxes still bind.

Modern colour spaces are a gap, not an error: `color()`, `lab()`, `lch()`,
`oklab()`, `oklch()`, `color-mix()`, `light-dark()`, and `device-cmyk()` leave
the declaration at base paint and add a `Gaps` entry. The cascade keywords
(`inherit`, `initial`, `unset`, `revert`, `revert-layer`) are a bind error,
because a keyframe has no element context to resolve them.

## The 2D matrix path

A 2D transform sample composes into the operation's matrix:

```text
M_target = Parent * M_anim * World^-1 * Base
```

`M_anim` is the sampled transform with the element's transform origin baked in,
and `Base` is the matrix the operation carried at bind. The element's static
transform is replaced, not stacked, while any transform deeper than the element
survives (`internal/animation/compose.go`, `applyTransform`). The operation's
`X` and `Y` stay at their laid-out values; translation, rotation, scale, and
skew live in the matrix. Replay draws the operation through that matrix, so each
frame moves the element without a parse, cascade, or layout.

A transform list with any 3D function takes the sample off this path, even when
the current value is flat, such as `rotateY(0deg)`.

## 3D and perspective: the re-projection path

A 3D transform list, a perspective value, or an element whose static transform
already participates in 3D cannot become a 2D operation matrix. The player
leaves the operation's matrix unwritten for that sample and reports its sampled
values as a `RestampValue`. The motion runtime collapses those to one
`Request` per element, and the page applies every request in one
`layout.RestampTransforms` pass per tick (`internal/motion/step.go`, `Step`;
`internal/motion/restamp.go`, `Request`; `internal/page/motion_wire.go`,
`applyRestamp`). A 2D-only page collects none and never reaches the pass.

The pass needs the retained box tree that `Config.RetainState` keeps. A bind
that reports `NeedsRetainedState` turns retained state on by itself and owes one
redraw. The frame the flag lands on does not move: that redraw commits a
retained display, and the next tick re-projects. A page that already drew with
retained state re-projects on its first tick. A failure keeps the display the
page already draws and records `RestampError`; the next sample tries again. A
request needs an element id, so an anonymous box is counted in `NeedsRestamp`
but cannot be re-projected.

Adopting the re-projected display bumps the generation, rebuilds the hit list,
rebinds every player to the new operations, and marks the engine patch's old
and new painted boxes. The arrays are shared with the retained layout state, so
a comparison against an earlier frame needs a copy.

## Damage

Each sample reports the union of the target's painted bounds before and after
the write (`internal/motion/step.go`, `Step`). A moving transform therefore
damages where the ink left and where it landed, and the dirty-bounds helper
(`internal/page/dirty`, `OpBounds`) maps a transformed operation's local paint
box through its matrix. The page feeds the
rects into its ordinary dirty set, which coalesces them and falls back to the
frame when the list overflows. A paint-only sample (colour or opacity) reuses
the before rectangle. A re-projection marks the engine patch's old and new
boxes. The driver does not compare sampled paint values: a sampled player
marks its bounds even when the value repeats. Turning the driver off with
`Page.Window().SetAnimations(false)` releases the players and stops the damage.
The transition switch does the same for transitions:
`SetTransitions(false)` releases the live renderers and stops their damage.

## Scroll-driven timelines and triggers

`animation-timeline: scroll()` reads the page scroller progress:
`offset / (content - viewport)`, 0 at the top and 1 at the end. `view()` reads
the animated element's visibility in the scrollport: 0 when its leading edge
reaches the scrollport's trailing edge, 1 when its trailing edge passes the
leading edge. A named timeline (`animation-timeline: --name`) resolves the
nearest declaration of that name on the element or an ancestor's
`Box.Timelines`; the engine exports `scroll-timeline-*` and `view-timeline-*`
there. A view timeline's subject and its start and end insets come from that
declaration, and `view-timeline-inset` shrinks the scrollport before the
progress is computed.

The runtime builds one `animation.FrameInput` per sample
(`internal/motion/timeline.go`, `input`). It carries the scroll progress, the
element's view progress, the named timeline values, the concrete `entry`,
`cover`, `exit`, and `contain` ranges, and the source and target progress for
`animation-range-center`. The maps are allocated at bind and cleared in place,
so a live sample allocates nothing. A dashed name no ancestor declares, a named
element scroller this engine does not implement, a missing view subject, or a
scroller with no range leaves the animation inactive. There is no wall-clock
fallback; `Probe().Motion().UnresolvedTimelines` and `InactiveTimelines` count
the two cases.

`animation-trigger` fires on activation. A click or tap on the element or a
descendant, Enter or Space on the focused control, and the press/release pair
raise the dashed names the activated element and its ancestors declare; the
release applies the second action of each name. `internal/csstrigger` resolves
the scope and the action list, and the runtime aligns each box animation index
to the effect index of the player it actually built, because one element can
bind several players. A timeline-only player does not request frames; a
dispatch wakes the players for one sample so a reset or pause poses the next
frame.

Limits: nested element scroll containers and `timeline-scope` are not
implemented, so a name that needs either stays inactive; `scroll()` and `view()`
ignore their axis argument and read the block axis; a length in an
`animation-range` edge such as `entry 10px` is not resolved, so only keywords
and percentages sample; a trigger that pauses a timeline effect keeps its held
phase until the next rebind, which then recomputes from the current scroll
position, because the player clock carries pause and reset but not the held
timeline offset.

## Counters

`Page.Probe().Motion()` returns one read-only `MotionStats` snapshot
(`page_probe.go`; `internal/motion/stats.go`).

| Field | Meaning |
|---|---|
| `Bound` | players installed by the last bind; zero after the driver releases them |
| `Active` | players that reported work at the last `Step` |
| `Sampled` | players the last `Step` sampled |
| `NeedsRestamp` | sampled players whose value a 2D matrix could not carry |
| `NeedsRetainedState` | the last bind found motion that needs the re-projection pass |
| `Restamped` | elements the page re-projected at the last tick |
| `RestampError` | the last re-projection failure; empty after a success |
| `Rebinds` | cumulative binds that replaced at least one bound player |
| `LastError` | the last bind failure; a failed box does not stop the others |
| `Gaps` | fidelity gaps the last bind accepted, one entry per element and reason; empty after a release |
| `UnresolvedTimelines` | dashed timeline references the last bind could not resolve; those animations stay inactive |
| `InactiveTimelines` | players the last `Step` left inactive because a resolved timeline had no range at that geometry |
| `TransitionsBound` | transition renderers installed by the last bind; zero after a release |
| `TransitionsStarted` | new renderers the last bind started from a captured style change |
| `TransitionsActive` | renderers that reported work at the last step |
| `TransitionsSampled` | renderers the last step sampled |
| `TransitionsCancelled` | live renderers the last bind or step dropped |
| `TransitionsSkipped` | captures or live entries the last bind could not use, such as a changed operation set or a field an animation owns |
| `TransitionsBlocked` | property changes the last capture saw but could not transition because the display or engine cannot express the value yet |
| `TransitionSkipReason` | the first blocked change's reason; empty when nothing was blocked |
| `TransitionsNoID` | transition declarations the last capture skipped because the element has no id |
| `TransitionsNoOld` | declarations the last capture skipped because the element was not in the previous display |
| `TransitionError` | the last transition bind failure; that property snaps and the other properties still bind |

`Active`, `Sampled`, `NeedsRestamp`, `Restamped`, `InactiveTimelines`,
`TransitionsActive`, and `TransitionsSampled` cover the last tick. `Bound`,
`NeedsRetainedState`, `Gaps`, `UnresolvedTimelines`, `TransitionsBound`,
`TransitionsStarted`, `TransitionsBlocked`, and `TransitionSkipReason` cover the
last bind. `TransitionsSkipped`, `TransitionsNoID`, `TransitionsNoOld`, and
`TransitionError` cover the last capture and bind. A blocked or skipped change
is not a failed draw; it means that property snapped instead of moving. A struct
with no driver returns the zero value. There are no layer, byte, re-bake, or
demotion fields because the page path does not composite animation layers yet;
do not read one counter as a promise of another.

## Named limits

- **A 3D mismatch costs the next pass.** If the page did not start with
  `Config.RetainState`, the tick that first samples a 3D transform or a
  perspective spends its frame turning retained state on. The sample lands on
  the next tick. Starting with `RetainState: true` removes the delay.
- **Perspective implicit endpoints are `none`.** Every parsed frame carries the
  element's 50% 50% origin resolved on its border box, and a keyframe that omits
  `perspective` is padded with a `none` camera at that default origin
  (`internal/cssanimation/perspective.go`, `defaultPerspective`;
  `internal/animation/compose.go`, `padPerspectiveStops`). The host cannot read
  the element's static perspective, so an animation that declares a camera only
  from 50% to 100% starts at no perspective rather than at the declared base
  camera. A keyframe that omits `perspective-origin` uses that resolved default
  at that endpoint. A `none` endpoint is discrete: the distance flips at the
  midpoint of the interval, while the origin interpolates component-wise.
- **No layers yet.** `opacity` writes each targeted operation's element
  opacity. Two operations from separate branches of the element whose painted
  boxes overlap fade independently instead of compositing as one group; the
  bind reports that as a gap, and the overlap scan is capped at 128 operations
  per element (`internal/motion/bind.go`, `opacityGap`). The window does not
  composite an animation layer yet, so a moving element still pays the replay
  of its damaged operations.
- `transform-origin` is not a keyframed property. A keyframe that declares it
  writes nothing and adds no gap; the runtime uses the element's laid-out
  transform origin.
- A transform list with an unsupported function or unit fails the bind for that
  box, and the other boxes still animate. A malformed value is not silently
  snapped to base paint.

## Manual packages (preview)

These packages live under `internal`, so code outside the module cannot import
them. They are the preview and implementation path, not the application API.
They are also how a caller supplies a custom clock, timeline progress, or
trigger event.

`cssanimation.Bind` builds a player for a box and a full-paint target set, the
whole-subtree behaviour the property catalog drives by hand;
`cssanimation.BindBox` is the narrower binder the automatic runtime uses
(`internal/cssanimation/bind.go`). `animation.Player.Step` samples a player with
a caller-supplied time, and `StepWithInput` also accepts scroll progress, view
progress, named timelines, and range values (`internal/animation/player.go`).

An animation with a CSS trigger waits for an event. The page dispatches
activation events itself; a manual caller can send its own through
`csstrigger.Dispatch`, which applies matching actions to animation players
(`internal/csstrigger/dispatch.go`).

`csstransition.Bind` and `csstransition.BindTo` select a transition record and
build a renderer; the caller steps it with time
(`internal/csstransition/bind.go`, `internal/transition/step.go`). The automatic
driver uses the same packages. The manual entry points remain the preview path
for a custom clock or a property whose phase has not landed.

`viewtransition.New` pairs elements with the same view-transition name in old
and new displays inside the selected roots. `At` samples a progress value, and
`Frame.Display` combines the old and new paint layers with the new display
(`internal/viewtransition`). The caller provides both displays and the progress;
OwnFrame does not capture them from a Page update.

If a page drives a manual player over an element whose CSS already drives it,
turn the matching switch off first: `Page.Window().SetAnimations(false)` for an
animation and `Page.Window().SetTransitions(false)` for a transition. Both paths
write the same retained operations, and two drivers on one operation overwrite
each other. Each switch is page-wide; there is no per-element handoff.

## Preview and validation

The preview runs against the Blinkless commit pinned in OwnFrame's `go.mod`.
To test unreleased engine changes, create a temporary workspace with both local
checkouts, then run the preview. Replace the paths below with the absolute
checkout paths:

```sh
workspace_dir="$(mktemp -d)"
cd "$workspace_dir"
go work init /path/to/blinkless /path/to/ownframe /path/to/ownframe/examples
cd /path/to/ownframe
GOWORK="$workspace_dir/go.work" go run ./examples/css-preview
```

The example builds the page, registers its click handler, and opens the window
(`examples/css-preview/run.go`). Each of the 43 numbered property cards has its
own Start button. The Start click publishes the card's class change with one
`SetData` and `Redraw`; the automatic driver samples the animations, timelines,
and paint transitions. The few cards the runtime does not sample yet keep an
isolated preview player (`examples/css-preview/manual_motion.go`), and the page
stops ticking when no sample is active (`examples/css-preview/demo.go`).

Cross-repository validation uses the same local Blinkless source: run
`make catalog-ownframe-check` from the Blinkless root. That target creates a
temporary workspace for both checkouts and checks all 43 property cases.
