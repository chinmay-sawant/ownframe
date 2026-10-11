# Run the accessibility and desktop input demo

The demo shows the page's semantic tree, sends accessibility actions, and lets
you try text editing and desktop IME composition.

## Start the demo

From the repository root, run:

```sh
go run ./examples/accessibility
```

## Inspect the accessibility tree

1. Click **Refresh tree**.
2. Find the named textbox, textarea, checkbox, radio buttons, select, list, and
   activation button in the snapshot.
3. Click **Focus display name**. Refresh the tree and confirm that the textbox
   reports focus.
4. Click **Set display name** and **Toggle checkbox**. Confirm that the live
   values change in the snapshot.
5. Click each error probe. The status line should name the error returned by
   the public API.
6. Confirm that the password node has no value in the snapshot.

The example tests this coverage with:

```sh
go test -p 1 ./examples/accessibility/...
```

Run that command from the repository root.

## Check text editing and IME

1. Click inside **Display name** between two characters. Type a character and
   confirm that it appears at the caret.
2. Use a desktop input method that supports composition. Type a preedit string
   and confirm that the page shows the temporary underlined text.
3. Move the caret within the field. Confirm that the input method's candidate
   window follows the caret.
4. Commit the composition. Confirm that the field value changes once.
5. Start another composition, move focus to a different field, and confirm
   that stale composition text does not appear in the newly focused field.
6. Start a composition and press Escape or close the window. Confirm that the
   pending text is cancelled.

Candidate-window placement depends on the desktop input method. Run this check
on each operating system you plan to support. The Go tests cover caret bounds
and session cleanup, but they do not replace that OS check.

## Check scrolling and the scrollbar

1. Make the window short enough to show the **Scrolling fixture** below the
   main panels.
2. Drag the scrollbar thumb. Confirm that the page moves and the pointer shows
   the resize cursor while it is over the thumb.
3. Click a control in the page. Confirm that the click activates the control
   without starting a scrollbar drag.

The focused tests for accessibility, IME, and pointer routing are:

```sh
go test -p 1 ./internal/page ./internal/window -run 'Test(Accessibility|IME|EscapeCancelsIME|ContentClickAndScrollbarDragKeepTheirEvents|HeldPressDoesNotStartThumbDrag)' -count=1
```

Run the full repository suite with `make test`.

## Know what this demo does not test

The tree preview checks the page API. It does not connect the window to a
screen reader. Native AT-SPI, UI Automation, and NSAccessibility adapters are
still open work. CSS-hidden elements may also remain in the tree because
Blinkless does not expose computed visibility. See the
[native accessibility status and validation guide](../../documentation/accessibility-native.md).
