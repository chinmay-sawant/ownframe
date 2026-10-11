package app

import (
	"context"
	"embed"
	"fmt"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/store"
	"strings"
)

//go:embed index.html style.css
var files embed.FS

type App struct {
	page    *ownframe.Page
	db      *store.Store
	view    View
	pickers chan string
	exports chan []byte
	imports chan []byte
}

func New(path string) (*App, error) {
	db, err := store.Open(path)
	storageWarning := ""
	if err != nil {
		db, err = store.Open(":memory:")
		if err != nil {
			return nil, fmt.Errorf("pet-care temporary storage: %w", err)
		}
		storageWarning = "Local storage could not open. Changes stay temporary for this session."
	}
	if err := db.SeedDemo(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("pet-care demo data: %w", err)
	}
	html, err := files.ReadFile("index.html")
	if err != nil {
		return nil, err
	}
	css, err := files.ReadFile("style.css")
	if err != nil {
		return nil, err
	}
	source := strings.Replace(string(html), "<!-- styles -->", "<style>"+string(css)+"</style>", 1)
	p, err := ownframe.New(ownframe.Config{Title: "Pet care", HTML: source, Width: 1120, Height: 820, MinWidth: 360, MinHeight: 600, DisableTouchZoom: true})
	if err != nil {
		db.Close()
		return nil, err
	}
	prepareImages(p)
	a := &App{page: p, db: db, pickers: make(chan string, 1), exports: make(chan []byte, 1), imports: make(chan []byte, 1), view: View{Screen: "login", Email: "demo@petcare.local", Password: "pawprints", StorageWarning: storageWarning}}
	p.Handle(ownframe.Handlers{Click: a.click, Submit: a.submit, Change: a.change})
	p.SetTick(a.Tick)
	if err := a.reload(context.Background()); err != nil {
		return nil, err
	}
	p.SetData(a.view)
	return a, nil
}

func (a *App) Page() *ownframe.Page { return a.page }
func (a *App) Close() error {
	if a.db == nil {
		return nil
	}
	return a.db.Close()
}
func (a *App) Redraw(ctx context.Context) error { return a.page.Redraw(ctx) }
