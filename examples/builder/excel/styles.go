package main

import "github.com/chinmay-sawant/ownframe/components/chadcn"

// Fixed geometry shared by every theme. The grid viewport takes whatever the
// window leaves after these bands.
const (
	titleBarHeight   = 36
	ribbonHeightFull = 88
	ribbonHeightMin  = 32
	formulaBarHeight = 28
	sheetBarHeight   = 28
	statusBarHeight  = 24

	rowHeaderWidth  = 40
	colHeaderHeight = 24
	scrollbarSize   = 12
	minColWidth     = 24
	minRowHeight    = 16

	windowWidth  = 1440
	windowHeight = 960
	minWinWidth  = 900
	minWinHeight = 600
)

// xlLightTokens and xlDarkTokens are the --xl-* spreadsheet tokens layered on
// top of chadcn.Light and chadcn.Dark.
//
//nolint:gosec // G101 false positive: CSS custom-property names, not credentials.
const xlLightTokens = `
:root {
 --xl-brand:#107c41; --xl-brand-hover:#0c6634; --xl-brand-strong:#107c41; --xl-brand-tint:#e6f4ea;
 --xl-titlebar-bg:#107c41; --xl-titlebar-fg:#ffffff;
 --xl-ribbon-bg:#f3f2f1; --xl-ribbon-group-bg:#ffffff; --xl-ribbon-border:#e1dfdd;
 --xl-formula-bg:#ffffff; --xl-sheet-bg:#ffffff; --xl-cell-text:#1f2328; --xl-cell-muted:#605e5c;
 --xl-gridline:#d4d4d4; --xl-gridline-frozen:#8a8886;
 --xl-header-bg:#f3f2f1; --xl-header-fg:#323130; --xl-header-active-bg:#d2e8db; --xl-header-active-fg:#107c41;
 --xl-selection-fill:#107c411f; --xl-selection-border:#107c41; --xl-cut-dash:#107c41;
 --xl-tabbar-bg:#f3f2f1; --xl-tab-active-bg:#ffffff; --xl-statusbar-bg:#f3f2f1; --xl-statusbar-fg:#323130;
 --xl-error-fg:#c42b1c; --xl-comment-corner:#c42b1c; --xl-negative-num:#b91c1c;
 --xl-ref-1:#2563eb; --xl-ref-1-fill:#2563eb1a; --xl-ref-2:#dc2626; --xl-ref-2-fill:#dc26261a;
 --xl-ref-3:#7c3aed; --xl-ref-3-fill:#7c3aed1a; --xl-ref-4:#0891b2; --xl-ref-4-fill:#0891b21a;
 --xl-ref-5:#d97706; --xl-ref-5-fill:#d977061a; --xl-ref-6:#db2777; --xl-ref-6-fill:#db27771a;
 --xl-cf-red-bg:#ffc7ce; --xl-cf-red-fg:#9c0006; --xl-cf-yellow-bg:#ffeb9c; --xl-cf-yellow-fg:#9c6500;
 --xl-cf-green-bg:#c6efce; --xl-cf-green-fg:#006100;
 --xl-bar-blue:#638ec6; --xl-bar-green:#63c384; --xl-bar-red:#ff555a; --xl-bar-amber:#ffb628;
 --xl-scale-low:#f8696b; --xl-scale-mid:#ffeb84; --xl-scale-high:#63be7b;
 --of-primary:#107c41; --of-on-primary:#ffffff; --of-primary-hover:#0c6634;
}
`

//nolint:gosec // G101 false positive: CSS custom-property names, not credentials.
const xlDarkTokens = `
:root {
 --xl-brand:#185c37; --xl-brand-hover:#217346; --xl-brand-strong:#33c481; --xl-brand-tint:#163826;
 --xl-titlebar-bg:#18181b; --xl-titlebar-fg:#f4f4f5;
 --xl-ribbon-bg:#202024; --xl-ribbon-group-bg:#27272b; --xl-ribbon-border:#35353b;
 --xl-formula-bg:#18181b; --xl-sheet-bg:#141417; --xl-cell-text:#f4f4f5; --xl-cell-muted:#a1a1aa;
 --xl-gridline:#2d2d33; --xl-gridline-frozen:#71717a;
 --xl-header-bg:#222226; --xl-header-fg:#d4d4d8; --xl-header-active-bg:#1d3b2b; --xl-header-active-fg:#4ade80;
 --xl-selection-fill:#33c48129; --xl-selection-border:#33c481; --xl-cut-dash:#4ade80;
 --xl-tabbar-bg:#1c1c20; --xl-tab-active-bg:#27272b; --xl-statusbar-bg:#18181b; --xl-statusbar-fg:#d4d4d8;
 --xl-error-fg:#f87171; --xl-comment-corner:#fb7185; --xl-negative-num:#f87171;
 --xl-ref-1:#60a5fa; --xl-ref-1-fill:#60a5fa24; --xl-ref-2:#f87171; --xl-ref-2-fill:#f8717124;
 --xl-ref-3:#a78bfa; --xl-ref-3-fill:#a78bfa24; --xl-ref-4:#22d3ee; --xl-ref-4-fill:#22d3ee24;
 --xl-ref-5:#fbbf24; --xl-ref-5-fill:#fbbf2424; --xl-ref-6:#f472b6; --xl-ref-6-fill:#f472b624;
 --xl-cf-red-bg:#4c1d24; --xl-cf-red-fg:#fca5a5; --xl-cf-yellow-bg:#4a3712; --xl-cf-yellow-fg:#fcd34d;
 --xl-cf-green-bg:#143d24; --xl-cf-green-fg:#86efac;
 --xl-bar-blue:#638ec6; --xl-bar-green:#63c384; --xl-bar-red:#ff555a; --xl-bar-amber:#ffb628;
 --xl-scale-low:#f8696b; --xl-scale-mid:#ffeb84; --xl-scale-high:#63be7b;
 --of-primary:#33c481; --of-on-primary:#0b2a1a; --of-primary-hover:#4ade80;
}
`

// ExcelLight and ExcelDark are the full theme sheets: chadcn base, --xl-*
// tokens, then the spreadsheet CSS.
//
//nolint:gochecknoglobals // chadcn.Stylesheet is a loaded var, so these cannot be consts; tests read them by name.
var (
	ExcelLight = chadcn.Light + chadcn.Stylesheet + xlLightTokens + appCSS + ribbonCSS
	ExcelDark  = chadcn.Dark + chadcn.Stylesheet + xlDarkTokens + appCSS + ribbonCSS
)

const appCSS = `
.xl-shell {
  position: relative; display: flex; flex-direction: column;
  height: 100vh; min-height: 100vh; overflow: hidden;
  background: var(--xl-sheet-bg); color: var(--xl-cell-text);
  font-family: sans-serif; font-size: 13px;
}
.xl-title {
  display: flex; align-items: center; gap: 8px; flex-shrink: 0;
  height: 36px; padding: 0 6px 0 10px;
  background: var(--xl-titlebar-bg); color: var(--xl-titlebar-fg);
}
.xl-badge {
  display: flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border-radius: 4px;
  background: #ffffff; color: #107c41; font-weight: 700; font-size: 13px;
}
.xl-brand-name { font-weight: 700; font-size: 13px; color: var(--xl-titlebar-fg); }
.xl-qat { display: flex; align-items: center; gap: 2px; margin-left: 6px; }
.xl-tbtn {
  display: flex; align-items: center; justify-content: center;
  min-width: 28px; height: 28px; padding: 0 8px; border-radius: 4px;
  border: 0; background: transparent; color: var(--xl-titlebar-fg); font-size: 14px;
}
.xl-tbtn:hover { background: #ffffff2e; }
.xl-tbtn--off { opacity: 0.45; }
.xl-doc-pill {
  padding: 3px 12px; border-radius: 4px; font-size: 12px; font-weight: 600;
  background: #ffffff1f; color: var(--xl-titlebar-fg); white-space: nowrap;
}
.xl-title-spacer { flex: 1; }
.xl-search {
  display: flex; align-items: center; gap: 6px; width: 260px; height: 26px; padding: 0 10px;
  border-radius: 4px; background: #ffffff26; color: var(--xl-titlebar-fg); font-size: 12px;
}
.xl-winctl { display: flex; align-items: center; gap: 0; flex-shrink: 0; margin-left: 6px; }
.xl-winbtn {
  display: flex; align-items: center; justify-content: center;
  width: 40px; height: 36px; border: 0; background: transparent;
  color: var(--xl-titlebar-fg); font-size: 15px;
}
.xl-winbtn:hover { background: #ffffff2e; }
.xl-ribbon-slot {
  box-sizing: border-box; position: relative; z-index: 30; flex-shrink: 0; overflow: visible;
  background: var(--xl-ribbon-bg); border-bottom: 1px solid var(--xl-ribbon-border);
}
.xl-ribbon-stub { display: flex; flex-direction: column; height: 100%; }
.xl-ribbon-tabs { display: flex; align-items: center; gap: 2px; height: 30px; padding: 0 8px; }
.xl-ribbon-tab {
  padding: 4px 12px; border-radius: 4px; font-size: 12px; color: var(--xl-header-fg);
}
.xl-ribbon-tab--active { background: var(--xl-brand-tint); color: var(--xl-brand); font-weight: 700; }
.xl-ribbon-body { padding: 6px 12px; font-size: 12px; color: var(--xl-cell-muted); }
.xl-fbar {
  box-sizing: border-box; display: flex; align-items: center; gap: 4px; flex-shrink: 0;
  height: 28px; padding: 0 6px;
  background: var(--xl-ribbon-bg); border-bottom: 1px solid var(--xl-ribbon-border);
}
.xl-namebox, .xl-fx {
  box-sizing: border-box; height: 22px; padding: 0 6px; border-radius: 2px;
  border: 1px solid var(--xl-ribbon-border);
  background: var(--xl-formula-bg); color: var(--xl-cell-text); font-size: 13px;
}
.xl-namebox { width: 96px; flex-shrink: 0; }
.xl-fx { flex: 1; min-width: 0; }
.xl-fsep { width: 1px; height: 16px; background: var(--xl-ribbon-border); }
.xl-fbtn {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 22px; border: 0; border-radius: 3px; background: transparent;
  color: var(--xl-header-fg); font-size: 13px;
}
.xl-fbtn:hover { background: var(--xl-brand-tint); }
.xl-fbtn--fx { font-style: italic; font-weight: 700; }
.xl-grid {
  position: relative; flex: 1; min-height: 0; min-width: 0; overflow: hidden;
  background: var(--xl-sheet-bg);
}
.xl-canvas { position: absolute; left: 40px; top: 24px; right: 0; bottom: 0; overflow: hidden; }
.xl-hline { position: absolute; left: 0; height: 1px; background: var(--xl-gridline); }
.xl-vline { position: absolute; top: 0; width: 1px; background: var(--xl-gridline); }
.xl-frozen-v { position: absolute; top: 0; bottom: 0; width: 2px; background: var(--xl-gridline-frozen); }
.xl-frozen-h { position: absolute; left: 0; right: 0; height: 2px; background: var(--xl-gridline-frozen); }
.xl-cell {
  position: absolute; box-sizing: border-box; overflow: hidden; white-space: nowrap;
  padding: 0 4px; line-height: 21px; font-size: 13px; color: var(--xl-cell-text);
}
.xl-cell--text { text-align: left; }
.xl-cell--num { text-align: right; }
.xl-cell--center { text-align: center; }
.xl-cell--left { text-align: left; }
.xl-cell--right { text-align: right; }
.xl-cell--err { color: var(--xl-error-fg); text-align: center; }
.xl-cell--wrap { white-space: normal; line-height: 15px; }
.xl-b { font-weight: 700; }
.xl-i { font-style: italic; }
.xl-u { text-decoration: underline; }
.xl-s { text-decoration: line-through; }
.xl-us { text-decoration: underline line-through; }
.xl-bar { position: absolute; left: 0; top: 2px; bottom: 2px; opacity: 0.75; }
.xl-note {
  position: absolute; top: 0; right: 0; width: 0; height: 0;
  border-top: 6px solid var(--xl-comment-corner); border-left: 6px solid transparent;
}
.xl-dd {
  position: absolute; box-sizing: border-box; display: flex; align-items: center; justify-content: center;
  width: 16px; background: var(--xl-header-bg); border: 1px solid var(--xl-gridline);
  color: var(--xl-header-fg); font-size: 9px;
}
.xl-sel { position: absolute; box-sizing: border-box; background: var(--xl-selection-fill); border: 1px solid var(--xl-selection-border); }
.xl-active { position: absolute; box-sizing: border-box; border: 2px solid var(--xl-selection-border); }
.xl-handle { position: absolute; width: 6px; height: 6px; background: var(--xl-brand-strong); border: 1px solid var(--xl-sheet-bg); }
.xl-cut { position: absolute; box-sizing: border-box; border: 2px dashed var(--xl-cut-dash); }
.xl-ref { position: absolute; box-sizing: border-box; }
.xl-ref-1 { border: 2px solid var(--xl-ref-1); background: var(--xl-ref-1-fill); }
.xl-ref-2 { border: 2px solid var(--xl-ref-2); background: var(--xl-ref-2-fill); }
.xl-ref-3 { border: 2px solid var(--xl-ref-3); background: var(--xl-ref-3-fill); }
.xl-ref-4 { border: 2px solid var(--xl-ref-4); background: var(--xl-ref-4-fill); }
.xl-ref-5 { border: 2px solid var(--xl-ref-5); background: var(--xl-ref-5-fill); }
.xl-ref-6 { border: 2px solid var(--xl-ref-6); background: var(--xl-ref-6-fill); }
.xl-editor {
  position: absolute; box-sizing: border-box; padding: 0 3px; z-index: 5;
  border: 2px solid var(--xl-selection-border); border-radius: 0;
  background: var(--xl-sheet-bg); color: var(--xl-cell-text); font-size: 13px;
}
.xl-editor-fake {
  position: absolute; box-sizing: border-box; padding: 0 4px; z-index: 5; line-height: 20px;
  border: 2px solid var(--xl-selection-border); overflow: hidden; white-space: nowrap;
  background: var(--xl-sheet-bg); color: var(--xl-cell-text);
}
.xl-corner {
  position: absolute; left: 0; top: 0; width: 40px; height: 24px; box-sizing: border-box;
  background: var(--xl-header-bg); border-right: 1px solid var(--xl-gridline); border-bottom: 1px solid var(--xl-gridline);
}
.xl-corner-tri {
  position: absolute; right: 4px; bottom: 4px; width: 0; height: 0;
  border-bottom: 9px solid var(--xl-gridline-frozen); border-left: 9px solid transparent;
}
.xl-colhdrs { position: absolute; left: 40px; top: 0; right: 0; height: 24px; overflow: hidden; background: var(--xl-header-bg); }
.xl-rowhdrs { position: absolute; left: 0; top: 24px; bottom: 0; width: 40px; overflow: hidden; background: var(--xl-header-bg); }
.xl-colh {
  position: absolute; top: 0; height: 24px; box-sizing: border-box; text-align: center; line-height: 23px;
  background: var(--xl-header-bg); color: var(--xl-header-fg); font-size: 12px;
  border-right: 1px solid var(--xl-gridline); border-bottom: 1px solid var(--xl-gridline);
}
.xl-rowh {
  position: absolute; left: 0; width: 40px; box-sizing: border-box; text-align: center; font-size: 12px;
  background: var(--xl-header-bg); color: var(--xl-header-fg);
  border-right: 1px solid var(--xl-gridline); border-bottom: 1px solid var(--xl-gridline);
}
.xl-hdr--on { background: var(--xl-header-active-bg); color: var(--xl-header-active-fg); font-weight: 700; }
.xl-vtrack { position: absolute; right: 0; width: 12px; background: var(--xl-header-bg); opacity: 0.85; }
.xl-htrack { position: absolute; bottom: 0; height: 12px; background: var(--xl-header-bg); opacity: 0.85; }
.xl-thumb { position: absolute; border-radius: 4px; background: var(--xl-gridline-frozen); opacity: 0.8; }
.xl-tip {
  position: absolute; z-index: 20; padding: 2px 8px; border-radius: 3px; font-size: 11px;
  background: var(--xl-titlebar-bg); color: var(--xl-titlebar-fg);
}
.xl-tabs {
  box-sizing: border-box; display: flex; align-items: center; gap: 2px; flex-shrink: 0; height: 28px; padding: 0 6px;
  background: var(--xl-tabbar-bg); border-top: 1px solid var(--xl-ribbon-border);
}
.xl-tab {
  display: flex; align-items: center; height: 24px; padding: 0 14px; border-radius: 0 0 4px 4px;
  font-size: 12px; color: var(--xl-header-fg); white-space: nowrap;
}
.xl-tab--active {
  background: var(--xl-tab-active-bg); color: var(--xl-brand); font-weight: 700;
  border-bottom: 2px solid var(--xl-brand);
}
.xl-tab-add {
  display: flex; align-items: center; justify-content: center; width: 24px; height: 24px;
  border: 0; border-radius: 12px; background: transparent; color: var(--xl-header-fg); font-size: 16px;
}
.xl-tab-add:hover { background: var(--xl-brand-tint); }
.xl-status {
  box-sizing: border-box; display: flex; align-items: center; gap: 16px; flex-shrink: 0; height: 24px; padding: 0 12px;
  background: var(--xl-statusbar-bg); color: var(--xl-statusbar-fg); font-size: 12px;
  border-top: 1px solid var(--xl-ribbon-border);
}
.xl-status-spacer { flex: 1; }
.xl-status-stat { white-space: nowrap; }
.xl-overlay-root { position: absolute; left: 0; top: 0; right: 0; bottom: 0; z-index: 50; }
`
