package cat

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/assets"
)

// Companion owns the UI. Inbox is safe for concurrent agent requests.
type Companion struct {
	Page      *ownframe.Page
	inbox     *Inbox
	animation *animation
	shown     bool
	version   uint64
	view      bubbleView
	// ReducedMotion disables bobbing and automatic expression changes.
	ReducedMotion bool
	// RandomBehavior randomizes idle expressions.
	RandomBehavior bool
	seconds        float64
}

type bubbleView struct {
	Shown                 bool
	Message, Source, Time string
}

// NewWithInbox connects an agent inbox to a cycling or fixed cat.
func NewWithInbox(variant int, inbox *Inbox) (*Companion, error) {
	files, err := fs.Glob(assets.Cats, "cat_images/*.png")
	if err != nil {
		return nil, err
	}
	if variant < 0 || variant > len(files) {
		return nil, fmt.Errorf("invalid variant %d", variant)
	}
	return newCompanion(variant, files, inbox)
}

func newCompanion(variant int, files []string, inbox *Inbox) (*Companion, error) {
	page, err := ownframe.New(ownframe.Config{Title: "Desktop cat", HTML: source, Width: windowWidth, Height: windowHeight})
	if err != nil {
		return nil, err
	}
	if inbox == nil {
		inbox = NewInbox()
	}
	c := &Companion{Page: page, inbox: inbox, shown: true}
	c.animation = &animation{page: page, files: files, current: max(0, variant-1), cycle: variant == 0}
	if err := c.animation.load(c.animation.current); err != nil {
		return nil, err
	}
	c.receive()
	page.Handle(ownframe.Handlers{Click: c.click})
	start := time.Now()
	page.SetTick(func(ctx context.Context) error { return c.tick(ctx, time.Since(start).Seconds()) })
	return c, nil
}
