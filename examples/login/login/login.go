// Package login is the sign-in example.
// The screen is an HTML template. Clicks and keystrokes are Go functions.
// ownframe opens the window. This package does not.
// Copy, cut, and select-all on a focused control belong to the library.
package login

import (
	"context"
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed login.html
var loginHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 480
	DefaultHeight = 640

	// MinWidth and MinHeight are the smallest frame the screen will draw.
	MinWidth  = 320
	MinHeight = 400

	undoLimit = 64
)

// View is the login screen data.
// The template prints Error and Status, and grays the button until Ready.
// Email and Password bind through data-bind.
type View struct {
	Error    string
	Status   string
	Ready    bool
	Email    string
	Password ownframe.Secret
}

// snap is one saved pair of field values. The password stays a
// Secret so a %v or slog dump of undo history cannot leak it.
type snap struct {
	Email    string
	Password ownframe.Secret
}

// App is the sign-in screen.
type App struct {
	page *ownframe.Page
	view View
	undo []snap
	redo []snap
}

// New parses the embedded login template and registers its handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Sign in",
		HTML:      loginHTML,
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{
		Click:      app.onClick,
		Submit:     app.onSubmit,
		BeforeEdit: app.onBeforeEdit,
		Change:     app.onChange,
		Undo:       app.onUndo,
		Redo:       app.onRedo,
	})
	page.SetData(&app.view)

	return app, nil
}

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// View returns a copy of the current login fields.
func (a *App) View() View {
	return a.view
}

// Size returns the frame size in CSS pixels.
func (a *App) Size() (int, int) {
	return a.page.Size()
}

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(width, height int) {
	a.page.SetSize(width, height)
}

// Generation increases by one on every successful Redraw.
func (a *App) Generation() uint64 {
	return a.page.Probe().Generation()
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// PNG returns the last PNG, or nil when nothing has been drawn.
func (a *App) PNG() []byte {
	return a.page.Probe().PNG()
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and applies the login action.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Type appends text to the focused field.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Edit().Type(ctx, text)
}

// Backspace drops the last rune of the focused field.
func (a *App) Backspace(ctx context.Context) error {
	return a.page.Edit().Backspace(ctx)
}

// Submit checks the email and password.
func (a *App) Submit(ctx context.Context) error {
	return a.page.Edit().Submit(ctx)
}

func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if box.Action == "login" && a.ready() {
		a.signIn()
	}

	a.page.SetData(&a.view)

	return nil
}

// onChange refreshes the button state after a field changed.
func (a *App) onChange(context.Context, ownframe.Box) error {
	a.refresh()

	return nil
}

// onBeforeEdit snapshots the fields for undo and clears any old message.
func (a *App) onBeforeEdit(context.Context, ownframe.Box) error {
	a.push()

	if a.view.Status == "" && a.view.Error == "" {
		return nil
	}

	a.view.Status = ""
	a.view.Error = ""
	a.page.SetData(&a.view)

	return nil
}

func (a *App) onUndo(context.Context) error {
	if len(a.undo) == 0 {
		return nil
	}

	a.redo = append(a.redo, a.snapshot())
	a.restore(a.undo[len(a.undo)-1])
	a.undo = a.undo[:len(a.undo)-1]
	a.refresh()

	return nil
}

func (a *App) onRedo(context.Context) error {
	if len(a.redo) == 0 {
		return nil
	}

	a.undo = append(a.undo, a.snapshot())
	a.restore(a.redo[len(a.redo)-1])
	a.redo = a.redo[:len(a.redo)-1]
	a.refresh()

	return nil
}

func (a *App) onSubmit(_ context.Context) error {
	if !a.ready() {
		return nil
	}

	a.signIn()
	a.page.SetData(&a.view)

	return nil
}

func (a *App) signIn() {
	status := ""
	errMsg := "Unknown email or password."
	// Demo credentials only. DO NOT COPY into production: no auth
	// boundary here; a real app must verify server side, not locally.
	if a.view.Email == "secret" && a.view.Password.Reveal() == "secret" {
		status = "Signed in."
		errMsg = ""
	}

	if a.view.Status == status && a.view.Error == errMsg {
		return
	}

	a.view.Status = status
	a.view.Error = errMsg
}

// snapshot copies both bound fields.
func (a *App) snapshot() snap {
	return snap{Email: a.view.Email, Password: a.view.Password}
}

// push saves the current fields for undo. A snapshot equal to the last one
// is dropped so no-op edits do not stack. Any redo history is discarded.
func (a *App) push() {
	next := a.snapshot()
	if n := len(a.undo); n == 0 || a.undo[n-1] != next {
		a.undo = append(a.undo, next)
		if len(a.undo) > undoLimit {
			a.undo = a.undo[len(a.undo)-undoLimit:]
		}
	}

	a.redo = nil
}

// restore writes a snapshot into View. The next Redraw pushes it.
func (a *App) restore(s snap) {
	a.view.Email = s.Email
	a.view.Password = s.Password
	a.page.Form().SetFormValue("email", s.Email)
	a.page.Form().SetFormValue("password", s.Password.Reveal())
}

// ready reports whether both fields carry text, so the sign-in button can
// turn blue and take a click only after the user filled them.
func (a *App) ready() bool {
	return a.view.Email != "" && a.view.Password.Reveal() != ""
}

// refresh syncs View.Ready with the fields. The redraw that follows every
// handler reads the new value from the template data.
func (a *App) refresh() {
	ready := a.ready()
	if a.view.Ready == ready {
		return
	}

	a.view.Ready = ready
	a.page.SetData(&a.view)
}

// Logout clears the fields, the undo and redo stacks, and the clipboard.
// Demo only; secret/secret stays the test login. DO NOT COPY.
func (a *App) Logout() {
	a.page.Form().SetFormValue("email", "")
	a.page.Form().SetFormValue("password", "")
	a.undo = nil
	a.redo = nil
	a.view = View{}
	a.page.SetData(&a.view)
	ownframe.Clear()
}
