package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strconv"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	windowWidth  = 1180
	windowHeight = 760
)

const pageHTML = `<!DOCTYPE html>
<html>
<head>
<style>
html, body {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  position: relative;
  background: #0f172a;
  color: #e2e8f0;
  font-family: sans-serif;
}
.shell {
  padding: 18px 20px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 10px;
  margin-bottom: 14px;
}
.title {
  font-size: 18px;
  font-weight: bold;
  color: #f8fafc;
}
.subtitle {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 3px;
}
.status-pill {
  box-sizing: border-box;
  width: 220px;
  height: 30px;
  flex-shrink: 0;
  overflow: hidden;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 12px;
  border-radius: 999px;
  background: #0f172a;
  border: 1px solid #38bdf8;
  color: #7dd3fc;
  font-size: 11px;
  font-weight: bold;
}
.grid {
  display: flex;
  gap: 14px;
  margin-bottom: 14px;
}
.panel {
  flex: 1;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 10px;
  padding: 14px 16px;
}
.panel-title {
  font-size: 14px;
  font-weight: bold;
  color: #f8fafc;
  margin-bottom: 4px;
}
.panel-desc {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 12px;
}
.btn-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.btn {
  padding: 6px 10px;
  border-radius: 6px;
  background: #334155;
  color: #f1f5f9;
  font-size: 12px;
  font-weight: bold;
  border: 1px solid #475569;
  cursor: pointer;
}
.btn-accent {
  background: #0284c7;
  border-color: #38bdf8;
  color: #ffffff;
}
.btn-warn {
  background: #9a3412;
  border-color: #fb923c;
  color: #ffedd5;
}
.target-card {
  box-sizing: border-box;
  height: 66px;
  overflow: hidden;
  padding: 12px 14px;
  border-radius: 8px;
  background: #0f172a;
  border: 2px solid #334155;
  margin-bottom: 10px;
}
.target-card.active {
  background: #172554;
  border-color: #38bdf8;
}
.target-card.flagged {
  background: #451a03;
  border-color: #f59e0b;
}
.target-card[data-priority="high"] {
  border-color: #f43f5e;
  background: #3b0764;
}
.card-head {
  height: 18px;
  overflow: hidden;
  white-space: nowrap;
  font-size: 13px;
  font-weight: bold;
  color: #f8fafc;
  margin-bottom: 4px;
}
.card-meta {
  height: 18px;
  overflow: hidden;
  white-space: nowrap;
  font-size: 12px;
  color: #cbd5e1;
}
.batch-row {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}
.batch-chip {
  box-sizing: border-box;
  flex: 1;
  height: 34px;
  overflow: hidden;
  white-space: nowrap;
  padding: 8px 10px;
  border-radius: 6px;
  background: #0f172a;
  border: 1px solid #334155;
  font-size: 12px;
  color: #cbd5e1;
}
.batch-chip.lit {
  background: #064e3b;
  border-color: #34d399;
  color: #d1fae5;
}
.feed-box {
  box-sizing: border-box;
  height: 156px;
  overflow: hidden;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 6px 8px;
}
.feed-item {
  padding: 6px 8px;
  border-radius: 5px;
  background: #1e293b;
  border: 1px solid #334155;
  font-size: 12px;
  color: #e2e8f0;
  margin-bottom: 5px;
}
.motion-grid {
  display: flex;
  gap: 12px;
}
.motion-card {
  flex: 1;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 12px;
}
.motion-label {
  font-size: 12px;
  font-weight: bold;
  color: #f8fafc;
  margin-bottom: 4px;
}
.motion-sub {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 10px;
}
.stage {
  height: 76px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1e293b;
  border-radius: 6px;
}
.stage-clip {
  height: 76px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1e293b;
  border-radius: 6px;
}
.radius-box {
  width: 54px;
  height: 54px;
  background: #38bdf8;
  border-radius: 6px;
  animation: morphRadius 1.8s ease-in-out infinite alternate;
}
.alpha-box {
  width: 72px;
  height: 44px;
  border-radius: 8px;
  background: rgba(52, 211, 153, 0.25);
  border: 1px solid #34d399;
  animation: pulseAlpha 1.6s ease-in-out infinite alternate;
}
.indiv-box {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: #a78bfa;
  animation: indivOrbit 2.2s ease-in-out infinite alternate;
}
.planar-box {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: #fb923c;
  animation: planarGlide 2s ease-in-out infinite alternate;
}
.clip-pill {
  width: 110px;
  height: 30px;
  border-radius: 15px;
  background: #22d3ee;
  color: #083344;
  font-size: 11px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: clipSweep 2.4s ease-in-out infinite alternate;
}
.blend-wrap {
  display: flex;
  gap: 10px;
  align-items: center;
}
.blend-backdrop {
  width: 72px;
  height: 46px;
  border-radius: 6px;
  background: #f43f5e;
  display: flex;
  align-items: center;
  justify-content: center;
}
.blend-chip {
  width: 42px;
  height: 28px;
  border-radius: 4px;
  background: #38bdf8;
  mix-blend-mode: multiply;
}
.blend-sibling {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: #34d399;
  animation: siblingBob 1.4s ease-in-out infinite alternate;
}
.toast-slot {
  position: absolute;
  right: 20px;
  bottom: 20px;
  box-sizing: border-box;
  width: 310px;
  height: 82px;
  overflow: hidden;
}
.toast-card {
  box-sizing: border-box;
  width: 310px;
  height: 82px;
  overflow: hidden;
  padding: 10px 14px;
  border-radius: 10px;
  background: #1e293b;
  border: 1px solid #38bdf8;
}
.toast-card.warn {
  background: #451a03;
  border-color: #f59e0b;
}
.toast-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 20px;
  margin-bottom: 4px;
}
.toast-title {
  height: 18px;
  overflow: hidden;
  white-space: nowrap;
  font-size: 12px;
  font-weight: bold;
  color: #f8fafc;
}
.toast-close {
  box-sizing: border-box;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  line-height: 20px;
  text-align: center;
  border-radius: 4px;
  background: #334155;
  color: #e2e8f0;
  font-size: 12px;
  font-weight: bold;
  cursor: pointer;
  flex-shrink: 0;
}
.toast-body {
  height: 36px;
  overflow: hidden;
  font-size: 11px;
  line-height: 17px;
  color: #cbd5e1;
}
@keyframes morphRadius {
  from { border-radius: 4px; }
  to   { border-radius: 27px; }
}
@keyframes pulseAlpha {
  from { background-color: rgba(52, 211, 153, 0.20); }
  to   { background-color: rgba(52, 211, 153, 0.95); }
}
@keyframes indivOrbit {
  from { translate: -18px 0px; rotate: -18deg; scale: 0.85; }
  to   { translate: 18px 0px; rotate: 18deg; scale: 1.12; }
}
@keyframes planarGlide {
  from { transform: translate3d(-20px, 0px, 0px) scale3d(0.9, 0.9, 1); }
  to   { transform: translate3d(20px, 0px, 0px) scale3d(1.1, 1.1, 1); }
}
@keyframes clipSweep {
  from { transform: translate(-48px, 0px); }
  to   { transform: translate(48px, 0px); }
}
@keyframes siblingBob {
  from { transform: translate(0px, -10px); }
  to   { transform: translate(0px, 10px); }
}
</style>
</head>
<body>
<div class="shell">
  <div class="header">
    <div>
      <div class="title">OwnFrame v0.0.9 — Retained DOM &amp; Motion Showcase</div>
      <div class="subtitle">DevTools + Perf ON by default. Click controls to patch the retained DOM in place and inspect Frame timings.</div>
    </div>
    <div id="status-pill" class="status-pill">Ready — 0 patches applied</div>
  </div>

  <div class="grid">
    <div class="panel">
      <div class="panel-title">1. In-Place Retained DOM Mutations (PatchElement / PatchElements)</div>
      <div class="panel-desc">No template rebuild — mutates the retained tree directly with Blinkless v0.1.11 and direct op diffing.</div>
      <div class="btn-row">
        <span class="btn btn-accent" data-action="toggle-class">ToggleClass</span>
        <span class="btn" data-action="flag-on">AddClass</span>
        <span class="btn" data-action="flag-off">RemoveClass</span>
        <span class="btn" data-action="attr-set">SetAttr</span>
        <span class="btn" data-action="attr-remove">RemoveAttr</span>
        <span class="btn" data-action="patch-text">PatchText</span>
        <span class="btn btn-accent" data-action="batch-patch">Batch 3x</span>
      </div>
      <div id="target-card" class="target-card">
        <div id="target-title" class="card-head">Target Card (#target-card)</div>
        <div id="target-meta" class="card-meta">State: idle | text updates: 0</div>
      </div>
      <div class="batch-row">
        <div id="chip-1" class="batch-chip">Slot A: 0</div>
        <div id="chip-2" class="batch-chip">Slot B: 0</div>
        <div id="chip-3" class="batch-chip">Slot C: 0</div>
      </div>
    </div>

    <div class="panel">
      <div class="panel-title">2. Structural Mutations (AppendHTML / InnerHTML / Remove)</div>
      <div class="panel-desc">Insert, replace, or detach child HTML fragments inside a retained container without rebuilding the page.</div>
      <div class="btn-row">
        <span class="btn btn-accent" data-action="append-item">AppendHTML</span>
        <span class="btn btn-warn" data-action="remove-last">Remove Last</span>
        <span class="btn" data-action="reset-feed">InnerHTML Reset</span>
      </div>
      <div id="feed-list" class="feed-box">
        <div id="feed-item-1" class="feed-item">#1 Initial retained feed entry</div>
        <div id="feed-item-2" class="feed-item">#2 Click AppendHTML, Remove Last, or InnerHTML Reset</div>
      </div>
    </div>
  </div>

  <div class="panel">
    <div class="panel-title">3. CSS Motion Fast Paths &amp; Tile-Surface Retained Layers</div>
    <div class="panel-desc">Watch the DevTools Frame tab: border-radius, color alpha, individual transforms, planar 3D, clipped layers, and blend siblings stay promoted with zero relayouts.</div>
    <div class="motion-grid">
      <div class="motion-card">
        <div class="motion-label">border-radius</div>
        <div class="motion-sub">In-place MaskRadius (0 relayout)</div>
        <div class="stage"><div class="radius-box"></div></div>
      </div>
      <div class="motion-card">
        <div class="motion-label">color alpha</div>
        <div class="motion-sub">In-place MaskAlpha</div>
        <div class="stage"><div class="alpha-box"></div></div>
      </div>
      <div class="motion-card">
        <div class="motion-label">translate/rotate/scale</div>
        <div class="motion-sub">Individual 2D compositor</div>
        <div class="stage"><div class="indiv-box"></div></div>
      </div>
      <div class="motion-card">
        <div class="motion-label">translate3d / scale3d</div>
        <div class="motion-sub">Planar 2D affine fast path</div>
        <div class="stage"><div class="planar-box"></div></div>
      </div>
      <div class="motion-card">
        <div class="motion-label">overflow: hidden</div>
        <div class="motion-sub">Clipped retained layer</div>
        <div class="stage-clip"><div class="clip-pill">CLIPPED</div></div>
      </div>
      <div class="motion-card">
        <div class="motion-label">blend sibling</div>
        <div class="motion-sub">Per-layer blend check</div>
        <div class="stage">
          <div class="blend-wrap">
            <div class="blend-backdrop"><div class="blend-chip"></div></div>
            <div class="blend-sibling"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
<div id="toast-slot" class="toast-slot"></div>
</body>
</html>`

const (
	toastDuration      = 5 * time.Second
	initialNextFeedSeq = 3
)

// App holds the interactive v0.0.9 showcase page and mutation counters.
type App struct {
	page         *ownframe.Page
	patchCount   int
	textUpdates  int
	batchCount   int
	feedIDs      []string
	nextFeedSeq  int
	toastVisible bool
	toastUntil   time.Time
	now          func() time.Time
}

// New builds the v0.0.9 retained DOM & motion showcase page.
func New(devtools, perf bool) (*App, error) {
	p, err := ownframe.New(ownframe.Config{
		Title:       "OwnFrame v0.0.9 — Retained DOM & Motion Showcase",
		HTML:        pageHTML,
		Width:       windowWidth,
		Height:      windowHeight,
		RetainState: true,
		DevTools:    devtools,
		Perf:        perf,
	})
	if err != nil {
		return nil, err
	}

	a := &App{
		page:        p,
		feedIDs:     []string{"feed-item-1", "feed-item-2"},
		nextFeedSeq: initialNextFeedSeq,
		now:         time.Now,
	}

	p.Handle(ownframe.Handlers{
		ClickPaint: a.handleClick,
	})
	p.SetTick(a.handleTick)

	return a, nil
}

// Page returns the underlying ownframe page.
func (a *App) Page() *ownframe.Page {
	return a.page
}

func (a *App) handleClick(ctx context.Context, box ownframe.Box) (bool, error) {
	if box.Action == "" {
		return false, nil
	}
	if err := a.handleAction(ctx, box.Action); err != nil {
		return false, err
	}

	return true, nil
}

func (a *App) handleAction(ctx context.Context, name string) error {
	start := time.Now()
	var (
		label string
		err   error
	)

	switch name {
	case "toggle-class":
		label = "ToggleClass(.active)"
		err = a.page.PatchElement(ctx, "target-card", ownframe.ElementEdit{
			ToggleClass: "active",
		})
	case "flag-on":
		label = "AddClass(.flagged)"
		err = a.page.PatchElement(ctx, "target-card", ownframe.ElementEdit{
			AddClass: "flagged",
		})
	case "flag-off":
		label = "RemoveClass(.flagged)"
		err = a.page.PatchElement(ctx, "target-card", ownframe.ElementEdit{
			RemoveClass: "flagged",
		})
	case "attr-set":
		label = `SetAttr(data-priority="high")`
		err = a.page.PatchElement(ctx, "target-card", ownframe.ElementEdit{
			Attribute: "data-priority",
			Value:     "high",
		})
	case "attr-remove":
		label = "RemoveAttr(data-priority)"
		err = a.page.PatchElement(ctx, "target-card", ownframe.ElementEdit{
			RemoveAttribute: "data-priority",
		})
	case "patch-text":
		a.textUpdates++
		label = "PatchText(#target-meta)"
		msg := fmt.Sprintf("State: patched in place | text updates: %d", a.textUpdates)
		err = a.page.PatchElement(ctx, "target-meta", ownframe.ElementEdit{
			Text: &msg,
		})
	case "batch-patch":
		a.batchCount++
		label = "PatchElements(3 chips)"
		n := strconv.Itoa(a.batchCount)
		t1 := "Slot A: " + n
		t2 := "Slot B: " + n
		t3 := "Slot C: " + n
		err = a.page.PatchElements(ctx, []ownframe.ElementPatch{
			{ID: "chip-1", Edits: []ownframe.ElementEdit{{Text: &t1}, {ToggleClass: "lit"}}},
			{ID: "chip-2", Edits: []ownframe.ElementEdit{{Text: &t2}, {ToggleClass: "lit"}}},
			{ID: "chip-3", Edits: []ownframe.ElementEdit{{Text: &t3}, {ToggleClass: "lit"}}},
		})
	case "append-item":
		seq := a.nextFeedSeq
		a.nextFeedSeq++
		id := fmt.Sprintf("feed-item-%d", seq)
		frag := fmt.Sprintf(`<div id="%s" class="feed-item">#%d Appended via ElementEdit.AppendHTML</div>`, id, seq)
		label = "AppendHTML(#" + id + ")"
		err = a.page.PatchElement(ctx, "feed-list", ownframe.ElementEdit{
			AppendHTML: frag,
		})
		if err == nil {
			a.feedIDs = append(a.feedIDs, id)
		}
	case "remove-last":
		if len(a.feedIDs) == 0 {
			return a.showToast(ctx, "No-op (Feed Empty)", "No feed items left to remove.", true, time.Since(start))
		}
		lastID := a.feedIDs[len(a.feedIDs)-1]
		label = "Remove(#" + lastID + ")"
		err = a.page.PatchElement(ctx, "feed-list", ownframe.ElementEdit{
			ID:     lastID,
			Remove: true,
		})
		if err == nil {
			a.feedIDs = a.feedIDs[:len(a.feedIDs)-1]
		}
	case "reset-feed":
		label = "InnerHTML(#feed-list)"
		frag := `<div id="feed-item-1" class="feed-item">#1 Reset via ElementEdit.InnerHTML</div>` +
			`<div id="feed-item-2" class="feed-item">#2 Ready for AppendHTML or Remove</div>`
		err = a.page.PatchElement(ctx, "feed-list", ownframe.ElementEdit{
			InnerHTML: &frag,
		})
		if err == nil {
			a.feedIDs = []string{"feed-item-1", "feed-item-2"}
			a.nextFeedSeq = initialNextFeedSeq
		}
	case "dismiss-toast":
		return a.hideToast(ctx)
	default:
		return nil
	}

	elapsed := time.Since(start)
	if err != nil {
		if errors.Is(err, ownframe.ErrPatchRefused) {
			return a.showToast(ctx, "Patch Refused (No-op)", fmt.Sprintf("%s had no effect (ErrPatchRefused) in %s", label, formatElapsed(elapsed)), true, elapsed)
		}

		return err
	}

	a.patchCount++
	if err := a.updateStatus(ctx, elapsed); err != nil {
		return err
	}

	return a.showToast(ctx, "Retained Patch Applied", fmt.Sprintf("%s completed in %s without full rebuild", label, formatElapsed(elapsed)), false, elapsed)
}

func (a *App) handleTick(ctx context.Context) error {
	if !a.toastVisible || a.toastUntil.IsZero() || a.now().Before(a.toastUntil) {
		return nil
	}

	return a.hideToast(ctx)
}

func (a *App) hideToast(ctx context.Context) error {
	if !a.toastVisible {
		return nil
	}

	a.toastVisible = false
	a.toastUntil = time.Time{}

	empty := ""
	err := a.page.PatchElement(ctx, "toast-slot", ownframe.ElementEdit{
		InnerHTML: &empty,
	})
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return nil
	}

	return err
}

func (a *App) updateStatus(ctx context.Context, d time.Duration) error {
	status := fmt.Sprintf("%d patches | last: %s", a.patchCount, formatElapsed(d))

	err := a.page.PatchElement(ctx, "status-pill", ownframe.ElementEdit{
		Text: &status,
	})
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return nil
	}

	return err
}

func (a *App) showToast(ctx context.Context, title, body string, warn bool, _ time.Duration) error {
	a.toastVisible = true
	a.toastUntil = a.now().Add(toastDuration)
	a.page.WantFrameAt(a.toastUntil)

	cls := "toast-card"
	if warn {
		cls = "toast-card warn"
	}
	frag := fmt.Sprintf(
		`<div id="toast-card" class="%s"><div class="toast-head"><div id="toast-title" class="toast-title">%s</div><button id="toast-close" type="button" class="toast-close" data-action="dismiss-toast">X</button></div><div id="toast-body" class="toast-body">%s</div></div>`,
		cls, html.EscapeString(title), html.EscapeString(body),
	)

	err := a.page.PatchElement(ctx, "toast-slot", ownframe.ElementEdit{
		InnerHTML: &frag,
	})
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return nil
	}

	return err
}

func formatElapsed(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	}

	return fmt.Sprintf("%d µs", d.Microseconds())
}
