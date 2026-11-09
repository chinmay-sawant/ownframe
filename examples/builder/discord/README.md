# Discord clone

A Discord-style chat shell built on `ui` + `chadcn`: server rail, channel
sidebar, message thread with reactions and composer, member list, voice bar,
search, settings, and toasts.

## Run

```sh
go run ./examples/builder/discord
go run ./examples/builder/discord -web
go run ./examples/builder/discord -web -addr 127.0.0.1:8152
```

`-web` serves a browser preview. `-headless` is the same as `-web`.
`-dev` opens the devtools dock. `-perf` enables devtools sampling.
`-debug-addr` serves window profiling on a loopback address.

## Transcript windowing

Landed: the transcript owns its scroll offset on the host (`chatScroll`
`chadcn.ScrollState` in `app.go`). The viewport (`dc-scroll-viewport`)
clips with hidden overflow and renders only the visible slice
(`visibleRange` in `scroll.go` picks message indexes from the offset, with
300 px overscan on both sides), so per-refresh cost no longer grows with
message count. The wheel handler (`scroll.go`) consumes wheel events over
the viewport (48 px per notch) and `PageUp`/`PageDown`/`Home`/`End` move
the window from the keyboard, so the 100vh shell chrome stays put.
`refresh` remeasures the viewport box and resets the scroll range against
the estimated total (`scrollTotal`); sending a message sets `scrollStale`
to follow to the bottom. Row heights are width-calibrated estimates, so
the scrollbar thumb stays truthful. Typing indicator, reactions, and
badges update via `PatchElement` against the visible slice, falling back
to a full refresh when a patch is refused.

## Memory bounds

Landed: history is capped at 500 messages per channel with drop-oldest on
append (in-place `copy`, no steady-state allocation). The height
estimator is pure arithmetic over `len()` and flags. The typing tick
allocates nothing on its steady path. Row state beyond the cap is dropped,
so long transcripts hold steady memory.

## Controls

- Server rail: switch servers, unread dots, mention badges.
- Channel list: collapse categories, switch text/voice channels.
- Messages: hover actions, react, reply, pin, edit own messages.
- Composer: per-channel drafts, typing indicator, send on submit.
- Voice bar: join a voice channel, mute, deafen.
- Top bar: channel topic, search, member list toggle, settings.
- Search overlay: filter messages across channels.
- Settings: profile, appearance, notifications tabs.
- User bar: profile popover, create-server dialog, toasts.
