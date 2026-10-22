# Desktop cat

```sh
go run ./examples/desktop-cat
```

The orange backpack cat sits at the bottom right of the current monitor.
It gently bobs and changes expression every eight seconds. The collection
contains 30 transparent PNGs of the same cat across 15 emotions, including
happy, sad, angry, sleepy, worried, and bittersweet.

The borderless window stays above normal windows. Empty areas pass mouse input to
the application underneath. Click the cat to reopen the latest notification;
click the HTML/CSS speech bubble to dismiss it. Drag the cat to move the
window across the screen. Only its visible pixels capture clicks and drags;
transparent pixels around the image stay click-through. The animation keeps
running while you work in another application. Press Ctrl+C in the terminal
that launched it to close it.

```sh
go run ./examples/desktop-cat -variant 1   # approved happy soft smile
go run ./examples/desktop-cat -variant 19  # sad pout
go run ./examples/desktop-cat -variant 9   # angry frown
go run ./examples/desktop-cat -margin 80
go run ./examples/desktop-cat -random-behavior
```

`-variant 0` cycles through every expression. Other numbers select one image.
`-margin 48` sets the inset from the monitor edges in CSS pixels. Increase it
to clear a taskbar or dock. Placement uses monitor bounds, including docks
and taskbars. The window manager controls final placement and stacking.

## Images

The assets live in [assets/cat_images](assets/cat_images). Image 01 is the
approved identity reference. The other images preserve its orange fur,
folded ears, face, and black backpack while changing expression and posture.
The built-in GPT image tool generated the PNGs. The prompt set is
[prompts.json](assets/cat_images/prompts.json). Open the
[expression gallery](assets/cat_images/index.html) to compare them.

The example embeds the PNGs, so it works from any working directory. Each
source is displayed in a 200 by 200 CSS-pixel box inside a 320 by 350 window.
The tick changes the retained image's position. An expression switch swaps
its registered PNG and relayouts the cached HTML without reparsing it.

## Desktop platforms

The [platform check](platforms.md) distinguishes API support, compile checks,
and live desktop evidence. Windows, macOS, and regular Linux use the same
`RunWithOptions` entry point. Linux requires a compositor for transparency.

Under WSL, the command builds and launches a native Windows executable.
WSLg did not preserve transparency or click-through in our desktop check.
The launcher requires Windows interop and this source checkout. It removes
its temporary executable when the cat exits.

## Browser previews

```sh
go run ./examples/desktop-cat -web -variant 19
```

`-web` serves a picture preview at `http://127.0.0.1:8128`. `Serve` does not
call frame ticks, so this mode does not cycle expressions or create a desktop
overlay. The preview applies agent notifications before serving each frame.

```sh
./scripts/browser.sh desktop-cat
```

The WASM build uses the repository's generic browser loader. The cat example
contains no JavaScript or JavaScript interop. The custom iframe wrapper,
browser notification form, and browser drag controls were removed. Desktop
click-through, dragging, and native Chrome media monitoring require the
native example. Set `OWNFRAME_BROWSER_PORT` when port 8092 is already in use.

## Agent notifications

The native command also starts the agent API on a Unix socket at
`$XDG_RUNTIME_DIR/ownframe-desktop-cat-<uid>.sock` (mode `0600`), so only
your user can reach it. Socket traffic stays plaintext HTTP with an
`Origin` check; the socket permissions are the auth. `-notify-addr` with
a `host:port` opts back into TCP (keep loopback: a LAN bind would expose
notification content); `-notify-addr off` disables the listener. Under
WSL the server stays in Linux and forwards
notifications to the Windows cat through its stdin pipe. An agent running
in WSL can use the same socket path. `-notify-addr` changes the address.

```sh
curl --fail-with-body --max-time 2 --unix-socket "$XDG_RUNTIME_DIR/ownframe-desktop-cat-$(id -u).sock" http://localhost/notify \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://localhost' \
  --data-binary '{"expression":"happy","message":"The work is done. Response is ready.","source":"Open Code"}'
```

TCP opt-in example (explicit `-notify-addr=127.0.0.1:6969`):

```sh
curl --fail-with-body --max-time 2 http://127.0.0.1:6969/notify \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:6969' \
  --data-binary '{"expression":"happy","message":"The work is done. Response is ready.","source":"Open Code"}'
```

`POST /notify` and `GET /state` require a matching `Origin` (scheme
and host); a missing `Origin` returns 403. Curl callers send the
`Origin` header explicitly. Responses set `nosniff`, `DENY` framing,
`Referrer-Policy: no-referrer`, and a restrictive `Permissions-Policy`.

| Route | Result |
|---|---|
| `POST /notify` | Accepts `expression`, `message`, and optional `source`; returns HTTP 202. Requires matching `Origin`; missing `Origin` is 403. |
| `GET /expressions` | Maps 15 emotion names and 30 individual pose names to image numbers. |
| `GET /state` | Returns the last accepted notification and its sequence number. Requires matching `Origin`; missing `Origin` is 403. |

Use `happy`, `sad`, `angry`, `sleepy`, `worried`, or another name returned by
`/expressions`. Individual poses include `happy-soft-smile`, `sad-tear`,
and `sleepy-yawn`. A notification selects its mapped PNG, stops automatic
expression cycling, and opens the bubble. Later notifications replace the
message and expression. Dismissing the bubble preserves the latest message and switches to a happy
face. Clicking the cat restores the notification expression.
The bubble includes the source and local receipt time. Messages accept up
to 160 characters and sources up to 32. Keep messages short for the bubble.
Unknown expressions, malformed JSON, and empty messages return HTTP 400.

For example, send a `sleepy` notification with "You should sleep now." or a
`curious` notification with "Have you had enough water?" These are incoming
messages; the example does not schedule reminders itself.

Notifications leave the cursor and active application alone. The example
reads the cursor to choose click-through regions; it does not move it.

A WASM build cannot listen as a local server. Run the native command
for the socket API.

## Response-ready skill

The repository skill is [catnotify](../../.agents/skills/catnotify/SKILL.md).
It sends one curl request when the agent's response is ready, without retries,
automatic application launches, or cursor movement. It is also linked into
this machine's global agent skill directories.

OpenCode discovers `.agents/skills/<name>/SKILL.md`; see its
[skill documentation](https://opencode.ai/docs/skills/). Ask the agent to use
`catnotify`, or invoke `$catnotify` in Codex. A skill supplies agent instructions;
it is not a mandatory completion hook. Restart an agent session if its skill
catalog was loaded before installation.

## Random behavior

`-random-behavior` chooses a different random idle expression every eight
seconds. A notification holds its requested expression until dismissed.
Dismissal chooses one of the two happy faces, then idle expressions resume.
Without the option, dismissal holds the happy soft smile. Set the option when launching the native example.

Dragging starts after a four-pixel mouse movement on the cat. A press and
release without dragging reopens the latest message. The desktop moves only
in response to that gesture. Native drag movement was checked with automated
state tests; a live desktop drag test was skipped to leave the user's mouse
alone. The custom browser drag implementation was removed with the JavaScript
wrapper.

## Native Chrome playback

The Windows example watches Chrome's system media session from Go through
WinRT. It also works through the WSL native launcher. A playing track that
settles over two consecutive samples produces "Now playing: <title>" with
the happy cat. Sampling runs every two seconds. Pausing and resuming the
same track does not repeat its notification. No Chrome extension or
JavaScript is required for media detection.

`-media=false` disables this behavior. `-media-snapshot` reads the currently
playing Chrome title and exits without opening a window or changing playback.
Under WSL, the native media monitor posts through the local HTTP server and
the launcher forwards the notification to the window. `/state` therefore
includes the latest media notification.

Windows reports the source as Chrome and supplies the title and artist.
It does not supply the web URL, so other media playing in Chrome can also
produce a notification. Chrome must publish a system media session for the
track to be visible. The monitor never plays, pauses, changes a track, moves
the mouse, or focuses Chrome. The previous JavaScript extension was removed.
Native media monitoring on regular Linux, macOS, and WASM is not implemented;
those platforms still accept agent notifications through their supported UI.

See [Microsoft's media session API](https://learn.microsoft.com/en-us/uwp/api/windows.media.control.globalsystemmediatransportcontrolssession)
and its [playback states](https://learn.microsoft.com/en-us/uwp/api/windows.media.control.globalsystemmediatransportcontrolssessionplaybackstatus).
