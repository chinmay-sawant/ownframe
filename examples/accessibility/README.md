# Accessibility and desktop input

Run the standalone desktop demo:

```sh
go run ./examples/accessibility
```

The semantic fixtures cover every exported accessibility role, node field,
and action. Use the action buttons to focus or update a textbox, toggle a
checkbox, select a radio option, change a select, activate a button, and
exercise each public error result.
The tree preview redacts password values. The form also lets you try ordinary
text input and desktop IME composition; the long page lets you test scrolling
and the scrollbar thumb cursor. Example tests check this coverage with
`go test ./examples/accessibility/...`.

The page tree is available to application code. Native screen-reader adapters
are still planned, so the preview is a way to inspect current semantics rather
than a Windows, Linux, or macOS assistive-technology bridge.
