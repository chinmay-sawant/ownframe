# Loglens · production log dashboard

Loglens is an interactive observability dashboard built with the `ui` builder
and the ChadCN component set in `components/chadcn`. It is a demonstration, not
a product: every screen is one HTML tree rendered by OwnFrame, every control is
a ChadCN constructor, and the shared data comes from the fixtures in `data.go`.

The example exercises the whole component set in a realistic shape. An overview
tab shows SLO burn, KPIs, a request-volume chart, service health, deployments,
on-call coverage, and edge regions. A logs tab filters, pages, selects, and
inspects entries. A traces tab shows a waterfall, a payload, an RTL card, and a
recent-trace list with a scrollbar. An incident room combines chat, decisions,
a playbook, snapshots, a timeline, and a postmortem draft. The top bar carries
product navigation, menus, and popovers; the screen also mounts a settings sheet,
a command palette, a dialog, an alert dialog, and a toast. Fifteen surfaces use
ChadCN's `Open` parameter, while the command palette owns a query and an active
row.

## Run

From the repository root:

```sh
go run ./examples/builder/loglens
go run ./examples/builder/loglens -web
go run ./examples/builder/loglens -dev
```

| Flag | Default | Effect |
|---|---|---|
| `-web` | `false` | Serve the dashboard through `ownframe.Serve` instead of opening a desktop window. |
| `-addr` | `127.0.0.1:8132` | Listen address for `-web`. |
| `-dev` | `false` | Open the developer tools dock at startup through `Page.Window().SetDevTools`. |

The window opens at 1440 × 960 with a minimum of 1100 × 720. It starts in the
dark palette with a hidden scrollbar.

## Architecture

`main.go` parses the flags and calls `run`. `run` creates the app, applies the
devtools flag, draws the first frame, then hands the page to `ownframe.Run` or
`ownframe.Serve`.

`app.go` owns every mutable value and the event handlers. `newApp` mounts one
`ui.View`, keeps `view.Page`, and installs `ownframe.Handlers` for clicks,
hover, and key-down. `refresh` rebuilds the whole tree through
`View.Update(ctx, a.screen())`; every panel returns a tree built from the fields
in `app`, and no control keeps hidden state.

`shell.go` composes the screen: the top bar, the body with sidebar and main
column, then the settings sheet, global dialogs, and command palette as root
siblings. It also renders the breadcrumb, page title, action buttons, and the
tab set. Each tab's content is built by its panel function, but only the active
panel is emitted, so a tab switch replaces the tree.

`data.go` holds the fixtures and the log query. `sampleLogs` has 24 entries,
`sampleServices` 6, `sampleSpans` 5, and `sampleMessages` 5. `visibleLogs`
filters by service, severity, and a case-insensitive query over the message,
service, and trace ID. `pageLogs` slices 8 rows at a time, `logPageCount` sizes
the pagination, and `selectedLog` resolves the detail drawer's row. The option
lists for time range, service, severity, sort, and notification channel live
here too.

| File | Responsibility |
|---|---|
| `main.go` | Flags, context, first draw, `Run` or `Serve`. |
| `app.go` | App state, mount, handlers, theme switching, toast and overlay helpers, form access. |
| `shell.go` | Screen root, top bar, sidebar, page header, tab set, tab titles and subtitles. |
| `data.go` | Fixtures, option lists, log filtering, paging, and selection. |
| `overview.go` | Landing tab: SLO alert, KPI cards, traffic chart, service table, deployments, on-call calendar, regions, toolkit. |
| `logs.go` | Request stream: filter bar, error alert, toolbar, row menu, data table, detail drawer. |
| `filters.go` | Filter bar: time-range select, service combobox, date picker, query and search inputs, sort, severity, live tail, auto refresh, context, sample rate, apply and reset. |
| `traces.go` | Waterfall and payload split, recent-trace scroll area, RTL cards, topology, span index, owner row. |
| `incident.go` | Incident header, chat, decisions, action items, playbook, snapshots, customer note, impact map, timeline, postmortem. |
| `settings.go` | Project settings sheet: routing, noise control, verification. |
| `dialogs.go` | Create-alert dialog, delete-view alert dialog, page-level toast mount. |
| `menus.go` | Product navigation menu, application menubar, account dropdown menu. |
| `popovers.go` | Live tooltip, on-call hover card, notifications popover, command palette. |
| `styles.go` | `appCSS`: dashboard grid, spacing, badges, filter bar, overlays, chat and timeline layout. Component paint stays in ChadCN. |

## State model

Everything that changes lives on `app`. ChadCN constructors receive the current
values and return a fresh tree; no component keeps hidden state or a lifecycle.

- **Rebuild loop.** A control either carries an `OnClick` closure built in the
  panel or a shell-level `data-action`. A closure mutates one or more fields and
  calls `a.refresh(ctx)`. `handleClick` handles the shell actions with prefixes
  or exact names: `nav:<tab>`, `theme:toggle`, `settings:open`, `settings:close`,
  `toast:close`, `command:open`, `dialog:open`, and `alert:open`. Everything
  else is a closure.
- **Tabs.** `panelTabs` passes `Active: a.tab`; its `OnSelect` sets `a.tab` and
  refreshes. The sidebar uses `nav:<tab>` actions through
  `chadcn.SidebarNavItem`, so both routes land on the same field.
- **Click state.** `open` fields such as `dialogOpen`, `alertOpen`,
  `settingsOpen`, `drawerOpen`, `commandOpen`, `selectOpen`, `comboboxOpen`,
  `dateOpen`, `contextOpen`, `notifOpen`, `userMenuOpen`, `navOpen`, and
  `menubarOpen` are owned by the app. A component's `OnOpenChange` writes the
  field and refreshes. Escape calls `closeOverlays`, which clears all of them.
- **Hover state.** `handleHover` compares the next box ID with
  `tooltip-trigger` and `hover-trigger`, sets `tooltipOpen` or `hoverOpen`, and
  refreshes only when one of them changes. ChadCN does not own hover.
- **Form values.** Text inputs, textareas, and OTP slots are stored on
  `Page.Form`, not in the widget. `formValue` and `setFormValue` wrap
  `page.Form()`. The filter Apply button reads `log-query`, `log-search`, and
  `log-sort`, then writes the query back into app state and the form. The
  settings Save button reads `settings-email`, `settings-channel`,
  `settings-webhook`, `settings-notes`, and the six `settings-otp-*` slots. The
  dialog Save reads `alert-email`, `alert-notes`, and `alert-otp-*`.
- **Controlled non-text state.** Checkbox, radio group, switch, toggle, toggle
  group, select, combobox, tabs, accordion, pagination, carousel, and the date
  picker take their value back in from `app` on every build.
- **Theme.** `themeSheet` layers `appCSS` after `chadcn.Dark` or `chadcn.Light`
  and `chadcn.Stylesheet`. `toggleTheme` calls `Page.SetTheme` and then
  refreshes, so tokens change without a remount.
- **Panel-local state.** `extra` is a small `map[string]any` for values that do
  not deserve a field: the settings-plan override, the span-index flag, and the
  incident accordion seeding guard.

## Component map

All 66 ChadCN components appear. The canonical list is the package README at
`components/chadcn/README.md`. The file column names where the constructor is
called; a second location is shown when the same component is reused.

### Primitives and display

| Component | File | Where it appears |
|---|---|---|
| **Aspect Ratio** | overview.go | Edge-regions mesh; span topology in traces.go and impact map in incident.go. |
| **Avatar** | shell.go | Sidebar footer; KPI, deployment, trace, incident, and popover rows. |
| **Badge** | shell.go | Page title; status, severity, count, and level badges across every panel. |
| **Card** | overview.go | KPI cards and every content card; reused by logs, traces, and incident. |
| **Direction** | traces.go | Arabic live-direction card and forced-LTR card; customer note in incident.go. |
| **Empty** | overview.go | End of the rollout feed; no-match logs and the postmortem placeholder. |
| **Item** | overview.go | Deployment and notification rows; trace owner and incident action items. |
| **Kbd** | shell.go | Page header shortcut; shortcut lists and dialog footers. |
| **Label** | filters.go | Filter captions, threshold labels, and form labels in settings.go. |
| **Progress** | overview.go | KPI progress bars; latency budget in traces.go and mitigation in incident.go. |
| **Separator** | overview.go | Toolkit divider; reused in logs, traces, settings, dialogs, and popovers. |
| **Skeleton** | overview.go | Loading KPI and rollout placeholders; reused in logs, traces, and incident. |
| **Spinner** | overview.go | Loading error-rate badge; reused in logs, traces, and incident. |
| **Table** | overview.go | Service-health table; span attributes and the incident timeline. |
| **Typography** | overview.go | H1/H2/Lead/P/Muted in the toolkit; H3/H4 and InlineCode in incident.go. |

### Actions and forms

| Component | File | Where it appears |
|---|---|---|
| **Button** | shell.go | Page actions; the primary control in every panel. |
| **Button Group** | shell.go | Refresh/Export pair; traffic actions, log actions, trace selector, incident actions, filters, and settings presets. |
| **Toggle** | filters.go | Live tail; the overview loading preview and the logs toolbar. |
| **Toggle Group** | overview.go | Traffic range 24h/7d/30d; severity selection in filters.go. |
| **Field** | filters.go | Query field; alert fields and settings fields in dialogs.go and settings.go. |
| **Input** | filters.go | Query box; alert email, settings email, and the incident reply. |
| **Textarea** | settings.go | Handover notes; alert condition notes in dialogs.go. |
| **Native Select** | filters.go | Sort order; notification channel in settings.go. |
| **Checkbox** | filters.go | Show context; mute resolved incidents and page the on-call. |
| **Radio Group** | settings.go | Minimum severity; alert severity in dialogs.go. |
| **Switch** | filters.go | Auto refresh; alerting, RTL mode, span index, and alert toggles. |
| **Slider** | filters.go | Sample rate; settings threshold and the dialog error-rate threshold. |
| **Input Group** | filters.go | Log search prefix and row-count suffix; webhook prefix in settings.go. |
| **Input OTP** | settings.go | Authenticator code; alert verification code in dialogs.go. |

### Feedback and disclosure

| Component | File | Where it appears |
|---|---|---|
| **Alert** | overview.go | SLO budget warning; error-share alert above the log table. |
| **Toast** | dialogs.go | Mounted once in `globalDialogs`; every `showToast` writes its title and body. |
| **Accordion** | incident.go | Three-step response playbook with a multi-open set. |
| **Collapsible** | incident.go | Advanced notes below the playbook. |
| **Tabs** | shell.go | Overview, Logs, Traces, and Incident room. |
| **Breadcrumb** | shell.go | `Acme › Production › <tab>` above the page title. |
| **Pagination** | logs.go | Eight-row pages in the log toolbar. |

### Overlays and menus

| Component | File | Where it appears |
|---|---|---|
| **Dialog** | dialogs.go | Create-alert rule editor. |
| **Alert Dialog** | dialogs.go | Delete-saved-view confirmation. |
| **Sheet** | settings.go | Right-side project settings panel. |
| **Drawer** | logs.go | Log-detail drawer with payload, trace, and attachment. |
| **Popover** | popovers.go | Notification inbox under the top-bar Notifications button. |
| **Tooltip** | popovers.go | Live freshness hint on the Live button. |
| **Hover Card** | popovers.go | On-call card for Ada Kim. |
| **Dropdown Menu** | menus.go | Account menu with Profile, Project settings, and Sign out. |
| **Context Menu** | logs.go | Row actions anchored below the toolbar trigger. |
| **Menubar** | menus.go | File, View, and Help menus. |
| **Navigation Menu** | menus.go | Platform and Resources panels. |
| **Command** | popovers.go | ⌘K palette with Navigate, Actions, and Services groups. |

### Selection

| Component | File | Where it appears |
|---|---|---|
| **Select** | filters.go | Time-range picker. |
| **Combobox** | filters.go | Filterable service picker. |
| **Calendar** | overview.go | On-call month grid; also inside the filter date picker. |
| **Date Picker** | filters.go | "From date" trigger and calendar panel. |

### Host-owned composites

| Component | File | Where it appears |
|---|---|---|
| **Scroll Area** | traces.go | Recent-traces list with the shared `ScrollState`. |
| **Resizable** | traces.go | Waterfall and payload split at a fixed 0.55 ratio. |
| **Sidebar** | shell.go | Project navigation with workspace and saved-view rows. |
| **Carousel** | incident.go | Three live snapshots with previous/next controls. |
| **Data Table** | logs.go | The request stream with row selection. |
| **Chart** | overview.go | Request-volume bars for this period and the previous one. |
| **Code Block** | logs.go | Log payload JSON; trace payload JSON in traces.go. |
| **Message** | incident.go | Incident-bridge rows with avatar, name, and time. |
| **Bubble** | incident.go | Default and primary chat bubbles. |
| **Attachment** | incident.go | Heap-dump tile in the chat; trace export in the log drawer. |
| **Marker** | incident.go | Status-page marker and next-up timeline rows. |
| **Message Scroller** | incident.go | Fixed-height chat viewport with a top fade. |
| **Questionnaire** | incident.go | Two decision questions with progress and choices. |
| **Trace** | logs.go | Drawer trace timeline; the traces waterfall in traces.go. |

## Open-state map

These are the 15 `Open` state captures from `components/chadcn/README.md`.
Command is not one of them: it takes `Query` and `Active` instead of `Open`, and
is listed after the table because the palette is still part of the demo.

| Component | State field | Where it opens |
|---|---|---|
| **AlertDialog** | `alertOpen` | Mounted in `dialogs.go` with no visible trigger; the host opens it with the `alert:open` action for the delete-saved-view confirmation. |
| **Combobox** | `comboboxOpen` | Service picker in the filter bar. |
| **ContextMenu** | `contextOpen`, `contextX`, `contextY` | Row actions button in the log toolbar; opens below the trigger. |
| **DatePicker** | `dateOpen`, `calendarMonth`, `fromDate` | From-date picker in the filter bar. |
| **Dialog** | `dialogOpen` | Create-alert rule editor: page header button, File > New alert, command row, or a host `dialog:open` action. |
| **Drawer** | `drawerOpen`, `drawerIndex` | Log detail, opened from the row menu's Open detail item. |
| **DropdownMenu** | `userMenuOpen` | Account button in the top bar. |
| **HoverCard** | `hoverOpen` | Hover on the "On call: Ada" trigger (`hover-trigger`). |
| **Menubar** | `menubarOpen` | File, View, and Help in the top bar. |
| **NavigationMenu** | `navOpen` | Platform and Resources in the top bar; Settings opens the sheet instead. |
| **Popover** | `notifOpen` | Notifications button in the top bar. |
| **Select** | `selectOpen` | Time-range picker in the filter bar. |
| **Sheet** | `settingsOpen` | Project settings: sidebar gear, navigation menu Settings, account menu Project settings, command row, or `settings:open`. |
| **Toast** | `toastOpen`, `toastTitle`, `toastBody` | Mounted in `globalDialogs`; `showToast` opens it from Refresh, Export, alert saves, filter apply/reset, and most incident and trace actions. |
| **Tooltip** | `tooltipOpen` | Hover on the Live trigger (`tooltip-trigger`). |

The command palette is controlled by `commandOpen`, `commandQuery`, and
`commandActive`. The ⌘K button in the top bar opens it; Escape closes it along
with every other overlay. The ⌘K label is a hint, not a global keybinding.

## Limits

- **No backend.** `data.go` is the only source of logs, services, spans, and
  chat messages. The app code makes no network, `fetch`, or IPC call and keeps
  nothing on disk, so a restart restores the sample state.
- **Host-owned behavior.** Sorting, filtering, paging, row selection, carousel
  position, accordion state, and every open flag are application code. The
  components only render the values they receive.
- **No live feed.** Live tail and Auto refresh change their own switches, and
  Refresh shows a toast; no rows stream in and no timer polls.
- **No JavaScript or canvas.** The chart is plain div geometry with CSS bars.
  The region mesh and impact map are div and span geometry. There is no live DOM,
  no `<a href>` navigation, and no script tag.
- **Fixed split.** `Resizable` renders the divider, but the example never routes
  a drag to `splitRatio`, so the waterfall stays at 0.55.
- **Overlay focus is partial.** Escape closes the tracked overlays and each
  component closes on its backdrop callback. The example does not install a
  focus trap around the dialog, sheet, or drawer.
- **Scroll is page-level.** The recent-traces pane declares a `ScrollState` and
  the incident chat has a fixed-height viewport, but the example does not route
  wheel or drag input to either pane. The window scrolls the page.
- **Sample data is synthetic.** Services, traces, and incident channels mirror
  an observability product but are not connected to one.

## Verification

From the repository root:

```sh
gofmt -l examples/builder/loglens
go vet -p 1 ./examples/builder/loglens
go test -p 1 ./examples/builder/loglens
go run ./examples/builder/loglens -dev
go run ./examples/builder/loglens -web -addr 127.0.0.1:8132
make build
make test
make lint
```

`coverage_test.go` constructs the app with `newApp`, sets each tab, renders
`overviewPanel`, `logsPanel`, `tracesPanel`, `incidentPanel`, `logFilters`,
`topMenus`, `topPopovers`, `settingsSheet`, `globalDialogs`, `commandOverlay`,
and `screen` through `ui.Render`, then asserts that the combined markup contains
all 66 `.of-` component roots as whole class tokens. The `screen` render is what
contributes `.of-sidebar`, `.of-breadcrumb`, and `.of-tabs`; `chadcn.Screen`
itself has no `.of-screen` class. The test needs no window and no running app.

`make build` runs `go vet -p 1 ./... ./examples/...`, which compiles every
example without linking binaries. `make test` runs both Go modules. `make lint`
runs the pinned `golangci-lint` over both modules plus the documentation, plan,
and website checks.
