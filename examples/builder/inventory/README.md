# Inventory (Northwind Goods)

Light-theme inventory-health dashboard built with the Go builder (`ui` + `components/chadcn`), modelled on the warehouse screenshot: KPI strip, stock-health table, bin-utilization heat grid, inbound POs, and value-by-category bar.

## Run

```sh
cd examples
go run ./builder/inventory -web -addr 127.0.0.1:8135
go run ./builder/inventory              # desktop window
```

## Layout

| File | Owns |
|---|---|
| `app.go` | `app` state, `ui.Mount`, `screen()` shell |
| `main.go` | flags, `Serve` vs `RunWithOptions` |
| `data.go` | SKU, PO, bin-grid, category fixtures |
| `shell.go` | topbar (brand, nav, search box, alerts, account) + `Inventory health` page header |
| `search.go` | search palette (`chadcn.Command` panel below the search box) |
| `kpis.go` | 5-KPI strip |
| `stock.go` | stock-health table |
| `bins.go` | bin-utilization heat grid |
| `inbound.go` | inbound PO timeline |
| `dialogs.go` | New PO dialog (`po:new` / `po:close` / `po:create`) |
| `hover.go` | bin hover handler (patches the `bin-detail` line) |
| `categories.go` | stacked value-by-category bar |
| `styles.go` | layout-only `appCSS` (ChadCN owns paint) |

Period pills (`30 days` / `90 days` / `12 months`) re-render through `handleClick`; everything else is static fixture data.

## Chrome and scrolling

LockView fit-to-window: the page mounts at a 1440×960 design canvas
(`Width`/`Height` in `app.go`) with `LockView: true`, so the canvas scales
to fit any window with no page scroll and no zoom — including small windows
down to the 900×600 minimum and large ones such as 1920×1030. The minimum
is 900×600 (`MinWidth`/`MinHeight`); below the design size the canvas
compresses instead of clipping.

Static header: page scroll is off and the shell is a `100vh` flex column with
the topbar in normal flow. The inner body viewport scrolls via the `Wheel`
handler (`scroll.go`). The desktop drag strip is the top band
(`y < inventoryTopbarHeight` in `main.go`); window controls win over drags
because OwnFrame skips a drag that starts on a button.

Window controls `−/□/×` sit at the right end of the topbar (`shell.go`) and
emit the `win-min`, `win-max`, and `win-close` actions.

## Bin heatmap

Cells shade red→green by cubic-capacity utilization, and each cell exposes its
aisle, level, and percent as hover text (`title` tooltip). The legend shows
the Empty→Full stops, and the summary line calls out the hottest- and
coldest-aisle averages.

Bin hover detail line: hovering a cell replaces the summary line with that
cell's aisle, level, percent-full, and status (`binCellTitle`), patched
directly into the `#bin-detail` element via `PatchElement` (`hover.go`); a
refused patch falls back to a full refresh. Moving off the grid restores
the hottest/coldest summary. Cells carry no `title` tooltip — the detail
line is the single hover readout, so no extra box pops over the number —
and each cell keeps an `aria-label` for assistive tech.

## Search palette

The topbar search box (`Search` + `⌘K` hint, `search.go`) toggles a command
palette panel directly below it (`#inv-search-panel`, `chadcn.Command`).
The panel has two groups — matching SKUs (max 5, by name/code/category)
and matching purchase orders (by vendor/number) — filtered live as you
type (case-insensitive substring; empty query shows the first five SKUs
and all orders). While open, a transparent click-away backdrop
(`.inv-search-backdrop`, `search:close`) covers the page, so clicking
anywhere outside the panel closes it, mirroring the mygo Popover/modal
backdrop. Picking a row closes the panel and pins a `Picked: <code>`
line under the search box (mygo `picked` behavior); `Esc` closes it too.

## Stock status filter

The `Stock health` pills (`stock.go`, `statusPills`) work like the mygo
reference: `All SKUs 4,200`, `Stockout 7`, `Reorder 16`, `Overstock 31`
emit `status:<filter>` actions, the table shows only matching rows, and
the active pill carries the dark-green fill (`#1d5c3f` on `#fafaf7`),
exactly the mygo active-pill highlight. `All SKUs` clears the filter.

## New PO dialog

`+ New PO` opens a `chadcn.Dialog` (`dialogs.go`, `po:new`) with the vendor,
items, and ETA rows plus a draft note. `Cancel`, the backdrop/close
control (`po:close`), creating (`po:create`), and `Esc` all close it.

## Alerts popover and account menu

The bell is an `Alerts` text button with a red unread dot (`●`), opening a
`chadcn.Popover` with the two zero-on-hand stock alerts and an update stamp
(`shell.go`, `alertsPopover`). The `AK` avatar opens a `chadcn.DropdownMenu`
with a Profile item and a Session group holding the danger `Sign out` row
(`accountMenu`). `Esc` closes both.

## Compact layout

Density is dashboard-compact (`styles.go`): 14px base type (12–13px for
subtitles, hints, and nav), 16px card padding, 10–14px gaps, a two-column
grid (stock + categories left, bins + inbound right), and a fixed 60px
topbar — everything fits the 1440×960 canvas with no page scroll.

## Overlay drag guard

The topbar is the CustomChrome drag strip (`Draggable`, `main.go`). Presses
on overlay padding carry no action, so without a guard a click there would
start a window drag instead of reaching the overlay. `Draggable` returns
false while any overlay is open (`overlaysOpen`: search palette,
notification/account popovers, PO dialog, export toast), so open overlays
own every press. Elements with `data-action` are controls by definition
and never start drags.

## Inbound selection and export toast

Inbound rows are selectable (`po:view:<PONo>` in `inbound.go`): the
selected row carries the green fill and the card header shows that order's
number, vendor, and amount. `View all` (`po:viewall`) clears the selection
and restores the open-order totals. `Export` (`export:run` in `stock.go`)
raises a bottom-right `chadcn.Toast` (`exportToast` in `dialogs.go`,
mirroring the loglens toast); its close control and `Esc` dismiss it.
