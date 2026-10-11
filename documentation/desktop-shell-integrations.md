# Desktop shell integrations

This note records the current boundary for native menus, tray icons,
notifications, and multiple windows. It describes the design work needed
before ownframe can expose those integrations.

## Current support

`Run` and `RunWithOptions` create one `shell` and pass it to
`ebiten.RunGameWithOptions`. Ebitengine's public API has no second-window
handle. Its `RunGameWithOptions` documentation also says not to call the
function more than once in one process. Ownframe cannot safely open a second
page by calling `Run` again or starting another game loop in a goroutine.

`host.ContextMenu` is the only menu hook. The window paints its rows in the
game canvas and dispatches the supported edit actions on the game update
thread. It is a portable in-window fallback. It does not create a native menu
bar or change the operating system's application menu.

Ebitengine exposes no public tray or desktop notification API. Ownframe has no
native adapter for these features. No tray icon or notification delivery is
currently available through the library.

## Why the native work is still open

The existing runner owns process-global Ebitengine settings and a single game
loop. A multi-window API needs an application owner above individual pages.
That owner must route input to each window, retain independent page state, and
define what happens when one window or the last window closes. Adding a second
call to `Run` would violate Ebitengine's runtime contract.

Native menus, tray icons, and notifications require operating-system APIs
that Ebitengine does not expose. Their callbacks may arrive outside the game
update thread, while page handlers currently run from the window loop. A
native adapter must enqueue actions for the loop instead of calling a page
directly from an OS callback.

Tray behavior also depends on the desktop environment, and notification
delivery may be disabled or denied by the operating system. An adapter must
report unsupported platforms and delivery failures to the application. It
must release native resources when its owner closes.

## Follow-on design

Before implementation, choose the supported desktop systems and confirm
whether platform-specific dependencies are acceptable. Keep the existing
`host.Screen` page contract. Add an application owner that creates and tracks
windows, assigns each window an ID, and owns shutdown. A window close request
must identify its window; closing the final window must stop the owner.

Represent native menu rows as stable action IDs and labels, with enabled state
that can be refreshed from the application. Send menu selections back as
owner events. Keep `host.ContextMenu` for the in-window edit menu and as the
fallback when a native menu adapter is unavailable.

Give tray adapters an explicit start and close lifecycle. Route tray clicks
through the application owner. Return an error for unsupported platforms and
for failures the backend can detect. Document what each backend can report
about notification permission. Do not treat a successful request as proof
that the user saw it. If notifications support click actions, dispatch those
actions through the same owner event queue.

Build the owner and event routing with fakes before adding OS adapters. Test
window creation, independent page state, close ordering, final-window
shutdown, menu action dispatch, tray teardown, and notification failures
without opening a native window. Add one runnable example for each integration
after an OS adapter exists. Record manual checks for each supported OS because
fakes cannot verify native menus, tray presence, notification permissions, or
window-manager behavior.

Do not add a platform package or claim native support until that platform has
an adapter and a manual verification route. If the project keeps the current
single-window Ebitengine model, mark multi-window support as deferred. The
in-window context menu remains available as the menu fallback.

## Current validation

Run the existing portable context-menu tests from the repository root:

```sh
go test ./internal/window -run 'Test(Menu|Escape)'
```

This checks the in-window menu model and actions. It does not validate native
menus, tray icons, notifications, or multiple windows.
