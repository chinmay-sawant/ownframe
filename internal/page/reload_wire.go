package page

import (
	"context"
	"html/template"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/reload"
)

// watchState is the file watch behind PollReload.
type watchState = reload.State

var _ host.Reloader = (*Page)(nil)

// watchFor builds the watch from the config and the bytes New read.
func watchFor(cfg Config, html []byte, htmlInfo os.FileInfo, theme []byte, themeInfo os.FileInfo) *watchState {
	return reload.Open(reload.Spec{
		File:      cfg.File,
		ThemeFile: cfg.ThemeFile,
		Hot:       !cfg.DisableHotReload,
		HTML:      html,
		HTMLInfo:  htmlInfo,
		CSS:       theme,
		CSSInfo:   themeInfo,
	})
}

// SetHotReload turns the file watch on or off. It does nothing without a
// watched file, and nothing on wasm and mobile, where a file is read once.
func (p *Page) SetHotReload(on bool) {
	if p.watch == nil {
		return
	}
	p.watch.SetHot(on)
}

// Watching reports whether a poll can reread a source file.
func (p *Page) Watching() bool {
	return p.watch.On()
}

// PollReload stats and rereads the watched files. It reports whether the
// page changed, and returns the first error seen. A read that does not parse
// keeps the last good template or theme and is retried on the next poll.
func (p *Page) PollReload(ctx context.Context) (bool, error) {
	if err := useContext(ctx); err != nil {
		return false, err
	}
	if !p.watch.On() {
		return false, nil
	}
	applied, first := p.watch.Poll(reload.Hooks{
		Files:    srcFiles,
		Limit:    maxSourceSize,
		Oversize: sourceTooLarge,
		Template: p.installWatched,
		Theme:    p.SetTheme,
	})
	if !applied {
		p.noteReloadError(first)
		return false, first
	}
	if err := p.Redraw(ctx); err != nil {
		p.noteReloadError(err)
		return false, err
	}
	p.noteReload(first)
	p.resolveStates()
	return true, first
}

func (p *Page) installWatched(tpl *template.Template, src string) error {
	p.tpl = tpl
	p.nav.Retarget(src)
	p.formDirty = true
	p.invalidateAccessibilityDocument()
	return nil
}

// noteReload records one successful reload for host.Stats. err is the first
// error seen in the same poll, nil when the poll was clean. The text is
// sanitized to one line so Stats never carries log-forging lines.
func (p *Page) noteReload(err error) {
	p.stats.reloads++
	if err != nil {
		p.stats.reloadErr = reload.Clean(err.Error())
		return
	}
	p.stats.reloadErr = ""
}

// noteReloadError records a failed poll that applied nothing.
func (p *Page) noteReloadError(err error) {
	if err != nil {
		p.stats.reloadErr = reload.Clean(err.Error())
	}
}

// resolveStates clears hover and active ids the new layout does not have.
// syncForm does the same for focus when Redraw runs.
func (p *Page) resolveStates() {
	ids := map[string]bool{}
	for _, box := range p.boxes {
		if box.ID != "" {
			ids[box.ID] = true
		}
	}
	if p.HoverID() != "" && !ids[p.HoverID()] {
		p.clearHoverID()
	}
	if p.activeID() != "" && !ids[p.activeID()] {
		p.setActiveID("")
	}
}
