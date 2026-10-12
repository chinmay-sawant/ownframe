# IPC

Electron IPC carries messages between a Node main process and a Chromium page process. go-gpui has neither. `Send`, `Listen`, `Handle`, and `Request` are calls inside this one process. The root file `ipc.go` forwards them to `internal/ipc`.

```go
cancel := gpui.Listen("tick", func(payload string) {
    // payload is the string passed to Send
})
gpui.Send("tick", "1")
cancel()

gpui.Handle("save", func(ctx context.Context, payload string) (string, error) {
    return "ok", nil
})
reply, err := gpui.Request(ctx, "save", "now")
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

The calls are safe to use from several goroutines. They do not open a socket, and the `-web` host does not expose them.

The example in [examples/ipc](../examples/ipc) shows both patterns: two listeners on one channel, one handler on another, the registered state of each, and the cancel that removes them.
