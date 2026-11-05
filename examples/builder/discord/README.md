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
