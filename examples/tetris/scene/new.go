package scene

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// New builds the game page and registers its handlers and frame callback.
func New(model Model, st Store, opts Options) (*Scene, error) {
	return newFromHTML(model, st, opts, pageHTML(gameHTML))
}

// newFromHTML is New with the template source handed in, for tests.
func newFromHTML(model Model, st Store, opts Options, html string) (*Scene, error) {
	now := opts.Now
	if now == nil {
		now = time.Now
	}

	step := opts.Stepper
	if step == nil {
		step = &game.Clock{}
	}

	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Ownframe Tetris",
		HTML:      html,
		Theme:     themeCSS(opts.Dark),
		Perf:      opts.Perf,
		Width:     Width,
		Height:    Height,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	status := opts.Status
	if status == "" {
		status = "READY"
	}

	s := &Scene{
		page:    page,
		model:   model,
		store:   st,
		now:     now,
		step:    step,
		focus:   opts.Focused,
		focused: true,
		dark:    opts.Dark,
		frame:   model.Frame(),
		runID:   model.RunID(),
		status:  status,
	}

	page.Handle(ownframe.Handlers{
		KeyDown: s.onKeyDown,
		KeyUp:   s.onKeyUp,
		Click:   s.onClick,
	})
	page.SetTick(s.Tick)
	page.SetData(s.view())

	if st != nil {
		s.settingsReq = st.Settings()
	}

	return s, nil
}
