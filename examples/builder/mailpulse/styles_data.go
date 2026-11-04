package main

// dataFixCSS tightens the Analytics, Reports, and Learn lanes without
// touching styles.go. Another lane wires this sheet into the theme after
// appCSS and hoverCSS, so these rules win at equal specificity.
//
// Analytics: the 12 Opens-by-week bars share one flex row with the
// Clicks-by-client card beside it, so every bar flexes, never forces a
// minimum width, and clips instead of overlapping at narrow columns.
// Reports: rows get generous padding, a wide meta-to-button gap, and a
// divider between rows only. Learn: the duration line owns the flexible
// meta while the Start button never shrinks, so the two never collide.
const dataFixCSS = `
.mp-bars { gap: 4px; overflow: hidden; min-width: 0; }
.mp-bar-col { flex: 1 1 0; min-width: 0; }
.mp-bar { min-height: 0; max-width: 40px; }
.mp-bar-col .mp-muted { font-size: 11px; white-space: nowrap; }
.mp-link-meta { display: flex; flex-direction: column; gap: 6px; min-width: 0; flex: 1; }
.mp-report-row { padding: 16px 8px; gap: 20px; }
.mp-report-row + .mp-report-row { border-top: 1px solid #f0f0ee; }
.mp-report-row:first-child { border-top: none; }
.mp-report-row .mp-engage-label { font-size: 13px; font-weight: 600; margin-bottom: 0; }
.mp-linkrow .of-button { flex-shrink: 0; }
.mp-learn-row { padding: 12px 0; gap: 12px; }
.mp-learn-row .mp-linkmain { flex: 1; min-width: 0; }
.mp-card .of-table { width: 100%; table-layout: fixed; }
.mp-card .of-table th, .mp-card .of-table td { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mp-card .of-table th:last-child, .mp-card .of-table td:last-child { width: 112px; min-width: 0; }
.of-table .mp-dots { flex-wrap: nowrap; min-width: 0; max-width: 100%; overflow: hidden; }
.of-table .mp-dot { flex: 1 1 0; min-width: 6px; max-width: 14px; width: auto; height: 8px; border-radius: 4px; }
@media (max-width: 1100px) { .mp-card .of-table th:last-child, .mp-card .of-table td:last-child { width: 96px; } .of-table .mp-dot { min-width: 5px; } }
`
