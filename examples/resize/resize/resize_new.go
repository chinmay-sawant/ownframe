package resize

import "github.com/chinmay-sawant/ownframe"

// New parses the embedded resize template.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Resize",
		HTML:      resizeHTML,
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	return &App{page: page}, nil
}
