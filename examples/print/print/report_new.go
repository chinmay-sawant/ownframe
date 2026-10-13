package print

import (
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded report and registers the click handler. webMode
// adds the /pdf hint to the page.
func New(webMode bool) (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Quarterly report",
		HTML:   reportHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page, SavePath: defaultSavePath()}
	app.view = View{Title: "Quarterly report", Web: webMode, Status: "Ready."}

	page.Handle(ownframe.Handlers{Click: app.onClick})
	page.SetData(&app.view)

	return app, nil
}

// defaultSavePath is ownframe-print-report.pdf under the user home directory.
func defaultSavePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "ownframe-print-report.pdf"
	}

	return filepath.Join(home, "ownframe-print-report.pdf")
}
