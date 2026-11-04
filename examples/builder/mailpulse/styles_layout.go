package main

// layoutFixCSS is the heat-readout, funnel-alignment, and narrow-window
// companion to appCSS. The integrator appends it to the theme after hoverCSS
// in newApp and themeSheet (see LAYOUT_REPORT.md). It only overrides layout
// properties; colors stay in appCSS. Grids inside .mp-stack, .mp-ov,
// .mp-cont, and .mp-camp hold separated cards with their own borders, so the
// narrow single-column rules below add no divider seams there: each cell
// keeps its full border and curves and the 16px gap separates the rows.
const layoutFixCSS = `
.mp-heat-readout {
  padding: 6px 12px;
  min-height: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mp-funnel-row { padding: 7px 4px 7px 0; }
.mp-funnel-top { min-width: 0; align-items: baseline; }
.mp-funnel-label {
  flex: 1 1 auto;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mp-funnel-stats { flex-shrink: 0; margin-left: auto; min-width: 132px; justify-content: flex-end; }
.mp-funnel-count { min-width: 64px; text-align: right; }
.mp-funnel-pct { min-width: 56px; text-align: right; }
.mp-funnel-row > .mp-dots { width: 100%; max-width: 100%; overflow: hidden; padding-right: 2px; box-sizing: border-box; }
@media (max-width: 1100px) {
  .mp-content { padding: 0 16px 24px; }
  .mp-grid-mid { grid-template-columns: 1fr; }
  .mp-grid-bot { grid-template-columns: 1fr; }
  .mp-card.mp-stack > .mp-grid-mid > *,
  .mp-card.mp-stack > .mp-grid-bot > *,
  .mp-ov > .mp-grid-mid > *,
  .mp-ov > .mp-grid-bot > *,
  .mp-cont > .mp-grid-bot > * {
    border: 1px solid #e8e8e6;
  }
  .mp-kpis { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .mp-kpi { min-width: 0; border-left: none; border-top: 1px solid #eee; }
  .mp-kpi:first-child { border-top: none; }
  .mp-kpi:nth-child(2) { border-top: none; }
  .mp-kpi:nth-child(even) { border-left: 1px solid #eee; }
}
`
