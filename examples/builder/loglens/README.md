# Loglens · production log dashboard

Loglens is an interactive observability dashboard built with the `ui` builder
and the ChadCN component set in `components/chadcn`. It is a demonstration, not
a product: every screen is one HTML tree rendered by OwnFrame, every control is
a ChadCN constructor, and its explicitly labeled demo telemetry is generated
locally from seed fixtures. It does not connect to a production collector.

The example exercises the component set in a realistic shape. An overview
tab shows feed-derived KPIs, event counts by service, service health, recent
events, on-call coverage, and edge regions. A logs tab filters, pages ten
newest-first entries at a time, selects, and inspects entries. Refresh and
Auto refresh add generated logs and matching traces. A traces tab shows a
waterfall, a payload, an RTL card, and ten newest-first traces per page. An
incident room combines chat, decisions, a playbook, snapshots, a timeline,
and a postmortem draft. The top bar carries
product navigation, menus, and popovers; the screen also mounts a settings sheet,
a command palette, a dialog, an alert dialog, and a toast. Fifteen surfaces use
ChadCN's `Open` parameter, while the command palette owns a query and an active
row.

## Run

From the repository root:

```sh
go run ./examples/builder/loglens
go run ./examples/builder/loglens -web
go run ./examples/builder/loglens -headless
go run ./examples/builder/loglens -dev
```

| Flag | Default | Effect |
|---|---|---|
| `-web` | `false` | Serve the dashboard through `ownframe.Serve` instead of opening a desktop window. |
| `-headless` | `false` | Alias for `-web`; serve the dashboard without a native window. |
| `-addr` | `127.0.0.1:8132` | Listen address for `-web`. |
| `-dev` | `false` | Open the developer tools dock at startup through `Page.Window().SetDevTools`; dock sampling also feeds `-debug-addr`. |
| `-perf` | `false` | Enable perf sampling for the devtools Frame tab and the `-debug-addr` endpoints. |
| `-debug-addr` | `""` | Serve window profiling (`/debug/state`, `/debug/pprof/*`) on this loopback address; off by default, needs `-tags ownframe_debug`. |

The window opens at 1440 × 960 with a minimum of 1100 × 720. It starts in the
dark palette with a hidden scrollbar. Native runs use custom window chrome: the
top bar is the drag region and draws minimize, maximize, and close controls.

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

`data.go` holds seed fixtures and the log query; `feed.go` maintains the bounded
generated log and trace stream. The initial sample has 24 entries,
`sampleServices` 6, `sampleSpans` 5, and `sampleMessages` 5. `visibleLogs`
filters by service, severity, and a case-insensitive query over the message,
service, and trace ID. `pageLogs` slices 10 rows at a time with newest entries
first, `logPageCount` sizes the pagination, and `selectedLog` resolves the
detail drawer's row. The option lists for time range, service, severity, sort,
and notification channel live here too.

| File | Responsibility |
|---|---|
| `main.go` | Flags, context, first draw, `Run` or `Serve`. |
| `app.go` | App state, mount, handlers, theme switching, toast and overlay helpers, form access. |
| `shell.go` | Screen root, top bar, sidebar, page header, tab set, tab titles and subtitles. |
| `data.go` | Seed fixtures, option lists, log filtering, paging, and feed-derived metrics. |
| `feed.go` | Bounded generated demo telemetry, newest-first log and trace pages, and changing region samples. |
| `overview.go` | Landing tab: feed-derived alert, KPI cards, event chart, service table, recent events, on-call calendar, and regions. |
| `logs.go` | Request stream: filter bar, error alert, toolbar, row menu, data table, and a payload/trace detail drawer. |
| `filters.go` | Filter bar: time-range select, service combobox, date picker, query and search inputs, sort, severity, live tail, auto refresh, context, sample rate, apply and reset. |
| `traces.go` | Selected trace waterfall and payload, ten-row pagination, one RTL direction example, and owner actions. |
| `incident.go` | Incident header, chat, decisions, action items, playbook, snapshots, customer note, impact map, timeline, postmortem. |
| `settings.go` | Project settings sheet: tabbed routing, noise control, and verification sections. |
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
- **Hover state.** The Live tooltip and On-call card use ChadCN's `CSSHover`
  mode. `handleHover` toggles a class on their mounted content with local page
  patches, so normal pointer movement does not rebuild the dashboard. It only
  rebuilds the view when it also needs to close another open top-bar menu.
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
  not deserve a field: the settings-plan override and the incident
  accordion seeding guard.

## Component map

The component coverage test checks all 64 component roots currently rendered
by LogLens. The standalone Scroll Area stays in the UI Builder gallery so the
logs and traces views can use pagination without nested scrolling. The
canonical component list is the package README at
`components/chadcn/README.md`. The file column names where the constructor is
called; a second location is shown when the same component is reused.

### Primitives and display

| Component | File | Where it appears |
|---|---|---|
| **Aspect Ratio** | overview.go | Edge-regions mesh and impact map in incident.go. |
| **Avatar** | shell.go | Sidebar footer; trace, incident, and popover rows. |
| **Badge** | shell.go | Page title; status, severity, count, and level badges across every panel. |
| **Card** | overview.go | KPI cards and every content card; reused by logs, traces, and incident. |
| **Direction** | traces.go | Arabic direction toggle; customer note in incident.go. |
| **Empty** | logs.go | No-match logs and the postmortem placeholder. |
| **Item** | overview.go | Recent events; trace owner and incident action items. |
| **Kbd** | shell.go | Page header shortcut; shortcut lists and dialog footers. |
| **Label** | filters.go | Filter captions, threshold labels, and form labels in settings.go. |
| **Progress** | overview.go | KPI progress bars; latency budget in traces.go and mitigation in incident.go. |
| **Separator** | incident.go | Response playbook separator; reused in logs, settings, dialogs, and popovers. |
| **Skeleton** | incident.go | Postmortem draft loading state. |
| **Spinner** | incident.go | Postmortem draft loading state. |
| **Table** | overview.go | Service-health table; span attributes and the incident timeline. |
| **Typography** | incident.go | H3/H4 and InlineCode in the incident room. |

### Actions and forms

| Component | File | Where it appears |
|---|---|---|
| **Button** | shell.go | Page actions; the primary control in every panel. |
| **Button Group** | shell.go | Refresh/Export pair; chart actions, log actions, trace selector, incident actions, filters, and settings presets. |
| **Toggle** | filters.go | Live tail and the logs toolbar. |
| **Toggle Group** | filters.go | Severity selection. |
| **Field** | filters.go | Query field; alert fields and settings fields in dialogs.go and settings.go. |
| **Input** | filters.go | Query box; alert email, settings email, and the incident reply. |
| **Textarea** | settings.go | Handover notes. |
| **Native Select** | filters.go | Sort order; notification channel in settings.go. |
| **Checkbox** | filters.go | Show context; mute resolved incidents and page the on-call. |
| **Radio Group** | settings.go | Minimum severity; alert severity in dialogs.go. |
| **Switch** | filters.go | Auto refresh; alerting, RTL mode, and alert toggles. |
| **Slider** | filters.go | Sample rate; settings threshold and the dialog error-rate threshold. |
| **Input Group** | filters.go | Log search prefix and row-count suffix; webhook prefix in settings.go. |
| **Input OTP** | settings.go | Authenticator code; alert verification code in dialogs.go. |

### Feedback and disclosure

| Component | File | Where it appears |
|---|---|---|
| **Alert** | overview.go | Feed-derived error summary; error-share alert above the log table. |
| **Toast** | dialogs.go | Mounted once in `globalDialogs`; every `showToast` writes its title and body. |
| **Accordion** | incident.go | Three-step response playbook with a multi-open set. |
| **Collapsible** | incident.go | Advanced notes below the playbook. |
| **Tabs** | shell.go | Overview, Logs, Traces, and Incident room. |
| **Breadcrumb** | shell.go | `Acme › Production › <tab>` above the page title. |
| **Pagination** | logs.go | Ten-row log pages; traces.go also pages ten traces. |

### Overlays and menus

| Component | File | Where it appears |
|---|---|---|
| **Dialog** | dialogs.go | Create-alert rule editor. |
| **Alert Dialog** | dialogs.go | Delete-saved-view confirmation. |
| **Sheet** | settings.go | Right-side project settings panel. |
| **Drawer** | logs.go | Log-detail drawer with Payload and Trace tabs. |
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
| **Resizable** | traces.go | Waterfall and payload split at a fixed 0.55 ratio. |
| **Sidebar** | shell.go | Project navigation with workspace and saved-view rows. |
| **Carousel** | incident.go | Three feed-derived sample snapshots with previous/next controls. |
| **Data Table** | logs.go | The request stream with row selection. |
| **Chart** | overview.go | Newest ten event counts compared with the previous ten. |
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
| **HoverCard** | CSS `:hover` | Hover on the "On call: Ada" trigger (`hover-trigger`). |
| **NavigationMenu** | `menubarOpen` | File, View, and Help in the top bar. |
| **NavigationMenu** | `navOpen` | Platform and Resources in the top bar; Settings opens the sheet instead. |
| **Popover** | `notifOpen` | Notifications button in the top bar. |
| **Select** | `selectOpen` | Time-range picker in the filter bar. |
| **Sheet** | `settingsOpen` | Project settings: sidebar gear, navigation menu Settings, account menu Project settings, command row, or `settings:open`. |
| **Toast** | `toastOpen`, `toastTitle`, `toastBody` | Mounted in `globalDialogs`; `showToast` opens it from Refresh, Export, alert saves, filter apply/reset, and most incident and trace actions. |
| **Tooltip** | CSS `:hover` | Hover on the Live trigger (`tooltip-trigger`). |

The command palette is controlled by `commandOpen`, `commandQuery`, and
`commandActive`. The ⌘K button in the top bar opens it; Escape closes it along
with every other overlay. The ⌘K label is a hint, not a global keybinding.

## Limits

- **No backend.** `feed.go` creates generated demo logs and traces from the
  bundled seed fixtures. Refresh and one-second Auto refresh append one to
  three records, update overview metrics and regional samples, and keep at most
  500 records. The app makes no network, `fetch`, or IPC call; this is changing
  showcase data, not production telemetry.
- **Host-owned behavior.** Sorting, filtering, paging, row selection, carousel
  position, accordion state, and every open flag are application code. The
  components only render the values they receive.
- **Auto refresh scope.** Auto refresh runs while enabled and appends generated
  records once per second. Manual refresh also returns logs and traces to page
  one, where the newest records appear first.
- **No JavaScript or canvas.** The chart is plain div geometry with CSS bars.
  The region mesh and impact map are div and span geometry. There is no live DOM,
  no `<a href>` navigation, and no script tag.
- **Fixed split.** `Resizable` renders the divider, but the example never routes
  a drag to `splitRatio`, so the waterfall stays at 0.55.
- **Overlay focus is partial.** Escape closes the tracked overlays and each
  component closes on its backdrop callback. The example does not install a
  focus trap around the dialog, sheet, or drawer.
- **Scroll is page-level.** Logs and traces use pagination; the incident chat
  keeps its fixed-height message viewport. The window scrolls the page.
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
all 64 rendered `.of-` component roots as whole class tokens. The `screen` render is what
contributes `.of-sidebar`, `.of-breadcrumb`, and `.of-tabs`; `chadcn.Screen`
itself has no `.of-screen` class. The test needs no window and no running app.

`make build` runs `go vet -p 1 ./... ./examples/...`, which compiles every
example without linking binaries. `make test` runs both Go modules. `make lint`
runs the pinned `golangci-lint` over both modules plus the documentation, plan,
and website checks.
