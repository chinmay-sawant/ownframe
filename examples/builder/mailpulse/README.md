# MailPulse · engagement demo

MailPulse is an email engagement dashboard built with the `ui` builder
package and the ChadCN component set in `components/chadcn`. It shows a
campaign overview card: KPI tiles with delta pills, a mid grid with an
open-rate heatmap beside a device funnel, and a bottom grid with top
links and an insight callout.

The host owns all state and layout. ChadCN constructors render the
sidebar, top bar controls, KPI cards, funnel bars, and link rows; the
`mp-*` classes in `styles.go` only arrange them on the light `#f6f6f5`
page. App-specific tiles that ChadCN does not ship (heatmap cells,
engagement dots, delta pills) live in the local `components/` package.

## Run

From the repository root:

```sh
go run ./examples/builder/mailpulse
go run ./examples/builder/mailpulse -web
```

From the `examples/` module directory:

```sh
go run ./builder/mailpulse -web
```

`-web` serves the dashboard through `ownframe.Serve` for the PNG
browser preview instead of opening a desktop window.

## Screenshot

The window shows a left sidebar, a top bar with search and actions, and
a single white card: a page head with title and date range, four KPI
tiles (opens, clicks, replies, bounces) with colored icon dots and
green/red delta pills, an 18-column purple open heatmap with a day
column beside a funnel with engagement dots, then top links with
per-link dots above an insight callout on a lavender background.
