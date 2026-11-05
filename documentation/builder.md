# Builder patterns

Package `ui` builds HTML with Go calls. ownframe owns the window, input, and click dispatch. Blinkless parses the HTML, cascades the CSS, and lays the page out. There is no JavaScript runtime and no reconciled component tree. The API has no compatibility guarantee yet.

## Build a tree

`Tag` returns an element and lowercases the tag name. `Div`, `Button`, and `Text` are shortcuts on top of `Tag`. `Text` is a span whose text is escaped. `Button` sets the label as that element's text. It is not `chadcn.Button`. `Children` appends nodes in order and skips nil children.

An unsupported tag fails at `Render` or `NewPage`, not at `Tag`. Void tags `input`, `img`, `br`, and `hr` cannot take text or children.

`Mount` installs the tree when the host will rebuild it. `config` is an `ownframe.Config` with empty `HTML` and `File`. Handle `err` before using `view`. The Clicks section says what `OnClick` does. The Pages that update section says when to call `Mount` instead of `NewPage`.

```go
view, err := ui.Mount(config, ui.Div().Children(
	ui.Text("Saved locally"),
	ui.Button("Save").OnClick(func(context.Context) error {
		return nil
	}),
))
```

## Typed CSS

Each property implemented in the [compatibility matrix](compatibility-matrix.md) has a generated method. The CSS name becomes Go spelling, so `justify-content` is `JustifyContent`. The method takes that property's value type. A string does not type-check. A keyword from another property does not type-check either.

Keyword constants belong to one property, for example `DisplayFlex`, `FlexDirectionColumn`, and `JustifyContentCenter`. `Inherit`, `Initial`, `Unset`, `Revert`, and `RevertLayer` are allowed on every property.

`Px` builds a length. `Percent` builds a percentage and is not a length, so a border width cannot take `Percent`. `Hex` builds a color. `Edges` takes one to four lengths in CSS shorthand order.

```go
ui.Div().Display(ui.DisplayFlex) // compiles
// ui.Div().Display("flex") does not compile
```

The compiler checks the property name and the value family. It does not check the full CSS grammar. Ranges, context, and whether the engine honors a value are runtime checks. A zero value, a nil value, a bad hex color, or a non-finite number fails at `Render` or `NewPage`.

## Raw values

`RawValue` fills one known property when no typed helper exists. You still call the generated method, so the compiler checks the property name. `RawValue` checks declaration boundaries only. Blinkless checks CSS meaning.

`CSSRaw` takes a property name and a value chosen at runtime. An unknown name fails at `Render` or `NewPage`. `SupportsProperty` reports whether the name is in the matrix. It does not prove the value works, and `CSSRaw` does not make the engine implement a missing property. Prefer a generated method when the name is known at compile time. Declare variables with `CustomProperty`, not `CSSRaw`.

## Custom properties

`CustomProperty` declares a variable on a node. `Var` references one and takes at most one fallback. Names are case-sensitive and must match `--[A-Za-z0-9_-]+`. A bad name, an empty value, or an extra declaration fails at `Render` or `NewPage`. Inheritance, fallback, and cycles follow Blinkless. A variable cannot add a layout property the engine does not have.

`ColorVar` and `LengthVar` stay typed and require a fallback of that family. `Color(Var("--accent"))` is the untyped reference.

## Clicks

`OnClick` registers a Go callback. `NewPage` and `Mount` connect it through the action dispatcher. `Render` returns a fragment and does not connect clicks, so the callback is dropped.

Do not combine `OnClick` with a hand-written `data-action` on the same node.

`View.Handle` keeps builder callbacks and forwards other actions to your handler. `Page.Handle` replaces that dispatch. A non-nil click callback passed to `Page.AddHandlers` does the same.

## Pages that update

`NewPage` builds the document, calls `ownframe.New`, and installs click dispatch. Use it when the tree will not be replaced. It returns a page and no view, so a later rebuild is `Page.Load`, not `View.Update`.

`Mount` is the same construction and keeps the callback map. The sample under Build a tree is that call. `view.Page` is the page you pass to `Run` or `Serve`. After host state changes, call `view.Update` on the page loop with the new tree. Do not call it from an after-redraw callback.

`Config.HTML` and `Config.File` must be empty. `NewPage` and `Mount` write the document, including the doctype and the escaped title.

`Update` replaces the current document and the callback map. It does not push history. Controls with a stable ID and type keep their live value, caret, selection, and focus. Removed controls lose that state. Rebuilding an input with a new `value` attribute does not overwrite a live edit. Call `view.Page.Form().SetFormValue` before the update when you need to reset it.

## Theme

For ChadCN controls, set `Config.Theme` to `chadcn.Dark` or `chadcn.Light`, then append `chadcn.Stylesheet`. The palette sets the `--of-*` tokens. The stylesheet sets paint and state. If the page already supplies those tokens, append `chadcn.Stylesheet` alone.

`theme.Dark` and `theme.Light` in package `ui/theme` are not `chadcn.Dark` and `chadcn.Light`. Both families set `--of-*` names. They are separate sheets.

`theme.Overrides` writes custom properties on one element, or a `:root` block for the whole page.

`Page.SetTheme` stores another sheet. `Page.Redraw` applies it. `SetTheme` does not draw by itself.

## Limits

- There is no JavaScript runtime.
- There is no component lifecycle, no state hooks, and no partial subtree update.
- Trees are mutable. Do not edit one tree from two goroutines.
- Typed checks cover value families, not the full CSS grammar.
- Text and attributes are escaped, including `{` so a Go template action cannot run.
