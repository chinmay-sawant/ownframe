# CSS animations and transitions

Blinkless reads CSS animation and transition declarations and writes their
values and `@keyframes` to the drawing list. OwnFrame's CSS runtime packages
bind that data to retained paint operations. They do not start by themselves
when a Page opens.

## Animations

`cssanimation.Bind` connects a box's animation records and matching keyframes
to selected paint operations. It takes the display, box, operation indexes,
starting paint values, and start time (`internal/cssanimation/bind.go:12-43`).
`animation.Player.Step` samples the animation with a caller-supplied time.
`StepWithInput` also accepts scroll progress, view progress, named timelines,
and range values (`internal/animation/player.go:22-26`,
`internal/animation/timing.go:62-70`).

An animation with a CSS trigger waits for an event. The caller sends that event
through `csstrigger.Dispatch`, which applies matching actions to animation
players (`internal/csstrigger/dispatch.go:11-23`).

## Transitions

`csstransition.Bind` selects a matching transition record and connects the
starting and current paint values to a renderer
(`internal/csstransition/bind.go:12-52`). The caller steps that renderer with
time (`internal/transition/step.go:10-44`). `allow-discrete` currently changes
the `position` paint value at the halfway point. Visibility and display need
operations to appear or disappear, which this renderer does not handle
(`internal/csstransition/discrete.go:3-7`).

## View transitions

`viewtransition.New` pairs elements with the same view-transition name in old
and new displays, inside the selected roots
(`internal/viewtransition/new.go:21-66`). `At` samples a progress value and
returns the old and new paint layers; `Frame.Display` combines those layers
with the new display (`internal/viewtransition/frame.go:23-43`,
`internal/viewtransition/display.go:5-31`). The caller provides both displays
and the progress value. OwnFrame does not capture them automatically from a
Page update. View-transition class styles are passed as Go values for opacity
and scale (`internal/viewtransition/viewtransition.go:27-28`); CSS
pseudo-element rules and `document.startViewTransition` are not wired into the
Page window.

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
(`examples/css-preview/run.go:13-32`). Each of the 43 numbered property cards
has its own Start button. The preview binds the selected animation or
transition, advances it on the Page tick, and stops ticking when no sample is
active (`examples/css-preview/player.go:31-60`).

These runtime packages live under `internal`, so code outside the OwnFrame
module cannot import them directly. Cross-repository validation uses the same
local Blinkless source: run `make catalog-ownframe-check` from the Blinkless
root. That target creates a temporary workspace for both checkouts and checks
all 43 property cases.
