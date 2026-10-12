# Forms

The template may contain `input`, `textarea`, and `select`. Each control needs an id. A control with no id is ignored. The library stores the value. `examples/forms` reads the stored values when the send control is clicked.

```go
page.Handle(gpui.Handlers{
    Click: func(_ context.Context, box gpui.Box) error {
        if box.Action != "send" {
            return nil
        }
        email := page.FormValue("email")
        page.SetData(struct{ Status string }{Status: email})
        return nil
    },
})
```

`Click` calls that handler, then draws the page again. A click on a control updates it before the handler runs. A nil handler ignores the event, and an error from the handler returns before the redraw. A later `Handle` call replaces every handler from the earlier call. `FormValue` returns the stored string. `FormChecked` reports a checkbox or a radio. `FormSelected` reports whether a control is focused and its whole value is selected. `FocusedField` is the id of the focused control, or empty when none is focused.

Supported `input` types:

- `text`, and an `input` with no type
- `password`, `email`, `search`, `tel`, `url`, and `number`
- `file`
- `checkbox` and `radio`

Any other `input` type is left alone.

Those text-like types, including `file`, take typed text. Click one to focus it. Click a textarea to focus it too, and typing edits that text. A click that lands on a non-control element blurs the focused control, and Tab does not move focus. `Type`, `Backspace`, `DeleteWord`, `Paste`, `SelectAll`, `Copy`, and `Cut` edit the focused text-like control, or the focused textarea, even when those handlers are nil. The method contracts are in [editing.md](editing.md). `SelectAll` selects the whole value, so `FormSelected` is true until the next edit, click, or blur. A password stores plaintext. The picture shows one bullet for each rune. `Copy` returns the plaintext.

Click a checkbox to toggle it. Click a radio to check it and uncheck the other radios with the same name. Click a select to cycle to the next option; a `multiple` select cycles one option at a time and paints one label. Those clicks focus the control. A disabled control does not toggle, focus, or take typing. The click handler still runs. Enter in the window calls `Handlers.Submit` and does not type a newline, and a click on a submit button is an ordinary click.

A file input opens the desktop dialog when `Run` installed a picker. On Linux `Pick` runs the first of `zenity`, `qarma`, `matedialog`, and `kdialog` found on `PATH`. Under WSL it first asks the Windows dialog through `powershell.exe` or `pwsh.exe` and converts the result with `wslpath`, or directly for `\\wsl.localhost` and `\\wsl$` paths, before it tries the Linux programs. That lookup uses `PATH` and the standard `/mnt/<drive>/Windows/System32/WindowsPowerShell/v1.0/powershell.exe` and `/mnt/<drive>/Program Files/PowerShell/*/pwsh.exe` installs, so `appendWindowsPath = false` in `/etc/wsl.conf` still finds the shell. Set `GPUI_FILEPICK_DEBUG=1` to print to stderr why a picker fell back. On Windows it calls `comdlg32!GetOpenFileNameW`. On macOS it runs `osascript`. The chosen path is stored as the value, `BeforeEdit` runs, and `Change` fires when the path differs. A cancel, a missing program, wasm, mobile, and `Serve` leave the old behavior: the field focuses, and the user types the file name. An empty file field paints the words "No file".

Before paint, a text-like input or a textarea is rewritten to a `span`. The engine paints an input value on one truncated line; the rewrite wraps a long value and adds the caret span and the selection background. Checkbox and radio stay `input` elements, and the engine paints the mark. A select stays a `select` and paints only the selected label. The rewrite keeps the author's attributes, such as `class`, `style`, `placeholder`, `name`, `autocomplete`, `maxlength`, `readonly`, `required`, `aria-*`, and `data-*`. The library enforces neither `maxlength` nor `readonly`; typing edits the control anyway. The span drops `value` and `type`, because the stored text carries the state, and the rewrite drops any `data-gpui-*` that the library manages. Library state appears as attributes: `data-gpui-field="input"`, `data-gpui-field="textarea"`, or `data-gpui-field="select"`, plus `data-gpui-focus="1"`, `data-gpui-selected="1"`, and `data-gpui-placeholder="1"` when they apply. A focused text field or textarea that is not selected carries an empty caret span, `<span data-gpui-caret="1"></span>`. A password paints one bullet for each rune. When the value is empty and the author set `placeholder`, the field paints the placeholder text. A file input still paints "No file". The default look is a `<style>` block. If an opening `<head ...>` tag is found, it is inserted immediately after it and before author CSS. Otherwise, it goes immediately before the first `</head>` tag if one exists; if neither tag is found, it is prepended to the document. The base rule sets `white-space: pre-wrap` on the rewritten span, so a long value wraps inside the box and grows its height. An author base rule of equal specificity wins because it comes later. A state rule adds a second attribute selector, so it beats an author rule of the same shape, but an id rule still wins. A `<button>` is not a form control. The layout engine's default stylesheet gives it a face, so a bare button lays out with a hit box. An `input` with `type=submit`, `type=button`, or `type=reset` is rewritten to a `button` element that keeps the other attributes; its `value` becomes the label, and an empty submit or reset value becomes `Submit` or `Reset`.

A control may carry `data-bind="Field"` to tie it to a field on the struct passed to `SetData`, which must be a pointer. An edit writes the new value into the field before the page is drawn, and `SetData` followed by `Redraw` pushes the field's value back into the control. `Handlers.BeforeEdit` runs before the built-in edit and before the field is written, so an app can snapshot for undo; an error aborts the edit. `Handlers.Change` runs after a control changes, bound or not, and receives that control's box; a nil function ignores the event. The full contract, including the field types, is in [binding.md](binding.md). A page may also use `:focus`, `:focus-visible`, `:hover`, `:active`, and `:checked`. The window passes the ids of the focused, hovered, and pressed elements, and the engine matches `:focus-visible` exactly like `:focus`. `:checked` follows the checkbox or radio `checked` attribute. The `data-gpui-*` attributes remain for styling without state CSS.

`Load`, `Back`, and `Forward` reset form state, because the document changed. `SetFormValue` and `SetFormChecked` do not redraw. Call `Redraw` after them. `SetFormValue` does nothing before that control has been drawn. For a select it marks each option whose value matches, and it clears the selection when that control is the focused one. `SetFormChecked` changes a checkbox or a radio and does nothing for any other control or a missing id; setting a radio true clears the other radios with the same name, and setting a radio false clears only that radio.

There is no native file dialog on wasm, on mobile, or on the `-web` page. Those keep the typed name.
