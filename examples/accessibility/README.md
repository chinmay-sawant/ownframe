# Accessibility and desktop input

Run the standalone desktop demo:

```sh
go run ./examples/accessibility
```

The labeled form lets you try ordinary text input and desktop IME composition.
Refresh the semantic tree to inspect roles, names, values, focus, and password
redaction. The second button changes the name by calling the accessibility
tree action API. The long page lets you test scrolling and the scrollbar thumb
cursor.

The page tree is available to application code. Native screen-reader adapters
are still planned, so the preview is a way to inspect current semantics rather
than a Windows, Linux, or macOS assistive-technology bridge.
