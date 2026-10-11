package telegram

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// App is the Telegram demo screen.
type App struct {
	page     *ownframe.Page
	view     View
	chats    []Chat
	threads  map[string][]Message
	contacts []Contact

	// Cross-thread state: the activity asks for a back press or a photo
	// and reports insets, and the next tick applies them.
	mu         sync.Mutex
	attachKind string
	attachID   string
	back       bool
	blur       bool
	photos     chan []byte
	photoN     int
	onList     atomic.Bool
	dark       atomic.Bool
	insetTop   atomic.Int64
	insetBot   atomic.Int64
	dirty      bool
	pullThread bool
	// lastScrollY, lastMove, and scrollDirty gather a phone scroll into one
	// settling redraw after the movement stops.
	lastScrollY int
	viewportH   int
	lastMove    time.Time
	scrollDirty bool
}

// New parses the embedded template and registers its images and handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:     "Telegram",
		HTML:      buildHTML(),
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app := seed()
	app.page = page
	app.photos = make(chan []byte, 2)
	app.view.Tab = "chats"
	app.rebuild()
	registerImages(page)
	page.Handle(ownframe.Handlers{
		Click:     app.onClick,
		Change:    app.onChange,
		Submit:    app.onSubmit,
		LongPress: app.onLongPress,
	})
	page.SetWindowing(true)
	page.SetViewportPinZ(2)
	page.SetScrollWindow(app.Pin)
	page.SetTick(app.Tick)
	page.SetData(&app.view)

	return app, nil
}
