# IPC

Electron IPC carries messages between a Node main process and a Chromium page process. ownframe has neither. `Send`, `Listen`, `Handle`, and `Request` are calls inside this one process. The root file `ipc.go` forwards them to `internal/ipc`.

```go
cancel := ownframe.Listen("tick", func(payload string) {
    // payload is the string passed to Send
})
ownframe.Send("tick", "1")
cancel()

ownframe.Handle("save", func(ctx context.Context, payload string) (string, error) {
    return "ok", nil
})
reply, err := ownframe.Request(ctx, "save", "now")
```

`Send` calls every current listener for that channel. A listener that panics does not stop the others. `Listen` returns a cancel func for that one listener.

`Handle` stores one reply function and returns a cancel func. A later `Handle` on the same channel replaces it. Cancel removes it only while it is still the current handler.

`Request` calls that function and returns its string and error.

These calls fail closed:

| Call | Result |
|------|--------|
| `Send`, `Listen`, or `Handle` with an empty channel | No callback runs; nothing is registered |
| `Request` with an empty channel, or with no handler | `ErrNoHandler` |
| `Request` with a nil context | Error text contains `nil context`, and the handler is not called |
| `Request` with an already canceled context | `ctx.Err()`, and the handler is not called |
| `Request` whose handler panics | `ErrHandlerPanicked`; the panic is recovered |
The calls are safe to use from several goroutines. They do not open a socket, and the `-web` host does not expose them.

The example in [examples/ipc](../examples/ipc) shows both patterns: two listeners on one channel, one handler on another, the registered state of each, and the cancel that removes them.

## Trust boundary

IPC calls are in process only. They are not an isolation boundary. Any goroutine
in the same process can listen on, replace the handler for, or send to a
channel by name, and there is no authentication or namespacing. Channel names
are not secrets. `Serve` does not expose IPC.

Recommended protections:

- Use IPC only between trusted components in one process. Do not build process
  or network isolation on top of it.
- Handlers must return promptly. A slow `Handle` blocks `Request` callers and
  there is no timeout beyond the caller's context.
- Guard against panics. `Send` recovers listener panics, and `Request`
  recovers a handler panic into `ErrHandlerPanicked`.
- Avoid user controlled channel names. Registration has no caps, so unbounded
  registration or oversized names can grow memory.
- Validate payloads at handler entry, since they are untyped strings.

See the security audit in
[plans/v0.0.2/security](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md).
