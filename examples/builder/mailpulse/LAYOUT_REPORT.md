# MailPulse layout fix: heat readout + funnel alignment + narrow compact

## What was checked (read-only; no Go file edited)

- `heatmap.go:150` — readout element `#mp-heat-readout.mp-heat-readout`, text from `heatCellLabel` (e.g. `Tue 9a · peak open rate · about 68%`).
- `funnel.go:16-25` — `.mp-funnel-row > .mp-funnel-top > (.mp-funnel-label + .mp-funnel-stats > .mp-funnel-count + .mp-funnel-pct)` + `DotBar` (20 dots) below.
- `styles.go:137-142` — funnel rules; `styles.go:160-170` — readout rules; `styles.go:81-148` — `.mp-kpis`, `.mp-grid-mid`, `.mp-grid-bot`, dividers, `.mp-content`.
- `tabs.go:103-116` — engagement stack: pageHead, kpiRow, grid-mid (heatmap + funnel), grid-bot (links + engage).
- `app.go:39-49` — `MinWidth: 1180`, `Width: 1440`, theme = Light + Stylesheet + appCSS + hoverCSS.
- `documentation/compatibility-matrix.md` — `@media` size features vs viewport supported (`@media` rows); `white-space: nowrap`, `text-overflow: ellipsis`, `min-height`, `grid-template-columns`, `:nth-child(odd/even/an+b/integer)` all Implemented.

## Screenshots

- `/tmp/mp-1440.png` — live `go run ./builder/mailpulse -web -addr 127.0.0.1:8131`, `GET /frame.png`, 1440x960, engagement tab, pre-fix baseline (`layoutFixCSS` not yet wired):
  - Readout (`Tue 9a · peak open rate · about 68%`): visibly TALL — `padding: 8px 12px` + `line-height: 1.5` + `min-height: 34px`. Confirms the "too tall" report.
  - Funnel at 1440: count + pct share one baseline, pct right-aligned in a stable column; `Upgraded to paid` does NOT push `0.2%` out at this width; 20-dot bars sit full-width below each row. Alignment correct here — but fragile (see below), so the fix strengthens it.
  - Dividers: grid-bot has a vertical divider (`*:nth-child(2)` border-left); grid-mid has NONE — the `.mp-funnel { border-left: ... }` rule in `styles.go:136` matches no element (funnel card carries no `mp-funnel` class). Cosmetic only; left as-is outside the media query.
- Narrow (1000px): NOT screenshotted. The `-web` frame follows page `Width/Height` config (fixed 1440x960) and resizing it needs a Go edit, which is out of scope for this workstream. Narrow behavior is covered by media-query correctness (matrix: size features evaluated vs viewport) plus the post-merge manual check below.

## Fix delivered: `styles_layout.go` (`layoutFixCSS`)

1. Readout — exact spec: `padding: 6px 12px; min-height: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis`. Colors untouched.
2. Funnel — `.mp-funnel-top { min-width: 0 }`; label gets `flex: 1 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis` (without `min-width: 0` the label's auto minimum width plus `flex-shrink: 0` stats would shove the pct column at narrow widths); stats pinned `flex-shrink: 0`; pct keeps `min-width: 52px; text-align: right`; dots forced full-width below with `width/max-width: 100%; overflow: hidden` so 20 fixed dots clip instead of overlapping.
3. Narrow `@media (max-width: 1100px)` — grids to `1fr` single column; second-child dividers left→top (`border-left: none` + `border-top`); dead `.mp-funnel` border cleared; `.mp-kpis` to 2 columns with divider rework (no left border except even children, no top border on first row) and `min-width: 0` so tiles fit (~350px each at 1000px window: 1000 − 248 sidebar − 32 content padding − card padding); `.mp-content` padding to `0 16px 24px`; `.mp-heat-grid` untouched (18 `minmax(0,1fr)` tracks already shrink; ~27px cells at 1000px). `gofmt` clean, `go vet` passes.

## Go changes requested (integrator edits `app.go`)

1. Wire the sheet — two spots, same edit:
   - `newApp`: `Theme: chadcn.Light + chadcn.Stylesheet + appCSS + hoverCSS + layoutFixCSS`
   - `themeSheet()`: `return chadcn.Light + chadcn.Stylesheet + appCSS + hoverCSS + layoutFixCSS`
2. Recommend `MinWidth: 1180` → `MinWidth: 1000` (`app.go:43`) so the 1000px window lays out at 1000 instead of horizontal-scrolling a 1180 floor. Sidebar stays fixed 248px (inline `Width(ui.Px(248))`, not touched); right column ≈ 752px at 1000, which fits the stacked rules above. Goal: no overlap at 1000px width.
3. Lint note: until (1) lands, `layoutFixCSS` is an unreferenced const and the `unused` gate may flag it — expected, resolves on merge.

## Post-merge verification (narrow)

Native `Run` (not `-web`): resize the desktop window to 1000px wide and confirm engagement shows single-column mid/bot grids with top dividers, 2-column KPIs, single-line readout with ellipsis on long hover labels, funnel pct column steady while hovering cells (readout text swaps via `PatchElement`), heat cells shrinking without overlap. Optional: open the devtools dock (F12) at 1440 to spot-check the stacked rules under a narrower content strip.
