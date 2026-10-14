package ui

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// New builds the page, restores the stored theme, and starts polling.
func New(ctx context.Context, cfg Config) (*App, error) {
	if cfg.Source == nil {
		return nil, ErrNoSource
	}

	cfg = withDefaults(cfg)

	settings := loadSettings(ctx, cfg.Store)

	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "System Monitor",
		HTML:      buildHTML(),
		Theme:     themeSource(settings.Dark),
		Width:     cfg.Width,
		Height:    cfg.Height,
		MinWidth:  cfg.MinWidth,
		MinHeight: cfg.MinHeight,
	})
	if err != nil {
		return nil, err
	}

	a := &App{page: page, cfg: cfg, src: cfg.Source}
	a.state = newState(settings.Dark, cfg.Live)
	a.sync()

	page.Handle(ownframe.Handlers{
		Click:   a.onClick,
		Change:  a.onChange,
		KeyDown: a.onKeyDown,
	})
	page.SetTick(a.Tick)

	if !cfg.noPump {
		a.startPumps(ctx)
	}

	return a, nil
}
