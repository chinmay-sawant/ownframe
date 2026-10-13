package remote

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/lg-remote/bridge"
	"github.com/chinmay-sawant/ownframe/examples/lg-remote/tv"
)

//go:embed remote.html
var remoteHTML string

// Option changes a new remote.
type Option func(*App)

// WithBluetooth shows the Wi-Fi / Bluetooth switch. Android turns it on.
func WithBluetooth(on bool) Option {
	return func(a *App) { a.view.ShowBluetooth = on }
}

// New parses the remote and registers click, key, and frame handlers.
func New(opts ...Option) (*App, error) {
	app := &App{
		dark:       true,
		async:      true,
		wantSearch: true,
		notes:      make(chan update, 8),
		view: View{
			FontSize:    16,
			Sensitivity: 100,
			Title:       "LG remote",
			Status:      "Searching this Wi-Fi for the TV.",
			Mode:        "Wi-Fi",
			ThemeLabel:  "Light",
			Hint:        "The phone and the TV use the same Wi-Fi.",
			PowerOn:     true,
			Face:        faceKeys(),
			More:        moreKeys(),
		},
	}

	for _, opt := range opts {
		opt(app)
	}

	html := remoteHTML
	if app.phone {
		html = phoneHTML + `{{define "phone-css"}}` + phoneCSS + `{{end}}`
	}

	minWidth, minHeight := 320, 480
	if app.phone {
		minWidth, minHeight = 240, 160
	}
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "LG remote",
		HTML:      html,
		Theme:     darkTheme,
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  minWidth,
		MinHeight: minHeight,
		LockView:  true,
	})
	if err != nil {
		return nil, err
	}

	app.page = page
	app.installPage()

	return app, nil
}

func (a *App) use() linker {
	if a.link == nil {
		a.link = tv.NewSession(bridge.Dir())
	}

	return a.link
}

// SetLink replaces the TV session. Tests use it.
func (a *App) SetLink(link linker) { a.link = link }

// SetAsync runs network work on the caller when false.
func (a *App) SetAsync(on bool) { a.async = on }
