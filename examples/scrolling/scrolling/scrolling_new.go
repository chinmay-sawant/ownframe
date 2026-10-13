package scrolling

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded scrolling template.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Scrolling",
		HTML:   scrollingHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	return &App{page: p}, nil
}
