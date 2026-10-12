# Binding

`data-bind="Field"` ties a form control to a field on the struct passed to `SetData`. Values move both ways. A user edit updates the field, and the next `Redraw` paints the field's value into the control. Without `data-bind`, the library still stores form values by control id, and `FormValue` and `FormChecked` read them. That is [forms.md](forms.md).

## Example

The template:

```html
<input id="email" type="text" data-bind="Email">
<textarea id="note" data-bind="Note"></textarea>
<select id="plan" data-bind="Plan">
  <option value="free">Free</option>
  <option value="pro">Pro</option>
</select>
<label><input id="remember" type="checkbox" data-bind="Remember"> Remember</label>
<input id="card" type="radio" name="pay" value="card" data-bind="Pay">
<input id="cash" type="radio" name="pay" value="cash" data-bind="Pay">
```

The Go:

```go
type Profile struct {
    Email    string
    Note     string
    Plan     string
    Remember bool
    Pay      string
}

view := &Profile{Email: "you@example.com", Plan: "free", Pay: "card"}

page, err := gpui.New(gpui.Config{HTML: profileHTML, Width: 420, Height: 360})
if err != nil {
    return err
}

page.SetData(view)
page.Handle(gpui.Handlers{
    Change: func(_ context.Context, box gpui.Box) error {
        log.Printf("data-bind %q changed, email is now %q", box.ID, view.Email)
        return nil
    },
})
```

For `data-bind`, `SetData` must receive a pointer to the struct. That pointer is how the library writes an edit back into the field.

## Controls and types

`data-bind` works on the form controls the library already tracks: `input`, `textarea`, and `select`. A control still needs an id. An element that is not a form control is ignored, and so is an `input` type the library does not handle.

The field type must match the control:

| Control | Field type |
|---------|------------|
| text-like input, textarea | `string` |
| select | `string` |
| checkbox | `bool` |
| radio | `string` |

A text-like input or textarea binds its text; a `file` input is text-like, and its picker stores the chosen path in the field. A select binds the value of the option the user moves to; `Redraw` paints the option whose `value` matches the field, even when the HTML marks another option selected. A field value that matches no option paints the first option. A checkbox binds whether it is checked. A radio binds its `value` when it is the checked one; `Redraw` checks the radio whose `value` matches the string, and choosing one radio unchecks the other radios that share its non-empty `name`.

## Directions

Edits write through. When a user changes a bound control, the library stores the new value and writes it into the struct field before the page is drawn again. Typing, backspace, delete-word, paste, and cut are edits of a text-like control or textarea. A click is an edit when it toggles a checkbox, checks a radio, or moves a select to another option. A `file` click is an edit when the picker returns a path different from the stored one. `Change` and the template both run after the write, so both see the new value.

Redraw pushes. `SetData` stores the data. The next `Redraw` reads every bound field and puts that value into its control, so the control paints what the struct holds. `SetData` alone does not touch the controls. Call `Redraw` after it.

## BeforeEdit

`Handlers.BeforeEdit` fires before a user edit changes a control, before the built-in edit and before the field is written. It receives the same `Box` as `Change`. Typing, backspace, delete-word, paste, and cut fire it; so do checkbox toggles, radio checks, select moves that change something, and a file pick that returns a new path.

- A nil `BeforeEdit` ignores the event.
- An error from `BeforeEdit` aborts the edit: the value stays, no field is written, no `Change` fires, and the page does not redraw.
- `Cut` fires it before clearing, so an app can snapshot the old value for undo. It fires even when the field is already empty.

## Change

`Handlers.Change` fires after a user edit changes a form control and before `Redraw`. When the binding resolves, the library has already written the new value into the struct field. It receives the last `Box` with that control's id, or `Box{ID: id}` when the page has no such box, so two controls bound to the same field report different boxes.

- A nil `Change` ignores the event. The redraw still happens, and a bound field still receives the write.
- An error from `Change` stops the redraw and is returned to the caller. The struct field keeps the new value.
- An edit that changes nothing does not fire `Change`; backspace in an empty field is one example. `Cut` is the exception: it fires even when the field is already empty.

## Limits

- For `data-bind`, the data must be a non-nil pointer to a struct. A value struct is inert for binding: `data-bind` is ignored in both directions, though the template still prints it. Use `&view`.
- Only exported fields bind. A name that is missing, unexported, or whose type does not match the control is ignored: nothing is written to the struct, and `Redraw` pushes nothing into that control. The control still edits like any form control, and its edits still fire `Change`.
- Binding follows the most recent `SetData`. The next `Redraw` reads the struct that call stored.
- A disabled control cannot be edited, so it never writes through.

The worked examples are [examples/bind](../examples/bind) (text, checkbox, radio, and select controls with `Change`), [examples/bind-hooks](../examples/bind-hooks) (`BeforeEdit` and `Change`), and [examples/login](../examples/login) (`BeforeEdit` snapshots fields for undo).
