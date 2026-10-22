---
name: catnotify
description: Send a local desktop-cat notification when an agent response is ready, using the Unix socket API.
---

# Cat response notification

When your response is ready, send one short notification before delivering it.
Use the actual agent name as `source`, such as `Open Code` or `Codex`.
Use `happy` for a ready response. Describe the real result in at most 160
characters. The bubble appears without moving the user's mouse.
The cat listens on a `0600` Unix socket by default; TCP is explicit opt-in.

```sh
curl --silent --show-error --fail-with-body --max-time 2 \
  --unix-socket "$XDG_RUNTIME_DIR/ownframe-desktop-cat-$(id -u).sock" \
  http://localhost/notify \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://localhost' \
  --data-binary '{"expression":"happy","message":"Response is ready. The work is done.","source":"Open Code"}'
```

`POST /notify` and `GET /state` require a matching `Origin` header;
a missing `Origin` returns 403. Send the `Origin` header explicitly.

Serialize dynamic text as JSON before passing it to curl. Keep shell commands
literal so quotes or dollar signs in a message cannot execute as shell code.
HTTP 202 means the server accepted the notification. If the server is offline,
finish the response normally and mention that the notification was unavailable.
Send once, without retries or launching the cat automatically.

## Reaching a Windows cat from WSL2

The curl above runs on the machine the agent runs on and talks to the
Unix socket. A TCP listener is explicit opt-in (`-notify-addr=127.0.0.1:6969`).
Under WSL2 with NAT
networking that is the Linux VM, not Windows. A cat that runs as a Windows app
binds `127.0.0.1` to the Windows loopback, so Linux `localhost:6969` refuses
the connection while the cat serves fine on Windows. Do not stop at that
refusal. Run curl through WSL interop, which executes it on Windows:

```sh
/mnt/c/Windows/System32/curl.exe --silent --show-error --fail-with-body --max-time 2 \
  http://127.0.0.1:6969/notify \
  -H 'Content-Type: application/json' \
  -H 'Origin: http://127.0.0.1:6969' \
  --data-binary '{"expression":"happy","message":"Response is ready. The work is done.","source":"Open Code"}'
```

`GET http://127.0.0.1:6969/state` with `-H 'Origin: http://127.0.0.1:6969'`
through the same `curl.exe` tells you whether
the server is up. HTTP 202 still means accepted, and the send-once rule stands.

When interop is unavailable, either enable mirrored networking in `.wslconfig`
(`networkingMode=mirrored`, then `wsl --shutdown`), start the Windows cat with
`-notify-addr=0.0.0.0:6969` and allow it through the firewall, or run
`examples/desktop-cat` from WSL: it builds the Windows overlay, keeps the 6969
server in WSL, and forwards notifications over stdin.

For another emotion with the default socket, query
`GET http://localhost/expressions` through `--unix-socket` (or
`GET http://127.0.0.1:6969/expressions` on a TCP opt-in listener).
Use `worried` for a blocker, `sad` for a failed task, or `sleepy` for a sleep
reminder the user requested. Only report completion when the work is complete.

Deliver notifications through HTTP only. Leave the cursor position and the
user's active application alone.
